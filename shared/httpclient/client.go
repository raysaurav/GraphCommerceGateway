package httpclient

import (
	"context"
	"errors"
	"net"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/sethvargo/go-envconfig"
	"github.com/sony/gobreaker"
)

// Config holds the configuration for the circuit breaker and Resty client
type Config struct {
	*HttpConfig `env:",prefix=COPPEL_HTTP_"`
}

type HttpConfig struct {
	RetryCount           int           `env:"RETRY_COUNT,default=2"`
	RetryWaitTime        time.Duration `env:"RETRY_WAIT_TIME,default=100ms"`
	RetryMaxWait         time.Duration `env:"RETRY_MAX_WAIT,default=1s"`
	ProxyURL             string        `env:"PROXY_URL,required"`
	NoProxy              string        `env:"NO_PROXY"`
	ConsecutiveFailures  uint32        `env:"CB_MAX_ALLOWED_FAILED_REQS,default=10"`
	MetricsResetInterval time.Duration `env:"CB_RESET_INTERVAL,default=10s"`
	OpenStateDuration    time.Duration `env:"CB_OPEN_STATE_DURATION,default=500ms"`
	MaxRequest           uint32        `env:"CB_SUCCESSFUL_REQ_TO_CLOSE,default=2"`
}

var (
	configInstance     *Config
	configOnce         sync.Once
	httpClientInstance *HttpClient
	httpClientOnce     sync.Once
)

func NewConfig() (*Config, error) {
	var err error
	configOnce.Do(func() {
		var cfg Config
		err = envconfig.Process(context.Background(), &cfg)
		if err == nil {
			configInstance = &cfg
		}
	})
	return configInstance, err
}

type HttpClient struct {
	RestyClient    *resty.Client
	CircuitBreaker *gobreaker.CircuitBreaker
}

type GoBreakerTransport struct {
	CB        *gobreaker.CircuitBreaker
	Transport http.RoundTripper
}

// RoundTrip executes the HTTP request using the underlying transport,
// wrapped with a circuit breaker to prevent calls when the breaker is open.
// It returns the HTTP response or an error if the breaker is tripped or the request fails.

func (t *GoBreakerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := t.CB.Execute(func() (interface{}, error) {
		return t.Transport.RoundTrip(req)
	})
	if err != nil {
		return nil, err
	}
	return resp.(*http.Response), nil
}

func NewHttpClient(cfg *Config) HttpClientInterface {
	httpClientOnce.Do(func() {
		cbSettings := gobreaker.Settings{
			Name:        "HttpClientCircuitBreaker",
			MaxRequests: cfg.MaxRequest,
			Interval:    cfg.MetricsResetInterval,
			Timeout:     cfg.OpenStateDuration,
			ReadyToTrip: func(counts gobreaker.Counts) bool {
				return counts.ConsecutiveFailures > cfg.ConsecutiveFailures
			},
		}

		circuitBreaker := gobreaker.NewCircuitBreaker(cbSettings)

		// Default transport
		transport := &http.Transport{
			MaxIdleConns:          200,
			MaxIdleConnsPerHost:   100,
			IdleConnTimeout:       90 * time.Second,
			TLSHandshakeTimeout:   10 * time.Second,
			ExpectContinueTimeout: 1 * time.Second,
			DialContext: (&net.Dialer{
				Timeout:   5 * time.Second,
				KeepAlive: 30 * time.Second,
			}).DialContext,
		}

		if cfg.ProxyURL != "" {
			noProxyList := strings.Split(cfg.NoProxy, ",")
			transport.Proxy = func(req *http.Request) (*url.URL, error) {
				for _, noProxy := range noProxyList {
					if strings.Contains(req.URL.Hostname(), strings.TrimSpace(noProxy)) {
						return nil, nil
					}
				}
				return url.Parse(cfg.ProxyURL)
			}
		}

		// Wrap with circuit breaker transport
		cbTransport := &GoBreakerTransport{
			CB:        circuitBreaker,
			Transport: transport,
		}

		client := resty.New().
			SetTransport(cbTransport).
			SetRetryCount(cfg.RetryCount).
			SetRetryMaxWaitTime(cfg.RetryMaxWait).
			SetTimeout(15 * time.Second).
			AddRetryCondition(func(r *resty.Response, err error) bool {
				return ShouldRetry(r, err)
			}).
			OnBeforeRequest(func(c *resty.Client, req *resty.Request) error {
				req.SetHeader("Accept-Encoding", "gzip")
				return nil
			})

		httpClientInstance = &HttpClient{
			RestyClient:    client,
			CircuitBreaker: circuitBreaker,
		}
	})

	return httpClientInstance
}

// ShouldRetry determines if a request should be retried based on the response or error
func ShouldRetry(resp *resty.Response, err error) (shouldRetry bool) {

	// Safety net: prevent panics from propagating
	defer func() {
		if r := recover(); r != nil {
			shouldRetry = false // fail-safe: don't retry
		}
	}()

	// Handle network errors
	if err != nil {

		// Check if the error is a network timeout
		var netErr net.Error
		if errors.As(err, &netErr) {
			if netErr.Timeout() {
				return true
			}
		}

		// Check for connection refused
		if strings.Contains(err.Error(), "connect: connection refused") {
			return true
		}
	}

	// Handle HTTP response
	if resp != nil {
		statusCode := resp.StatusCode()

		// Retry if status code is 0 (may indicate connection issue or no response)
		if statusCode == 0 {
			return true
		}

		// Retry on 5xx server errors (transient issues)
		if statusCode >= 500 && statusCode < 600 {
			return true
		}

		// Do not retry on 429 Too Many Requests (respect rate limiting)
		if statusCode == http.StatusTooManyRequests {
			return false
		}

		if statusCode == http.StatusConflict {

			if b := resp.Body(); len(b) > 0 {
				body := strings.ToLower(string(b))
				if strings.Contains(body, "concurrent modification") || strings.Contains(body, "/concurrent-modification") {
					return true
				}
			}

			// If 409 but not concurrent modification → don't retry
			// zl.Info().Msg("Not retrying 409 since it's not a concurrent modification")
			return false
		}
	}

	// Default: do not retry
	return false
}

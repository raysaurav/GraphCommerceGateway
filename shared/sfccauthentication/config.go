package sfccauthentication

import (
	"context"
	"log"
	"sync"

	"github.com/sethvargo/go-envconfig"
)

type AuthConfig struct {
	ShortCode                string `env:"SFCC_SHORT_CODE,required" validate:"required"`
	OrganizationID           string `env:"SFCC_ORG_ID,required" validate:"required"`
	BaseURL                  string `env:"SFCC_BASE_URL,required" validate:"required"`
	Username                 string `env:"SFCC_USERNAME,required" validate:"required"`
	Password                 string `env:"SFCC_PASSWORD,required" validate:"required"`
	GrantType                string `env:"SFCC_GRANT_TYPE,required" validate:"required"`
	RedirectURI              string `env:"SFCC_REDIRECT_URI,required" validate:"required"`
	SiteId                   string `env:"SFCC_SITE_ID,required" validate:"required"`
	Hint                     string `env:"SFCC_HINT,required" validate:"required"`
	IDPOrigin                string `env:"SFCC_IPD_ORIGIN,required" validate:"required"`
	ChannelID                string `env:"SFCC_CHANNEL_ID,required" validate:"required"`
	AuthRedisTTL             int    `env:"SFCC_REDIS_TTL,required" validate:"required"`
	Environment              string `env:"DBT_ENVIRONMENT,default=DEV"`
	AsyncTokenRefreshEnabler bool   `env:"TOKEN_REFRESH_FLAG,default=false"`
	RefreshLockTTL           int    `env:"REFRESH_LOCK_TTL,default=10"`
	TokenRedisRetryCount     int    `env:"TOKEN_REDIS_RETRY_COUNT,default=2"`
	TokenRedisRetryDelay     int    `env:"TOKEN_REDIS_RETRY_DELAY,default=200"`
}

var (
	configOnce     sync.Once
	configInstance *AuthConfig
)

func NewConfig() (*AuthConfig, error) {
	var err error
	configOnce.Do(func() {
		var cfg AuthConfig
		err = envconfig.Process(context.Background(), &cfg)
		if err == nil {
			configInstance = &cfg
		} else {
			log.Printf("Error processing environment config: %v", err)
		}
	})
	return configInstance, err
}

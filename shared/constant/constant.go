package constant

import "time"

const (
	FormatPretty = "pretty"
	FormatJSON   = "json"

	CorrelationIDHeader = "X-Correlation-ID"
	UsidHeader          = "x-custom-usid"
	GinContextKey       = "GinContextKey"
	// For - Http Server Configuration
	ReadHeaderTimeout = 5 * time.Second   // Set a read header timeout to mitigate Slowloris attacks
	ReadTimeout       = 30 * time.Second  // Maximum duration for reading the entire request
	WriteTimeout      = 30 * time.Second  // Maximum duration before timing out writes of the response
	IdleTimeout       = 120 * time.Second // Maximum amount of time to wait for the next requestt
	XForwardedForKey  = "x-forwarded-for" //Auth Lib constants
	CONTENT_TYPE      = "application/x-www-form-urlencoded"

	// SFCC sfcc_authentication parameters
	GRANT_TYPE                        = "grant_type"
	REDIRECT_URI                      = "redirect_uri"
	CHANNEL_ID_NAME                   = "channel_id"
	AUTHORIZATION                     = "Authorization"
	KEY_CHANNEL_ID                    = "coppel"
	CONSTANT_TYPE_NAME                = "Content-Type"
	ACCESS_TOKEN                      = "x-access-token"
	CUSTOMER_ID                       = "x-customer-id"
	REFRESH_TOKEN                     = "x-refresh-token"
	BASKET_ID                         = "x-basket-id"
	CMS_ERROR_UID                     = "error_labels"
	CMS_ENV                           = "develop"
	CMS_MAPPER_MOD                    = "cms_error_mapper"
	EMPTY_STRING                      = ""
	BASKET_ERROR_LABELS               = "basket_error_labels"
	ACCOUNT_ERROR_LABELS              = "account_error_labels"
	CHECKOUT_ERROR_LABELS             = "checkout_error_labels"
	BROWSE_ERROR_LABELS               = "browse_error_labels"
	PRODUCT_ERROR_LABELS              = "product_error_labels"
	XLivePreviewKey                   = "x-live-preview"
	XChannelKey                       = "x-channel"
	XStoreIdKey                       = "x-store-id"
	XUserAgentKey                     = "user-agent"
	XAppSessionIDKey                  = "x-app-session-id"
	RECAPTCHA_GRAPHQL_QUERY_REGEX     = `(query|mutation)\s+\w+\s*\(([^)]+)\)\s*{\s*(\w+)`
	X_RECAPTCHA_ACTION_HEADER         = "X-Recaptcha-Action"
	X_RECAPTCHA_TOKEN_HEADER          = "X-Recaptcha-Token"
	RECAPTCHA_VALIDATED_FOR_USER      = "userRecaptchaValidated"
	RECAPTCHA_VALIDATION_FAILED       = "reCAPTCHA validation failed"
	RECAPTCHA_ALLOWED_IP              = "10.149.95.242,10.148.108.54"
	RECAPTCHA_CACHE_TTL               = 24 * time.Hour
	ENVIRONMENT                       = "x-env"
	ERR_RECAPTCHA_SCORE_LESS_THAN     = "the reCAPTCHA score is less than threshold"
	REDIS_RECAPTCHA_BLOCK_TYPE        = "recaptchaValidationFailed"
	ERR_RECAPTCHA_ERROR_CODE          = "RECAPTCHA_ERROR"
	MOCK_ENVIRONMENT                  = "skip_recaptcha"
	SFCC_GUEST_TOKEN_NS               = "sfcc:token:guest:"
	SFCC_AKAMAI_TOKEN_NS              = "sfcc:token:akamai:"
	SFCC_REGISTERED_CUSTOMER_TOKEN_NS = "sfcc:token:rc:"
	XChannelIOSValue                  = "app_IOS"
	XChannelAndroidValue              = "app_Android"
	XChannelWebValue                  = "web"
	XCacheCDN                         = "x-cache-cdn"
	DefaultErrorMessage               = "Algo salió mal. Intenta más tarde"
	GUEST                             = "GUEST"
	REGISTERED                        = "REGISTERED"

	RedisError = "redis-error"
)

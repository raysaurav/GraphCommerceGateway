package config

type Config struct {
	MiscConfig `env:",prefix=BASE_"`
	SFCCConfig `env:",prefix=SFCC_"`
}

type MiscConfig struct {
	DisableProductService bool `env:"DISABLE_PRODUCT_SERVICE"`
}

type SFCCConfig struct {
	Tenant       string `env:"TENANT,required"`
	ShortCode    string `env:"SHORT_CODE,required"`
	BaseUrl      string `env:"BASE_URL,required"`
	OrgId        string `env:"ORGANIZATION_ID,required"`
	ApiVersion   string `env:"API_VERSION,default=v1"`
	GrantType    string `env:"GRANT_TYPE,required"`
	Hint         string `env:"HINT,required"`
	IDPOrigin    string `env:"IDP_ORIGIN,required"`
	ClientId     string `env:"CLIENT_ID,required"`
	ClientSecret string `env:"CLIENT_SECRET,required"`
	SiteId       string `env:"SITE_ID,default=coppel"`
}

package appinfo

// AppInfo holds the application configuration loaded from environment variables
type AppInfo struct {
	ServiceName string `env:"SERVICE_NAME"`
	Port        string `env:"PORT,default=8080"`
	Environment string `env:"ENVIRONMENT,default=dev"`
	Version     string `env:"SERVICE_VERSION"`
}

// NewAppInfo creates a new instance of AppInfo struct
func NewAppInfo() *AppInfo {
	return &AppInfo{}
}
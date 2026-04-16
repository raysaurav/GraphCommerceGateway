package appinfo

// Global variable to access the singleton handler
var AppInfoHandler AppInfoInterface

// init function automatically initializes the singleton handler when package is imported
func initHandler() {
	AppInfoHandler = GetHandler()
}

// InitializeAppInfo provides the entry point to get AppInfo instance with loaded environment variables
func InitializeAppInfo() (*AppInfo, error) {
	initHandler()

	appInfoInstance, err := AppInfoHandler.GetAppInfo()
	if err != nil {
		return nil, err
	}
	return appInfoInstance, nil
}

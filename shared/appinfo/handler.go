package appinfo

import (
	"context"
	"sync"

	"github.com/sethvargo/go-envconfig"
)

// AppInfoInterface defines the contract for AppInfo operations
type AppInfoInterface interface {
	GetAppInfo() (*AppInfo, error)
}

// appInfoHandler implements AppInfoInterface with singleton pattern
type appInfoHandler struct {
	appInfo *AppInfo
	once    sync.Once
}

// GetAppInfo implements AppInfoInterface and returns AppInfo with loaded environment variables
func (h *appInfoHandler) GetAppInfo() (*AppInfo, error) {
	var err error
	h.once.Do(func() {
		h.appInfo = NewAppInfo()
		err = envconfig.Process(context.Background(), h.appInfo)
	})
	if err != nil {
		return nil, err
	}
	return h.appInfo, nil
}

// Global singleton instance
var (
	handlerInstance *appInfoHandler
	handlerOnce     sync.Once
)

// GetHandler returns the singleton instance of AppInfoInterface
func GetHandler() AppInfoInterface {
	handlerOnce.Do(func() {
		handlerInstance = &appInfoHandler{}
	})
	return handlerInstance
}

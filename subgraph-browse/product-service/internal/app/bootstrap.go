package app

import (
	ginMain "github.com/gin-gonic/gin"
	"github.com/raysaurav/GraphCommerceGateway/shared/appinfo"
	"github.com/raysaurav/GraphCommerceGateway/shared/gin"
)

type Bootstrap struct {
	AppInfo *appinfo.AppInfo
	Engine  *ginMain.Engine
	Handler *gin.HandlerInitilization
}

func NewBootstrap(appInfo *appinfo.AppInfo,
	engine *ginMain.Engine,
	handler *gin.HandlerInitilization) *Bootstrap {
	return &Bootstrap{
		AppInfo: appInfo,
		Engine:  engine,
		Handler: handler,
	}
}

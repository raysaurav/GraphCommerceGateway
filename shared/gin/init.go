package gin

import (
	"io"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/raysaurav/GraphCommerceGateway/shared/appinfo"
)

var (
	ginEngineInstance *gin.Engine
	singleInstance    sync.Once
)

func NewGinEngine(cfg *appinfo.AppInfo) *gin.Engine {
	singleInstance.Do(func() {
		mode := gin.DebugMode

		if cfg.Environment != "" {
			mode = cfg.Environment
		}

		gin.SetMode(mode)

		newEngine := gin.New()

		// Disable Gin's default logging by redirecting log outputs to io.Discard.
		// This silences all informational and error logs produced by Gin. If you want
		// to enable logging for debugging or audit, remove or modify these assignments.
		gin.DefaultWriter = io.Discard
		gin.DefaultErrorWriter = io.Discard

		newEngine.Use(MapJSONData())
		newEngine.Use(ErrorHandler())
		newEngine.Use(gin.Recovery())

		ginEngineInstance = newEngine
	})
	return ginEngineInstance
}

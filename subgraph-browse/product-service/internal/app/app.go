package app

import (
	"github.com/raysaurav/GraphCommerceGateway/shared/appinfo"
	"github.com/raysaurav/GraphCommerceGateway/shared/gin"
	"github.com/raysaurav/GraphCommerceGateway/subgraph-browse/product-service/provider"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(
		appinfo.InitializeAppInfo,

		gin.NewGinEngine,
		gin.NewHandlerInitializer,

		// httpclient.InitHttpClient,

		// sfccauthentication.NewAuthClient,

		provider.LoadConfig,
		// provider.NewBrowseService,
		provider.NewHandler,

		NewBootstrap,
	),
	fx.Invoke(StartServer),
)

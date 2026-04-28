package provider

import (
	"context"
	"fmt"
	"sync"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/raysaurav/GraphCommerceGateway/shared/config"
	"github.com/raysaurav/GraphCommerceGateway/shared/httpclient"
	"github.com/raysaurav/GraphCommerceGateway/subgraph-browse/product-service/graph"
	"github.com/raysaurav/GraphCommerceGateway/subgraph-browse/product-service/internal/domain"
	"github.com/raysaurav/GraphCommerceGateway/subgraph-browse/product-service/internal/service"
	"github.com/sethvargo/go-envconfig"
)

var (
	svc     domain.BrowseService
	svcOnce sync.Once

	configInstance *config.Config
	configOnce     sync.Once
)

func LoadConfig() *config.Config {
	configOnce.Do(func() {
		var cfg config.Config
		err := envconfig.Process(context.Background(), &cfg)
		if err != nil {
			fmt.Println("Issue in loading environment files")
		}
		configInstance = &cfg
	})
	return configInstance
}

func NewBrowseService(cfg *config.Config, HttpClient httpclient.HttpClientInterface) domain.BrowseService {
	svcOnce.Do(func() {
		svc = &service.Service{
			HttpClient: HttpClient,
			Config:     cfg,
		}
	})
	return svc
}

func NewHandler() *handler.Server {
	srv := handler.NewDefaultServer(graph.NewExecutableSchema(
		graph.Config{
			Resolvers: &graph.Resolver{},
		},
	))

	var mb int64 = 1 << 20
	srv.AddTransport(transport.MultipartForm{MaxMemory: 50 * mb, MaxUploadSize: 25 * mb})
	return srv
}

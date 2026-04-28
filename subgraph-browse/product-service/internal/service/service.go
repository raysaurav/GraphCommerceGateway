package service

import (
	"context"

	"github.com/go-resty/resty/v2"
	"github.com/raysaurav/GraphCommerceGateway/shared/config"
	"github.com/raysaurav/GraphCommerceGateway/shared/httpclient"
	"github.com/raysaurav/GraphCommerceGateway/subgraph-browse/product-service/graph/model"
)

type Service struct {
	RestyClient *resty.Client
	HttpClient  httpclient.HttpClientInterface
	Config      *config.Config
}

func (s *Service) GetProductDetails(ctx context.Context, productId string) *model.ProductDetails {
	return &model.ProductDetails{
		ProductID:   "123",
		ProductName: "My Product",
	}
}

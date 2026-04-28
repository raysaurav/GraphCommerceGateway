package domain

import (
	"context"

	"github.com/raysaurav/GraphCommerceGateway/subgraph-browse/product-service/graph/model"
)

type BrowseService interface {
	GetProductDetails(ctx context.Context, productId string) *model.ProductDetails
}

package graph

//go:generate go tool gqlgen generate
import (
	"github.com/raysaurav/GraphCommerceGateway/subgraph-browse/product-service/graph/model"
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require
// here.

type Resolver struct {
	todos []*model.Todo
	// BrowseService domain.BrowseService
}

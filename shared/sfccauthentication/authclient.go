package sfccauthentication

import (
	"context"

	"github.com/raysaurav/GraphCommerceGateway/shared/sfccauthentication/model"
)

type AuthClient interface {
	TokenFetchStrategy(ctx context.Context) (*model.TokenInfo, error)
}

type AuthWrapper struct {
	httpClient httpclient.HttpClientInterface
}

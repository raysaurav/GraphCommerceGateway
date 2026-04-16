module github.com/raysaurav/GraphCommerceGateway/subgraph-browse/product-service

go 1.25.1

require (
	github.com/joho/godotenv v1.5.1
	github.com/raysaurav/GraphCommerceGateway/shared/config v0.0.0-00010101000000-000000000000
	github.com/sethvargo/go-envconfig v1.3.0
)

require (
	github.com/go-viper/mapstructure/v2 v2.5.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/hashicorp/golang-lru/v2 v2.0.7 // indirect
)

replace (
	github.com/raysaurav/GraphCommerceGateway/shared => ../../shared
	github.com/raysaurav/GraphCommerceGateway/shared/config => ../../shared/config
)

require (
	github.com/99designs/gqlgen v0.17.89
	github.com/agnivade/levenshtein v1.2.1 // indirect
	github.com/google/uuid v1.6.0 // indirect
	github.com/sosodev/duration v1.4.0 // indirect
	github.com/vektah/gqlparser/v2 v2.5.32
	golang.org/x/sync v0.20.0 // indirect
)

//go:generate go run github.com/99designs/gqlgen generate

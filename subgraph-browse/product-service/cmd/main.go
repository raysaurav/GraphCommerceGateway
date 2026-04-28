package main

import (
	"context"
	"fmt"

	"github.com/joho/godotenv"
	"github.com/raysaurav/GraphCommerceGateway/shared/config"
	"github.com/raysaurav/GraphCommerceGateway/subgraph-browse/product-service/internal/app"
	"github.com/sethvargo/go-envconfig"
	"go.uber.org/fx"
)

func loadEnvironmentVariables() {
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Printf("Error loading .env file: %v", err)
	}
	fmt.Println()
}

func main() {
	loadEnvironmentVariables()

	var cfg config.Config
	if err := envconfig.Process(context.Background(), &cfg); err != nil {
		fmt.Printf("Error parsing environment variables: %v\n", err)
		return
	}
	fmt.Printf("Config initialized: %+v\n", cfg.ClientId)

	fx.New(
		app.Module,
	).Run()
}

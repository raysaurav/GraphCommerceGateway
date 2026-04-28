package app

import (
	"context"
	"fmt"

	"go.uber.org/fx"
)

func StartServer(lifeCycle fx.Lifecycle, bootstrap *Bootstrap) {
	lifeCycle.Append(
		fx.Hook{
			OnStart: func(ctx context.Context) error {
				bootstrap.Handler.RegisterHandler(bootstrap.Engine)
				fmt.Println("SFCC Browse Service started on port : " + bootstrap.AppInfo.Port)
				go func() {
					_ = bootstrap.Engine.Run(":" + (bootstrap.AppInfo.Port))
				}()
				return nil
			},
			OnStop: func(ctx context.Context) error {
				fmt.Println("SFCC Browse Service Stopped..")
				return nil
			},
		},
	)
}

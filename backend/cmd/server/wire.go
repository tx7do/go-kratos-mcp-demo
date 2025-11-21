//go:build wireinject
// +build wireinject

// The build tag makes sure the stub is not built in the final build.

package main

import (
	"github.com/google/wire"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/log"

	"go-kratos-mcp-demo/internal/server"
	"go-kratos-mcp-demo/internal/service"

	"go-kratos-mcp-demo/api/gen/go/conf"
)

// initApp init kratos application.
func initApp(log.Logger, *conf.RecommendConfig) (*kratos.App, func(), error) {
	panic(wire.Build(server.ProviderSet, service.ProviderSet, newApp))
}

package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/go-kratos/kratos/v2"
	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"

	"go-kratos-mcp-demo/api/gen/go/conf"

	"github.com/go-kratos/kratos/v2/transport/http"
	"github.com/tx7do/kratos-transport/transport/mcp"
)

// go build -ldflags "-X main.Version=x.y.z"
var (
	// Name is the name of the compiled software.
	Name string
	// Version is the version of the compiled software.
	Version string
	// flagconf is the config flag.
	flagconf string
)

func init() {
	flag.StringVar(&flagconf, "conf", "./configs", "config path, eg: -conf config.yaml")
}

func newApp(
	lg log.Logger,
	ms *mcp.Server,
	hs *http.Server,
) *kratos.App {
	app := kratos.New(
		kratos.Name("recommend-mcp-server"),
		kratos.Version(Version),
		kratos.Logger(lg),
		kratos.Server(
			ms,
			hs,
		),
	)

	return app
}

func main() {
	logger := log.NewStdLogger(os.Stdout)

	cfg := config.New(
		config.WithSource(
			file.NewSource(flagconf),
		),
	)

	if err := cfg.Load(); err != nil {
		panic(err)
	}

	var rc conf.RecommendConfig
	if err := cfg.Scan(&rc); err != nil {
		panic(err)
	}

	app, cleanup, err := initApp(logger, &rc)
	if err != nil {
		panic(err)
	}
	defer cleanup()

	// start and wait for stop signal
	if err := app.Run(); err != nil {
		fmt.Println(err)
		panic(err)
	}
}

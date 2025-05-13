package main

import (
	"context"
	"flag"
	"fmt"
	"go-map-proxy/internal/config"
	"go-map-proxy/internal/handler"
	"go-map-proxy/internal/middleware"
	"go-map-proxy/internal/utils"
	"go-map-proxy/pkg/logger"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

var (
	VERSION    = "0.0.1"
	BUILD_TIME = "2025-05-10T00:00:00Z"
)

var configPath string

func init() {
	// -c <config file path, default is ./config.yaml>
	flag.StringVar(&configPath, "c", "config.yaml", "config file path")

	flag.Usage = func() {
		fmt.Printf("go-map-proxy version: %s, build time: %s\n", VERSION, BUILD_TIME)
		flag.PrintDefaults()
	}

	flag.Parse()

	if err := config.InitConfig(configPath); err != nil {
		logger.Fatalf("init config failed: %v", err)
	}

	// init map cache
	utils.InitMapCache(config.Cfg.Cache.Path)

	fmt.Printf("config: %+v\n", config.Cfg)
}

func StartServer() {
	// setup echo server
	e := echo.New()
	e.Logger.SetLevel(log.INFO)

	// register middlewares
	middleware.RegisterMiddleware(e)

	// register handlers
	handler.RegisterHandlers(e)

	// graceful shutdown
	// listen exit signal
	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	address := fmt.Sprintf("%s:%d", config.Cfg.Server.Host, config.Cfg.Server.Port)

	// start server
	go func() {
		if err := e.Start(address); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	// wait for interrupt signal to gracefully shut down the server with a timeout of 10 seconds.
	<-signalCtx.Done()
	logger.Infof("shutting down the server gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
}

func main() {

	// init logger
	logger.InitLogger(&logger.LoggerCfg{
		LogLevel:   config.Cfg.Log.Level,
		EnableFile: config.Cfg.Log.EnableFile,
		LogPath:    config.Cfg.Log.FilePath,
	})

	// start server
	StartServer()

}

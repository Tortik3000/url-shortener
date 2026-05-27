package main

import (
	"log"

	"go.uber.org/zap"

	"github.com/edkin/url-shortener/internal/app"
	"github.com/edkin/url-shortener/internal/config"
	"github.com/edkin/url-shortener/pkg/logger"
)

func main() {
	zapLogger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("failed to initialize logger: %v", err)
	}
	logs := logger.NewZap(zapLogger)

	cfg, err := config.New()
	if err != nil {
		logs.Fatal("failed to load config", logger.Error(err))
	}

	app.Run(logs, cfg)
}

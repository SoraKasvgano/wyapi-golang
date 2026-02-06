package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	assets "wyapi-golang"
	"wyapi-golang/internal/api"
	"wyapi-golang/internal/config"
	"wyapi-golang/internal/cookie"
	"wyapi-golang/internal/downloader"
	"wyapi-golang/internal/netease"

	httpSwagger "github.com/swaggo/http-swagger"
)

func main() {
	configPath := flag.String("config", "config.json", "config file path")
	flag.Parse()

	cfg, created, err := config.LoadOrCreate(*configPath)
	if err != nil {
		panic(err)
	}

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: parseLogLevel(cfg.Log.Level)}))
	slog.SetDefault(logger)

	cookieManager := cookie.NewManager(cfg.Cookie.File)
	if err := cookieManager.EnsureFile(); err != nil {
		logger.Error("failed to ensure cookie file", slog.String("error", err.Error()))
	}

	neteaseClient := netease.NewClient(time.Duration(cfg.Server.RequestTimeoutSeconds) * time.Second)
	downloaderSvc := downloader.NewDownloader(neteaseClient, cookieManager, cfg.Download.Dir)

	openAPIData, _ := fs.ReadFile(assets.OpenAPI, "docs/openapi.json")

	frontendFS, err := fs.Sub(assets.Frontend, "frontend/dist")
	if err != nil {
		logger.Warn("frontend assets not found", slog.String("error", err.Error()))
	}

	var staticHandler http.Handler
	if err == nil {
		staticHandler = api.NewSPAHandler(frontendFS, "index.html")
	}

	swaggerHandler := httpSwagger.Handler(httpSwagger.URL("/openapi.json"))

	handler := api.NewHandler(cfg, neteaseClient, cookieManager, downloaderSvc, openAPIData)
	router := api.NewRouter(handler, cfg, staticHandler, swaggerHandler)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeoutSeconds) * time.Second,
		IdleTimeout:  time.Duration(cfg.Server.IdleTimeoutSeconds) * time.Second,
	}

	logger.Info("WyAPI-Golang starting", slog.String("addr", addr))
	if created {
		logger.Info("config.json created", slog.String("api_token", cfg.Security.APIToken))
	} else if cfg.Security.APIToken != "" {
		logger.Info("api token loaded", slog.String("api_token", cfg.Security.APIToken))
	}

	if cfg.Security.RequireToken {
		logger.Info("api token required for protected routes")
	} else {
		logger.Warn("api token enforcement is disabled; enable require_token in config.json for security")
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("server stopped", slog.String("error", err.Error()))
	}
}

func parseLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

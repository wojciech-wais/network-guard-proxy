package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/networkguard/proxy/internal/api"
	"github.com/networkguard/proxy/internal/config"
	"github.com/networkguard/proxy/internal/proxy"
	"github.com/networkguard/proxy/internal/storage"
)

const version = "0.1.0"

func main() {
	configPath := flag.String("config", "config.yaml", "path to configuration file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	store, err := storage.NewSQLiteStore(cfg.Storage.DSN)
	if err != nil {
		log.Fatalf("Failed to init storage: %v", err)
	}
	defer store.Close()

	// Resolve web directory relative to executable
	webDir := findWebDir()

	// Admin server
	adminAPI := api.NewServer(store, version, webDir)
	adminServer := &http.Server{
		Addr:    cfg.Admin.ListenAddress,
		Handler: adminAPI,
	}

	// Proxy server
	proxyHandler, err := proxy.NewProxyServer(cfg.Proxy.UpstreamURL, store, cfg.Logging.MaxLogs)
	if err != nil {
		log.Fatalf("Failed to create proxy: %v", err)
	}
	proxyServer := &http.Server{
		Addr:    cfg.Proxy.ListenAddress,
		Handler: proxyHandler,
	}

	// Start servers
	go func() {
		log.Printf("Admin server listening on %s", cfg.Admin.ListenAddress)
		if err := adminServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Admin server error: %v", err)
		}
	}()

	go func() {
		log.Printf("Proxy server listening on %s (upstream: %s)", cfg.Proxy.ListenAddress, cfg.Proxy.UpstreamURL)
		if err := proxyServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Proxy server error: %v", err)
		}
	}()

	log.Printf("Network Guard Proxy v%s started", version)

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	adminServer.Shutdown(ctx)
	proxyServer.Shutdown(ctx)
	log.Println("Shutdown complete")
}

func findWebDir() string {
	// Try relative to working directory first
	if info, err := os.Stat("web"); err == nil && info.IsDir() {
		return "web"
	}
	// Try relative to executable
	exe, err := os.Executable()
	if err == nil {
		dir := filepath.Join(filepath.Dir(exe), "web")
		if info, err := os.Stat(dir); err == nil && info.IsDir() {
			return dir
		}
	}
	return ""
}

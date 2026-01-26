package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ollama2openai/config"
	"ollama2openai/middleware"
	"ollama2openai/ollama"
	"ollama2openai/pkg/logger"
	"ollama2openai/router"
)

func main() {
	// Load configuration
	configPath := "config/config.yaml"
	if envPath := os.Getenv("CONFIG_PATH"); envPath != "" {
		configPath = envPath
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	log.Printf("Starting Ollama2OpenAI Proxy on %s", cfg.GetAddress())
	log.Printf("Ollama URL: %s", cfg.OllamaURL)

	// Initialize logger
	logLevel := logger.ParseLevel(cfg.LogLevel)
	appLogger := logger.NewStdLogger(logLevel)
	appLogger.Info("Logger initialized", logger.String("level", cfg.LogLevel))

	// Verify Ollama connection and print models
	if err := verifyOllamaConnection(cfg.OllamaURL); err != nil {
		log.Fatalf("Ollama connection failed: %v", err)
	}

	// Create dependencies
	ollamaClient := ollama.NewClient(cfg.OllamaURL, cfg.GetTimeout())
	usageTracker := middleware.NewUsageStats()

	// Create a custom ServeMux to handle routes
	mux := http.NewServeMux()

	// Setup routes with dependency injection
	rt := router.NewRouter(cfg, ollamaClient, usageTracker, appLogger)
	rt.SetupRoutes(mux)

	// Create server
	server := &http.Server{
		Addr: cfg.GetAddress(),
		Handler: middleware.Recovery(appLogger)(
			middleware.RequestID(
				middleware.LoggingMiddleware(appLogger)(
					middleware.WithAuth(mux, cfg),
				),
			),
		),
		ReadTimeout:  cfg.GetTimeout(),
		WriteTimeout: cfg.GetTimeout(),
	}

	// Start server in a goroutine
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	log.Printf("Server started successfully")

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	if err := server.Close(); err != nil {
		log.Printf("Error during server shutdown: %v", err)
	}

	fmt.Println("Server exited gracefully")
}

// verifyOllamaConnection checks if Ollama is running and prints available models
func verifyOllamaConnection(ollamaURL string) error {
	log.Printf("Verifying Ollama connection...")

	client := ollama.NewClient(ollamaURL, 10*time.Second)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	resp, err := client.Tags(ctx)
	if err != nil {
		return fmt.Errorf("cannot connect to Ollama at %s: %w", ollamaURL, err)
	}

	if len(resp.Models) == 0 {
		log.Printf("Warning: No models found in Ollama")
	} else {
		log.Printf("Ollama is connected. Available models:")
		for _, m := range resp.Models {
			log.Printf("  - %s", m.Name)
		}
	}

	return nil
}

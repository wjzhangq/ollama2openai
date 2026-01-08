package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"ollama2openai/config"
	"ollama2openai/middleware"
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

	// Create a custom ServeMux to handle routes
	mux := http.NewServeMux()

	// Setup routes
	router.SetupRoutes(mux, cfg)

	// Create server
	server := &http.Server{
		Addr:         cfg.GetAddress(),
		Handler:      middleware.WithLogging(middleware.WithAuth(mux, cfg)),
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

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	config, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	configureLogging(config.LogLevel)
	server := newPrintServer(config)
	address := config.address()
	httpServer := &http.Server{
		Addr:              address,
		Handler:           server.Handler(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       time.Duration(config.ReadTimeoutSeconds) * time.Second,
		WriteTimeout:      time.Duration(config.WriteTimeoutSeconds) * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	serverErrors := make(chan error, 1)

	log.Printf("starting server on %s", address)
	go func() {
		serverErrors <- httpServer.ListenAndServe()
	}()

	shutdownContext, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	select {
	case err := <-serverErrors:
		if err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	case <-shutdownContext.Done():
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := httpServer.Shutdown(ctx); err != nil {
			log.Printf("graceful shutdown failed: %v", err)
			if closeErr := httpServer.Close(); closeErr != nil {
				log.Printf("server close failed: %v", closeErr)
			}
		}
	}
}

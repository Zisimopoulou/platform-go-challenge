package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/Zisimopoulou/platform-go-challenge/internal/api"
	"github.com/Zisimopoulou/platform-go-challenge/internal/core"
	"github.com/Zisimopoulou/platform-go-challenge/internal/data"
)

func main() {
	// Provide a JWT secret via environment variable JWT_SECRET (for example: "mysecret")
	if os.Getenv("JWT_SECRET") == "" {
		log.Println("JWT_SECRET not set, using default 'dev-secret' (not safe for production)")
		os.Setenv("JWT_SECRET", "dev-secret")
	}

	// Wire dependencies
	store := data.NewInMemoryStore()
	svc := core.NewService(store)
	h := api.NewHandler(svc)

	mux := http.NewServeMux()
	mux.Handle("/users/", http.StripPrefix("/users", h))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK); w.Write([]byte("ok")) })
	mux.HandleFunc("/auth/login", api.LoginHandler) // new login endpoint

	addr := ":8080"
	srv := &http.Server{
		Addr:    addr,
		Handler: api.WithMiddleware(mux),
	}

	// graceful shutdown
	go func() {
		log.Printf("server listening on %s", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("server forced to shutdown: %v", err)
	}
	log.Println("server stopped")
}

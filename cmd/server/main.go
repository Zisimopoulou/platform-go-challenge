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
	if os.Getenv("JWT_SECRET") == "" && os.Getenv("APP_JWT_SECRET") == "" {
		log.Fatal("JWT_SECRET or APP_JWT_SECRET environment variable must be set")
	}

	store := data.NewInMemoryStore()
	svc := core.NewService(store)
	h := api.NewHandler(svc)

	mux := http.NewServeMux()
	mux.Handle("/users/", http.StripPrefix("/users", h))
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK); w.Write([]byte("ok")) })
	mux.HandleFunc("/auth/login", api.LoginHandler)
	mux.HandleFunc("/auth/refresh", api.RefreshHandler)

	addr := ":8080"
	srv := &http.Server{
		Addr:    addr,
		Handler: api.WithMiddleware(mux),
	}

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

package tests

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"

	"github.com/Zisimopoulou/platform-go-challenge/internal/api"
	"github.com/Zisimopoulou/platform-go-challenge/internal/core"
	"github.com/Zisimopoulou/platform-go-challenge/internal/data"
)

func TestConcurrentAccess(t *testing.T) {
	// Provide default JWT secret for testing
	if os.Getenv("JWT_SECRET") == "" {
		log.Println("JWT_SECRET not set, using default 'dev-secret' (not safe for production)")
		os.Setenv("JWT_SECRET", "dev-secret")
	}

	// Initialize in-memory store and service
	store := data.NewInMemoryStore()
	svc := core.NewService(store)
	h := api.NewHandler(svc)

	// Wrap handler with AuthMiddleware
	mux := http.NewServeMux()
	mux.Handle("/", api.AuthMiddleware(h))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	// Test user and JWT
	user := "concurrentuser"
	token, err := api.GenerateToken(user)
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{}

	// Add initial favorite
	addReq := map[string]interface{}{
		"type":        "insight",
		"description": "initial",
		"payload":     map[string]string{"text": "test"},
	}
	b, _ := json.Marshal(addReq)
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/"+user+"/favorites", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()

	// Concurrent reads
	var wg sync.WaitGroup
	numRequests := 20

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, _ := http.NewRequest(http.MethodGet, srv.URL+"/"+user+"/favorites", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			res, err := client.Do(req)
			if err != nil {
				t.Error("Concurrent request failed:", err)
				return
			}
			defer res.Body.Close()

			if res.StatusCode != http.StatusOK {
				t.Errorf("Concurrent request failed with status: %d", res.StatusCode)
			}
		}()
	}

	wg.Wait()
}

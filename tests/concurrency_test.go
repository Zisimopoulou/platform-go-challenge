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

	"github.com/Zisimopoulou/platform-go-challenge/internal/core"
	"github.com/Zisimopoulou/platform-go-challenge/internal/api"
	"github.com/Zisimopoulou/platform-go-challenge/internal/data"
)

func TestConcurrentAccess(t *testing.T) {
	if os.Getenv("JWT_SECRET") == "" {
		log.Println("JWT_SECRET not set, using default 'dev-secret' (not safe for production)")
		os.Setenv("JWT_SECRET", "dev-secret")
	}

	store := data.NewInMemoryStore()
	svc := core.NewService(store)
	
	mux := http.NewServeMux()
	mux.Handle("/users/", http.StripPrefix("/users", api.NewHandler(svc)))
	mux.HandleFunc("/auth/login", api.LoginHandler)
	
	handler := api.WithMiddleware(mux)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	user := "concurrentuser"
	
	loginReq := map[string]string{"userId": user}
	loginBody, _ := json.Marshal(loginReq)
	loginResp, err := http.Post(srv.URL+"/auth/login", "application/json", bytes.NewReader(loginBody))
	if err != nil {
		t.Fatal(err)
	}
	defer loginResp.Body.Close()

	var tokenResp api.TokenResponse
	json.NewDecoder(loginResp.Body).Decode(&tokenResp)
	accessToken := tokenResp.AccessToken

	client := &http.Client{}

	addReq := map[string]interface{}{
		"type":        "insight",
		"description": "initial",
		"payload":     map[string]string{"text": "test"},
	}
	b, _ := json.Marshal(addReq)
	req, _ := http.NewRequest(http.MethodPost, srv.URL+"/users/"+user+"/favorites", bytes.NewReader(b))
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Content-Type", "application/json")
	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()

	var wg sync.WaitGroup
	numRequests := 20

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, _ := http.NewRequest(http.MethodGet, srv.URL+"/users/"+user+"/favorites", nil)
			req.Header.Set("Authorization", "Bearer "+accessToken)
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
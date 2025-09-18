package tests

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Zisimopoulou/platform-go-challenge/internal/core"
	"github.com/Zisimopoulou/platform-go-challenge/internal/data"
	"github.com/Zisimopoulou/platform-go-challenge/internal/api"
)

func TestAuthScenarios(t *testing.T) {
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

	client := &http.Client{}

	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/users/user123/favorites", nil)
	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected 401 for no auth header, got %d", res.StatusCode)
	}

	req, _ = http.NewRequest(http.MethodGet, srv.URL+"/users/user123/favorites", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	res, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected 401 for invalid format, got %d", res.StatusCode)
	}

	loginReq := map[string]string{"userId": "user1"}
	loginBody, _ := json.Marshal(loginReq)
	loginResp, err := http.Post(srv.URL+"/auth/login", "application/json", bytes.NewReader(loginBody))
	if err != nil {
		t.Fatal(err)
	}
	defer loginResp.Body.Close()

	var tokenResp api.TokenResponse
	json.NewDecoder(loginResp.Body).Decode(&tokenResp)
	accessToken := tokenResp.AccessToken

	req, _ = http.NewRequest(http.MethodGet, srv.URL+"/users/user2/favorites", nil)
	req.Header.Set("Authorization", "Bearer "+accessToken)
	res, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusForbidden {
		t.Errorf("Expected 403 for accessing other user's data, got %d", res.StatusCode)
	}
}
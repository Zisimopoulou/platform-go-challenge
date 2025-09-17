package tests

import (
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Zisimopoulou/platform-go-challenge/internal/api"
	"github.com/Zisimopoulou/platform-go-challenge/internal/core"
	"github.com/Zisimopoulou/platform-go-challenge/internal/data"
)


func TestAuthScenarios(t *testing.T) {
 	if os.Getenv("JWT_SECRET") == "" {
		log.Println("JWT_SECRET not set, using default 'dev-secret' (not safe for production)")
		os.Setenv("JWT_SECRET", "dev-secret")
	}

	store := data.NewInMemoryStore()
	svc := core.NewService(store)
	h := api.NewHandler(svc)

	mux := http.NewServeMux()
	mux.Handle("/", api.AuthMiddleware(h))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := &http.Client{}

	// Test 1: No authorization header
	req, _ := http.NewRequest(http.MethodGet, srv.URL+"/user123/favorites", nil)
	res, err := client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected 401 for no auth header, got %d", res.StatusCode)
	}

	// Test 2: Invalid token format
	req, _ = http.NewRequest(http.MethodGet, srv.URL+"/user123/favorites", nil)
	req.Header.Set("Authorization", "InvalidFormat")
	res, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusUnauthorized {
		t.Errorf("Expected 401 for invalid format, got %d", res.StatusCode)
	}

	// Test 3: Expired token (you'd need to generate one with past expiry)
	// This requires custom token generation with expiry

	// Test 4: User tries to access another user's data
	user1Token, _ := api.GenerateToken("user1")
	req, _ = http.NewRequest(http.MethodGet, srv.URL+"/user2/favorites", nil)
	req.Header.Set("Authorization", "Bearer "+user1Token)
	res, err = client.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	// This should either 404 or 403 depending on your implementation
	if res.StatusCode != http.StatusNotFound && res.StatusCode != http.StatusForbidden {
		t.Errorf("Expected 404/403 for accessing other user's data, got %d", res.StatusCode)
	}
}

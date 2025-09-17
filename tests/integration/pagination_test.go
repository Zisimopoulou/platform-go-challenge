package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Zisimopoulou/platform-go-challenge/internal/core"
	"github.com/Zisimopoulou/platform-go-challenge/internal/data"
	"github.com/Zisimopoulou/platform-go-challenge/internal/models"
	"github.com/Zisimopoulou/platform-go-challenge/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPagination(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret-32-chars-long-for-testing-only")

	store := data.NewInMemoryStore()
	svc := core.NewService(store)

	mux := http.NewServeMux()
	mux.Handle("/users/", http.StripPrefix("/users", api.NewHandler(svc)))
	mux.HandleFunc("/auth/login", api.LoginHandler)

	handler := api.WithMiddleware(mux)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	user := "paginationuser"

	loginReq := map[string]string{"userId": user}
	loginBody, _ := json.Marshal(loginReq)
	loginResp, err := http.Post(srv.URL+"/auth/login", "application/json", bytes.NewReader(loginBody))
	require.NoError(t, err)
	defer loginResp.Body.Close()

	var tokenResp api.TokenResponse
	json.NewDecoder(loginResp.Body).Decode(&tokenResp)
	authHeader := "Bearer " + tokenResp.AccessToken

	client := &http.Client{}

	for i := 1; i <= 15; i++ {
		addReq := map[string]interface{}{
			"type":        "insight",
			"description": fmt.Sprintf("Insight %d", i),
			"payload":     map[string]string{"text": fmt.Sprintf("Content %d", i)},
		}
		b, _ := json.Marshal(addReq)
		req, _ := http.NewRequest(http.MethodPost, srv.URL+"/users/"+user+"/favorites", bytes.NewReader(b))
		req.Header.Set("Authorization", authHeader)
		req.Header.Set("Content-Type", "application/json")
		res, _ := client.Do(req)
		res.Body.Close()
	}

	t.Run("Default pagination", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, srv.URL+"/users/"+user+"/favorites", nil)
		req.Header.Set("Authorization", authHeader)

		res, err := client.Do(req)
		require.NoError(t, err)
		defer res.Body.Close()

		assert.Equal(t, http.StatusOK, res.StatusCode)

		var result models.PaginatedFavorites
		err = json.NewDecoder(res.Body).Decode(&result)
		require.NoError(t, err)

		assert.Len(t, result.Favorites, 15)
		assert.Equal(t, 15, result.TotalCount)
		assert.Equal(t, 50, result.Limit)
		assert.Equal(t, 0, result.Offset)
		assert.False(t, result.HasMore)
	})

	t.Run("Limit 5 items", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, srv.URL+"/users/"+user+"/favorites?limit=5", nil)
		req.Header.Set("Authorization", authHeader)

		res, err := client.Do(req)
		require.NoError(t, err)
		defer res.Body.Close()

		var result models.PaginatedFavorites
		json.NewDecoder(res.Body).Decode(&result)

		assert.Len(t, result.Favorites, 5)
		assert.Equal(t, 15, result.TotalCount)
		assert.Equal(t, 5, result.Limit)
		assert.Equal(t, 0, result.Offset)
		assert.True(t, result.HasMore)
	})

	t.Run("Limit 5 with offset 10", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, srv.URL+"/users/"+user+"/favorites?limit=5&offset=10", nil)
		req.Header.Set("Authorization", authHeader)

		res, err := client.Do(req)
		require.NoError(t, err)
		defer res.Body.Close()

		var result models.PaginatedFavorites
		json.NewDecoder(res.Body).Decode(&result)

		assert.Len(t, result.Favorites, 5)
		assert.Equal(t, 15, result.TotalCount)
		assert.Equal(t, 5, result.Limit)
		assert.Equal(t, 10, result.Offset)
		assert.False(t, result.HasMore)
	})

	t.Run("Invalid pagination parameters", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, srv.URL+"/users/"+user+"/favorites?limit=abc&offset=def", nil)
		req.Header.Set("Authorization", authHeader)

		res, err := client.Do(req)
		require.NoError(t, err)
		defer res.Body.Close()

		var result models.PaginatedFavorites
		json.NewDecoder(res.Body).Decode(&result)

		assert.Len(t, result.Favorites, 15)
		assert.Equal(t, 50, result.Limit)
		assert.Equal(t, 0, result.Offset)
	})
}

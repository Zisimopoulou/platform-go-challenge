package tests

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Zisimopoulou/platform-go-challenge/internal/api"
	"github.com/Zisimopoulou/platform-go-challenge/internal/core"
	"github.com/Zisimopoulou/platform-go-challenge/internal/data"
	"github.com/Zisimopoulou/platform-go-challenge/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFavoritesFlow(t *testing.T) {
	t.Setenv("JWT_SECRET", "test-secret-32-chars-long-for-testing-only")

	store := data.NewInMemoryStore()
	svc := core.NewService(store)

	mux := http.NewServeMux()
	mux.Handle("/users/", http.StripPrefix("/users", api.NewHandler(svc)))
	mux.HandleFunc("/auth/login", api.LoginHandler)
	mux.HandleFunc("/auth/refresh", api.RefreshHandler)

	handler := api.WithMiddleware(mux)
	srv := httptest.NewServer(handler)
	defer srv.Close()

	user := "user123"

	// Login to get token
	loginReq := map[string]string{"userId": user}
	loginBody, _ := json.Marshal(loginReq)
	loginResp, err := http.Post(srv.URL+"/auth/login", "application/json", bytes.NewReader(loginBody))
	require.NoError(t, err)
	defer loginResp.Body.Close()

	require.Equal(t, http.StatusOK, loginResp.StatusCode)

	var tokenResp api.TokenResponse
	err = json.NewDecoder(loginResp.Body).Decode(&tokenResp)
	require.NoError(t, err)
	require.NotEmpty(t, tokenResp.AccessToken)

	authHeader := "Bearer " + tokenResp.AccessToken
	client := &http.Client{}

	// -------------------- ADD FAVORITE --------------------
	t.Log("Adding favorite...")
	addReq := map[string]interface{}{
		"type":        "insight",
		"description": "initial",
		"payload":     map[string]string{"text": "40% of millennials..."},
	}
	b, _ := json.Marshal(addReq)
	reqAdd, _ := http.NewRequest(http.MethodPost, srv.URL+"/users/"+user+"/favorites", bytes.NewReader(b))
	reqAdd.Header.Set("Authorization", authHeader)
	reqAdd.Header.Set("Content-Type", "application/json")

	resAdd, err := client.Do(reqAdd)
	require.NoError(t, err)
	defer resAdd.Body.Close()

	require.Equal(t, http.StatusCreated, resAdd.StatusCode, "Expected 201 Created when adding favorite")

	var created map[string]string
	err = json.NewDecoder(resAdd.Body).Decode(&created)
	require.NoError(t, err)
	require.NotEmpty(t, created["favoriteId"])
	favID := created["favoriteId"]
	t.Logf("Created favorite with ID: %s", favID)

	// -------------------- LIST FAVORITES --------------------
	t.Log("Listing favorites...")
	reqList, _ := http.NewRequest(http.MethodGet, srv.URL+"/users/"+user+"/favorites", nil)
	reqList.Header.Set("Authorization", authHeader)

	resList, err := client.Do(reqList)
	require.NoError(t, err)
	defer resList.Body.Close()

	require.Equal(t, http.StatusOK, resList.StatusCode, "Expected 200 OK when listing favorites")

	// Debug: Read the response body to see what's being returned
	bodyBytes, _ := io.ReadAll(resList.Body)
	t.Logf("Response body: %s", string(bodyBytes))

	// Reset the body for JSON decoding
	resList.Body = io.NopCloser(bytes.NewReader(bodyBytes))

	var paginatedResponse models.PaginatedFavorites
	err = json.NewDecoder(resList.Body).Decode(&paginatedResponse)
	require.NoError(t, err)

	assert.Len(t, paginatedResponse.Favorites, 1, "Expected 1 favorite")
	assert.Equal(t, 1, paginatedResponse.TotalCount, "Expected total count of 1")

	if len(paginatedResponse.Favorites) > 0 {
		asset := paginatedResponse.Favorites[0].Asset
		assert.Equal(t, "initial", asset.Description, "Expected description 'initial'")
	}

	// -------------------- UPDATE FAVORITE --------------------
	t.Log("Updating favorite...")
	updateReq := map[string]string{"description": "updated"}
	updateBody, _ := json.Marshal(updateReq)
	reqUpdate, _ := http.NewRequest(http.MethodPut, srv.URL+"/users/"+user+"/favorites/"+favID, bytes.NewReader(updateBody))
	reqUpdate.Header.Set("Authorization", authHeader)
	reqUpdate.Header.Set("Content-Type", "application/json")

	resUpdate, err := client.Do(reqUpdate)
	require.NoError(t, err)
	defer resUpdate.Body.Close()

	assert.Equal(t, http.StatusNoContent, resUpdate.StatusCode, "Expected 204 No Content when updating favorite")

	// -------------------- VERIFY UPDATE --------------------
	t.Log("Verifying update...")
	reqListUpdated, _ := http.NewRequest(http.MethodGet, srv.URL+"/users/"+user+"/favorites", nil)
	reqListUpdated.Header.Set("Authorization", authHeader)

	resListUpdated, err := client.Do(reqListUpdated)
	require.NoError(t, err)
	defer resListUpdated.Body.Close()

	var updatedResponse models.PaginatedFavorites
	err = json.NewDecoder(resListUpdated.Body).Decode(&updatedResponse)
	require.NoError(t, err)

	if len(updatedResponse.Favorites) > 0 {
		assetUpdated := updatedResponse.Favorites[0].Asset
		assert.Equal(t, "updated", assetUpdated.Description, "Expected updated description 'updated'")
	}

	// -------------------- DELETE FAVORITE --------------------
	t.Log("Deleting favorite...")
	reqDelete, _ := http.NewRequest(http.MethodDelete, srv.URL+"/users/"+user+"/favorites/"+favID, nil)
	reqDelete.Header.Set("Authorization", authHeader)

	resDelete, err := client.Do(reqDelete)
	require.NoError(t, err)
	defer resDelete.Body.Close()

	assert.Equal(t, http.StatusNoContent, resDelete.StatusCode, "Expected 204 No Content when deleting favorite")

	// -------------------- VERIFY EMPTY LIST --------------------
	t.Log("Verifying empty list...")
	reqListEmpty, _ := http.NewRequest(http.MethodGet, srv.URL+"/users/"+user+"/favorites", nil)
	reqListEmpty.Header.Set("Authorization", authHeader)

	resListEmpty, err := client.Do(reqListEmpty)
	require.NoError(t, err)
	defer resListEmpty.Body.Close()

	var emptyResponse models.PaginatedFavorites
	err = json.NewDecoder(resListEmpty.Body).Decode(&emptyResponse)
	require.NoError(t, err)

	assert.Len(t, emptyResponse.Favorites, 0, "Expected 0 favorites after deletion")
	assert.Equal(t, 0, emptyResponse.TotalCount, "Expected total count of 0 after deletion")
}

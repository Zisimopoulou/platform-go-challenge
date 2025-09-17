package tests

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/Zisimopoulou/platform-go-challenge/internal/api"
	"github.com/Zisimopoulou/platform-go-challenge/internal/core"
	"github.com/Zisimopoulou/platform-go-challenge/internal/data"
)

func TestFavoritesFlow(t *testing.T) {
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
	user := "user123"
	token, err := api.GenerateToken(user)
	if err != nil {
		t.Fatal(err)
	}
	authHeader := "Bearer " + token
	client := &http.Client{}

	// -------------------- ADD FAVORITE --------------------
	addReq := map[string]interface{}{
		"type":        "insight",
		"description": "initial",
		"payload":     map[string]string{"text": "40% of millennials..."},
	}
	b, _ := json.Marshal(addReq)
	reqAdd, _ := http.NewRequest(http.MethodPost, srv.URL+"/"+user+"/favorites", bytes.NewReader(b))
	reqAdd.Header.Set("Authorization", authHeader)
	reqAdd.Header.Set("Content-Type", "application/json")
	resAdd, err := client.Do(reqAdd)
	if err != nil {
		t.Fatal(err)
	}
	defer resAdd.Body.Close()

	if resAdd.StatusCode != http.StatusCreated {
		t.Fatalf("expected 201 Created, got %d", resAdd.StatusCode)
	}

	var created map[string]string
	json.NewDecoder(resAdd.Body).Decode(&created)
	favID := created["favoriteId"]

	// -------------------- LIST FAVORITES --------------------
	reqList, _ := http.NewRequest(http.MethodGet, srv.URL+"/"+user+"/favorites", nil)
	reqList.Header.Set("Authorization", authHeader)
	resList, err := client.Do(reqList)
	if err != nil {
		t.Fatal(err)
	}
	defer resList.Body.Close()

	if resList.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK, got %d", resList.StatusCode)
	}

	var list []map[string]interface{}
	json.NewDecoder(resList.Body).Decode(&list)
	if len(list) != 1 {
		t.Fatalf("expected 1 favorite, got %d", len(list))
	}

	asset := list[0]["asset"].(map[string]interface{})
	if asset["description"] != "initial" {
		t.Fatalf("expected description 'initial', got %v", asset["description"])
	}

	// -------------------- UPDATE FAVORITE --------------------
	updateReq := map[string]string{"description": "updated"}
	edb, _ := json.Marshal(updateReq)
	reqUpdate, _ := http.NewRequest(http.MethodPut, srv.URL+"/"+user+"/favorites/"+favID, bytes.NewReader(edb))
	reqUpdate.Header.Set("Authorization", authHeader)
	reqUpdate.Header.Set("Content-Type", "application/json")
	resUpdate, err := client.Do(reqUpdate)
	if err != nil {
		t.Fatal(err)
	}
	defer resUpdate.Body.Close()

	if resUpdate.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204 No Content, got %d", resUpdate.StatusCode)
	}

	// Verify update
	reqListUpdated, _ := http.NewRequest(http.MethodGet, srv.URL+"/"+user+"/favorites", nil)
	reqListUpdated.Header.Set("Authorization", authHeader)
	resListUpdated, err := client.Do(reqListUpdated)
	if err != nil {
		t.Fatal(err)
	}
	defer resListUpdated.Body.Close()

	var listUpdated []map[string]interface{}
	json.NewDecoder(resListUpdated.Body).Decode(&listUpdated)
	assetUpdated := listUpdated[0]["asset"].(map[string]interface{})
	if assetUpdated["description"] != "updated" {
		t.Fatalf("expected updated description 'updated', got %v", assetUpdated["description"])
	}

	// -------------------- DELETE FAVORITE --------------------
	reqDelete, _ := http.NewRequest(http.MethodDelete, srv.URL+"/"+user+"/favorites/"+favID, nil)
	reqDelete.Header.Set("Authorization", authHeader)
	resDelete, err := client.Do(reqDelete)
	if err != nil {
		t.Fatal(err)
	}
	defer resDelete.Body.Close()

	if resDelete.StatusCode != http.StatusNoContent {
		t.Fatalf("expected 204 No Content, got %d", resDelete.StatusCode)
	}

	// -------------------- VERIFY EMPTY LIST --------------------
	reqListEmpty, _ := http.NewRequest(http.MethodGet, srv.URL+"/"+user+"/favorites", nil)
	reqListEmpty.Header.Set("Authorization", authHeader)
	resListEmpty, err := client.Do(reqListEmpty)
	if err != nil {
		t.Fatal(err)
	}
	defer resListEmpty.Body.Close()

	var listEmpty []map[string]interface{}
	json.NewDecoder(resListEmpty.Body).Decode(&listEmpty)
	if len(listEmpty) != 0 {
		t.Fatalf("expected 0 favorites, got %d", len(listEmpty))
	}
}

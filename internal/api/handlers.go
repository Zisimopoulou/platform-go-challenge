package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"strconv"

	"github.com/Zisimopoulou/platform-go-challenge/internal/core"
	"github.com/Zisimopoulou/platform-go-challenge/internal/models"
)

type Handler struct {
	svc *core.Service
}

func NewHandler(svc *core.Service) *http.ServeMux {
	h := &Handler{svc: svc}
	mux := http.NewServeMux()

	baseHandler := http.HandlerFunc(h.handle)

	protectedHandler := AuthMiddleware(baseHandler)

	mux.Handle("/", protectedHandler)
	return mux
}

func (h *Handler) handle(w http.ResponseWriter, r *http.Request) {
	userID, ok := FromContextUserID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	path := strings.Trim(r.URL.Path, "/")
	parts := strings.Split(path, "/")
	if len(parts) < 1 {
		writeError(w, http.StatusNotFound, "invalid path")
		return
	}

	requestedUser := parts[0]
	if requestedUser != userID {
		writeError(w, http.StatusForbidden, "access denied")
		return
	}

	switch {
	case len(parts) == 2 && parts[1] == "favorites" && r.Method == http.MethodGet:
		h.handleListFavorites(w, r, userID)

	case len(parts) == 2 && parts[1] == "favorites" && r.Method == http.MethodPost:
		h.handleAddFavorite(w, r, userID)

	case len(parts) == 3 && parts[1] == "favorites" && r.Method == http.MethodDelete:
		favID := parts[2]
		h.handleDeleteFavorite(w, r, userID, favID)

	case len(parts) == 3 && parts[1] == "favorites" && r.Method == http.MethodPut:
		favID := parts[2]
		h.handleUpdateFavorite(w, r, userID, favID)

	default:
		writeError(w, http.StatusNotFound, "not found")
	}
}

 func (h *Handler) handleAddFavorite(w http.ResponseWriter, r *http.Request, userID string) {
	var asset models.RawAsset
	if err := json.NewDecoder(r.Body).Decode(&asset); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

 	if err := validateAsset(&asset); err != nil {
		validationErrorResponse(w, err)
		return
	}

	id, err := h.svc.AddFavorite(userID, asset)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]string{"favoriteId": id})
}

 func (h *Handler) handleUpdateFavorite(w http.ResponseWriter, r *http.Request, userID, favID string) {
	var body struct {
		Description string `json:"description" validate:"required,max=500"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		writeError(w, http.StatusBadRequest, "invalid body")
		return
	}

 	if err := validateStruct(body); err != nil {
		validationErrorResponse(w, err)
		return
	}

	if err := h.svc.UpdateDescription(userID, favID, body.Description); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) handleDeleteFavorite(w http.ResponseWriter, r *http.Request, userID, favID string) {
	if err := h.svc.DeleteFavorite(userID, favID); err != nil {
		writeError(w, http.StatusNotFound, err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func writeError(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func (h *Handler) handleListFavorites(w http.ResponseWriter, r *http.Request, userID string) {
	limit, offset := getPaginationParams(r)
	
	favs, err := h.svc.ListFavorites(userID, limit, offset)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, favs)
}

func getPaginationParams(r *http.Request) (int, int) {
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")
	
	limit := 50
	offset := 0
	
	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
			if limit > 100 {
				limit = 100
			}
		}
	}
	
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}
	
	return limit, offset
}

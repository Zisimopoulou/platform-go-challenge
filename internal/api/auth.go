package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type ctxKeyUserID struct{}

type TokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken"`
}

func GenerateTokens(userID string) (string, string, int64, error) {
	secret := getJWTSecret()
	expiresAt := time.Now().Add(time.Hour * 1).Unix()
	accessClaims := jwt.MapClaims{"sub": userID, "exp": expiresAt, "type": "access"}
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessString, err := accessToken.SignedString(secret)

	if err != nil {
		return "", "", 0, err
	}

	refreshExpiresAt := time.Now().Add(time.Hour * 24 * 7).Unix()
	refreshClaims := jwt.MapClaims{"sub": userID, "exp": refreshExpiresAt, "type": "refresh"}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshString, err := refreshToken.SignedString(secret)

	if err != nil {
		return "", "", 0, err
	}

	return accessString, refreshString, expiresAt, nil
}

func getJWTSecret() []byte {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = os.Getenv("APP_JWT_SECRET")
	}
	if secret == "" {
		panic("JWT_SECRET or APP_JWT_SECRET environment variable not set")
	}
	return []byte(secret)
}

func FromContextUserID(ctx context.Context) (string, bool) {
	v := ctx.Value(ctxKeyUserID{})
	if v == nil {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			writeError(w, http.StatusUnauthorized, "missing authorization header")
			return
		}
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			writeError(w, http.StatusUnauthorized, "invalid authorization header")
			return
		}
		tokenStr := parts[1]
		secret := getJWTSecret()
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, errors.New("unexpected signing method")
			}
			return secret, nil
		})
		if err != nil || !token.Valid {
			writeError(w, http.StatusUnauthorized, "invalid token")
			return
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			writeError(w, http.StatusUnauthorized, "invalid token claims")
			return
		}
		if claims["type"] != "access" {
			writeError(w, http.StatusUnauthorized, "invalid token type")
			return
		}
		sub, ok := claims["sub"].(string)
		if !ok || sub == "" {
			writeError(w, http.StatusUnauthorized, "invalid token subject")
			return
		}
		r = r.WithContext(context.WithValue(r.Context(), ctxKeyUserID{}, sub))
		next.ServeHTTP(w, r)
	})
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var body struct {
		UserID string `json:"userId"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.UserID == "" {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	accessToken, refreshToken, expiresAt, err := GenerateTokens(body.UserID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresAt,
	})
}

func RefreshHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var body RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.RefreshToken == "" {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	secret := getJWTSecret()
	token, err := jwt.Parse(body.RefreshToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("unexpected signing method")
		}
		return secret, nil
	})
	if err != nil || !token.Valid {
		writeError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || claims["type"] != "refresh" {
		writeError(w, http.StatusUnauthorized, "invalid refresh token")
		return
	}
	userID, ok := claims["sub"].(string)
	if !ok || userID == "" {
		writeError(w, http.StatusUnauthorized, "invalid token subject")
		return
	}
	accessToken, refreshToken, expiresAt, err := GenerateTokens(userID)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err.Error())
		return
	}
	writeJSON(w, http.StatusOK, TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    expiresAt,
	})
}
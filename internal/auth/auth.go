package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"smsotp/internal/database"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var jwtSecret = []byte("smsotp-secret-change-me-in-production")

type contextKey string

const UserIDKey contextKey = "userID"

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type ChangePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

type LoginResponse struct {
	Token           string `json:"token"`
	Username        string `json:"username"`
	DefaultPassword bool   `json:"default_password"`
}

type AdminInfo struct {
	ID              int    `json:"id"`
	Username        string `json:"username"`
	DefaultPassword bool   `json:"default_password"`
}

func EnsureAdmin() {
	var count int
	database.DB.QueryRow("SELECT COUNT(*) FROM admin").Scan(&count)
	if count == 0 {
		hash, _ := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		database.DB.Exec("INSERT INTO admin (username, password_hash) VALUES (?, ?)", "admin", string(hash))
	}
}

func isDefaultPassword(hash string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte("admin123")) == nil
}

func HandleLogin(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}

	var id int
	var hash string
	err := database.DB.QueryRow("SELECT id, password_hash FROM admin WHERE username = ?", req.Username).Scan(&id, &hash)
	if err != nil {
		http.Error(w, `{"error":"Invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		http.Error(w, `{"error":"Invalid credentials"}`, http.StatusUnauthorized)
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  id,
		"username": req.Username,
		"exp":      time.Now().Add(24 * time.Hour).Unix(),
	})
	tokenStr, err := token.SignedString(jwtSecret)
	if err != nil {
		http.Error(w, `{"error":"Failed to generate token"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{
		Token:           tokenStr,
		Username:        req.Username,
		DefaultPassword: isDefaultPassword(hash),
	})
}

func HandleMe(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(float64)
	var username, hash string
	err := database.DB.QueryRow("SELECT username, password_hash FROM admin WHERE id = ?", int(userID)).Scan(&username, &hash)
	if err != nil {
		http.Error(w, `{"error":"User not found"}`, http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(AdminInfo{
		ID:              int(userID),
		Username:        username,
		DefaultPassword: isDefaultPassword(hash),
	})
}

func HandleChangePassword(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(UserIDKey).(float64)

	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid request"}`, http.StatusBadRequest)
		return
	}

	var hash string
	err := database.DB.QueryRow("SELECT password_hash FROM admin WHERE id = ?", int(userID)).Scan(&hash)
	if err != nil {
		http.Error(w, `{"error":"User not found"}`, http.StatusNotFound)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.CurrentPassword)); err != nil {
		http.Error(w, `{"error":"Current password is incorrect"}`, http.StatusBadRequest)
		return
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, `{"error":"Failed to hash password"}`, http.StatusInternalServerError)
		return
	}

	database.DB.Exec("UPDATE admin SET password_hash = ? WHERE id = ?", string(newHash), int(userID))

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "Password changed successfully"})
}

func JWTMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})
		if err != nil || !token.Valid {
			http.Error(w, `{"error":"Unauthorized"}`, http.StatusUnauthorized)
			return
		}

		claims := token.Claims.(jwt.MapClaims)
		ctx := context.WithValue(r.Context(), UserIDKey, claims["user_id"])
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

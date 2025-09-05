package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
	"io"
	"os"
	"path/filepath"
	"fmt"

	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET"))

type contextKey string
const userContextKey = contextKey("user")

// signupHandler creates a new user
func signupHandler(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(creds.Password), 8)
	if err != nil {
		http.Error(w, "Failed to hash password", http.StatusInternalServerError)
		return
	}

	_, err = db.Exec("INSERT INTO users (email, password) VALUES ($1, $2)", creds.Email, string(hashedPassword))
	if err != nil {
		http.Error(w, "User already exists", http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "User created successfully"})
}


// loginHandler authenticates a user and returns a JWT
func loginHandler(w http.ResponseWriter, r *http.Request) {
    var creds Credentials
    if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    var user User
    row := db.QueryRow("SELECT id, email, password, avatar_url FROM users WHERE email = $1", creds.Email)
    if err := row.Scan(&user.ID, &user.Email, &user.Password, &user.AvatarURL); err != nil {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

    if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

    expirationTime := time.Now().Add(24 * time.Hour)
    claims := &Claims{
        UserID: user.ID,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(expirationTime),
        },
    }

    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    tokenString, err := token.SignedString(jwtKey)
    if err != nil {
        http.Error(w, "Failed to create token", http.StatusInternalServerError)
        return
    }

	w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}


// getHomeContentHandler serves dynamic content for the home page
func getHomeContentHandler(w http.ResponseWriter, r *http.Request) {
	media := []Media{
		{ID: 1, Title: "Exploring the Alps", Type: "image", URL: "https://placehold.co/600x400/000000/FFFFFF?text=Alps"},
		{ID: 2, Title: "Ocean Documentary", Type: "video", URL: "https://placehold.co/600x400/0000FF/FFFFFF?text=Video"},
		{ID: 3, Title: "Annual Report 2024", Type: "document", URL: "#"},
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(media)
}


// dashboardHandler serves user-specific data
func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(userContextKey).(User)
    if !ok {
        http.Error(w, "User not found in context", http.StatusInternalServerError)
        return
    }
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

// uploadAvatarHandler handles user avatar uploads
func uploadAvatarHandler(w http.ResponseWriter, r *http.Request) {
    user, ok := r.Context().Value(userContextKey).(User)
    if !ok {
        http.Error(w, "User not found in context", http.StatusInternalServerError)
        return
    }

    r.ParseMultipartForm(10 << 20) // 10 MB
    file, handler, err := r.FormFile("avatar")
    if err != nil {
        http.Error(w, "Error retrieving the file", http.StatusBadRequest)
        return
    }
    defer file.Close()

    // Create uploads directory if it doesn't exist
    os.MkdirAll("./uploads", os.ModePerm)

    // Create a unique filename
    ext := filepath.Ext(handler.Filename)
    fileName := fmt.Sprintf("%d%s", user.ID, ext)
    filePath := filepath.Join("./uploads", fileName)
    
    dst, err := os.Create(filePath)
    if err != nil {
        http.Error(w, "Unable to create the file", http.StatusInternalServerError)
        return
    }
    defer dst.Close()

    _, err = io.Copy(dst, file)
    if err != nil {
        http.Error(w, "Unable to save the file", http.StatusInternalServerError)
        return
    }

    avatarURL := fmt.Sprintf("/uploads/%s", fileName)
    _, err = db.Exec("UPDATE users SET avatar_url = $1 WHERE id = $2", avatarURL, user.ID)
    if err != nil {
        http.Error(w, "Failed to update user avatar URL", http.StatusInternalServerError)
        return
    }

	w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"avatarUrl": avatarURL})
}


// jwtMiddleware protects routes that require authentication
func jwtMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		tokenStr := r.Header.Get("Authorization")
		if tokenStr == "" {
			http.Error(w, "Missing auth token", http.StatusUnauthorized)
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}
		
		var user User
		row := db.QueryRow("SELECT id, email, avatar_url FROM users WHERE id = $1", claims.UserID)
		if err := row.Scan(&user.ID, &user.Email, &user.AvatarURL); err != nil {
			http.Error(w, "User not found", http.StatusUnauthorized)
			return
		}
		
		// Add user to context
		ctx := context.WithValue(r.Context(), userContextKey, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

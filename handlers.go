package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v4"
	"golang.org/x/crypto/bcrypt"
)

var jwtKey = []byte(os.Getenv("JWT_SECRET"))

func signupHandler(c *gin.Context) {
	var user User
	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 8)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to hash password"})
		return
	}

	_, err = db.Exec(
		"INSERT INTO users (prenom, nom, telephone, email, password) VALUES ($1, $2, $3, $4, $5)",
		user.Prenom, user.Nom, user.Telephone, user.Email, string(hashedPassword))
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"message": "User created successfully"})
}

func loginHandler(c *gin.Context) {
	var creds Credentials
	if err := c.BindJSON(&creds); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	var user User
	row := db.QueryRow("SELECT id, email, password, avatar_url FROM users WHERE email = $1", creds.Email)
	if err := row.Scan(&user.ID, &user.Email, &user.Password, &user.AvatarURL); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(creds.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid credentials"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func getHomeContentHandler(c *gin.Context) {

	media := []Post{
		{ID: 1, Title: "Exploring the Alps with Gin", Medias: []Media{
			{Type: "image", Url: "https://placehold.co/600x400/000000/FFFFFF?text=Alps"},
		}},
		{ID: 2, Title: "Ocean Documentary", Medias: []Media{
			{Type: "video", Url: "https://placehold.co/600x400/0000FF/FFFFFF?text=Video"},
		},},
			{ID: 3, Title: "Annual Report 2024", Medias: []Media{
				{Type: "document", Url: "#"},
			},
		},
	}

	c.JSON(http.StatusOK, media)
}

func dashboardHandler(c *gin.Context) {
	userCtx, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found in context"})
		return
	}

	_user, ok := userCtx.(User)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user type in context"})
		return
	}

	user, err := getUserById(_user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid user type in context"})
		return
	} 

	c.JSON(http.StatusOK, user)
}

func uploadAvatarHandler(c *gin.Context) {
	userCtx, exists := c.Get("user")
	if !exists {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found in context"})
		return
	}
	user := userCtx.(User) // We can be reasonably sure of the type here

	file, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Error retrieving the file"})
		return
	}

	// Create uploads directory if it doesn't exist
	os.MkdirAll("./uploads", os.ModePerm)

	// Create a unique filename
	ext := filepath.Ext(file.Filename)
	fileName := fmt.Sprintf("%d%s", user.ID, ext)
	filePath := filepath.Join("./uploads", fileName)

	// Save the file
	if err := c.SaveUploadedFile(file, filePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Unable to save the file"})
		return
	}

	avatarURL := fmt.Sprintf("/uploads/%s", fileName)
	_, err = db.Exec("UPDATE users SET avatar_url = $1 WHERE id = $2", avatarURL, user.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user avatar URL"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"avatarUrl": avatarURL})
}

// jwtMiddleware protects routes that require authentication
func jwtMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := c.GetHeader("Authorization")
		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Missing auth token"})
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			return
		}

		var user User
		row := db.QueryRow("SELECT id, email, avatar_url FROM users WHERE id = $1", claims.UserID)
		if err := row.Scan(&user.ID, &user.Email, &user.AvatarURL); err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "User not found"})
			return
		}

		// Add user to context
		c.Set("user", user)
		c.Next()
	}
}


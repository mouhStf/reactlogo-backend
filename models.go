package main

import "github.com/golang-jwt/jwt/v4"

// Credentials struct for login/signup
type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// User struct for database
type User struct {
	ID        int     `json:"id"`
	Email     string  `json:"email"`
	Password  string  `json:"-"` // Omit from JSON responses
	AvatarURL *string `json:"avatarUrl"`
}

// Claims struct for JWT
type Claims struct {
	UserID int `json:"userId"`
	jwt.RegisteredClaims
}

// Media struct for home page content
type Media struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Type  string `json:"type"` // "image", "video", "document"
	URL   string `json:"url"`
}

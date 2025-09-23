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
	Prenom    string  `json:"prenom"`
	Nom       string  `json:"nom"`
	Telephone string  `json:"telephone"`
	Email     string  `json:"email"`
	Password  string  `json:"password"` // Omit from JSON responses
	AvatarURL *string `json:"avatarUrl"`
}

// Claims struct for JWT
type Claims struct {
	UserID int `json:"userId"`
	jwt.RegisteredClaims
}

type Media struct {
	Type string `json:"type"`
	Url  string `json:"url"`
}

type Post struct {
	ID      int     `json:"id"`
	Title   string  `json:"title"`
	Content string  `json:"content"`
	Medias  []Media `json:"medias"`
}

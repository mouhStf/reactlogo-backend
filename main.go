package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

var db *sql.DB

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	initDB()
	defer db.Close()

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{
			"http://localhost:3000",
			"http://localhost:5173",
			"https://reactlogo-frontend.vercel.app",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// Serve static files
	router.Static("/uploads", "./uploads")
	router.Static("/public", "./public")

	api := router.Group("/api")
	{
		api.POST("/signup", signupHandler)
		api.POST("/login", loginHandler)
		api.GET("/home", getHomeContentHandler)

		// Protected routes group
		protected := api.Group("/")
		protected.Use(jwtMiddleware()) // Apply JWT middleware
		{
			protected.GET("/dashboard", dashboardHandler)
			protected.POST("/upload-avatar", uploadAvatarHandler)
		}
	}

	fmt.Println("Server starting on port 8080...")

	if err := router.Run(":8080"); err != nil {
		log.Fatal("Failed to run server:", err)
	}
}

func initDB() {
	var err error
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	db, err = sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	// Create users table if it doesn't exist
	createTableQuery := `
    CREATE TABLE IF NOT EXISTS users (
        id SERIAL PRIMARY KEY,
	      prenom TEXT,
	      nom TEXT,
	      telephone TEXT,
        email TEXT NOT NULL UNIQUE,
        password TEXT NOT NULL,
        avatar_url TEXT
    );`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatal("Failed to create users table:", err)
	}
	fmt.Println("Database connected and table initialized.")
}

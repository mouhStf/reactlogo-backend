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
	// Load environment variables from .env file
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// Initialize database connection
	initDB()
	defer db.Close()

	// Create a new Gin router. Default() includes logger and recovery middleware.
	router := gin.Default()

	// CORS configuration
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

	// Serve static files (avatars)
	router.Static("/uploads", "./uploads")

	// API routes group
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
	// Start the server
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

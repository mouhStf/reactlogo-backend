package main

import (
	"fmt"
	"log"
	"os"
	"time"
	"strings"
	"net"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)


func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	host := strings.TrimSpace(os.Getenv("HOST"))
	if host == "" {
		host = "::" // unspecified IPv6
	}
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "8080"
	}

	initDB()
	defer db.Close()

	router := gin.Default()

	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{
			"http://localhost:3000",
			"http://localhost:5173",
			"https://djolof-shop.vercel.app",
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
		api.GET("/blog", blogHandler)
		api.GET("/article/:id", getBlogPost)
		api.GET("/article/side", getBlogPostSide)

		// Protected routes group
		protected := api.Group("/")
		protected.Use(jwtMiddleware()) // Apply JWT middleware
		{
			protected.GET("/dashboard", dashboardHandler)
			protected.POST("/upload-avatar", uploadAvatarHandler)
		}
	}

	fmt.Println("Server starting on port 8080...")

	addr := net.JoinHostPort(host, port)
	if err := router.Run(addr); err != nil {
		log.Fatal("Failed to run server:", err)
	}
}

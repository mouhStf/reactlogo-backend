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

type BlogRequestInfo struct {
	Page int `form:"page"`
	Category int `form:"category"`
	Tags []int `form:"tag"`
}

func blogHandler(c *gin.Context) {
	var info BlogRequestInfo
	err := c.ShouldBind(&info)
	if err != nil {
		info.Page = 1
	}
	if info.Page <= 0 {
		info.Page = 1
	}

	articles, s, err := getArticles(info.Category, info.Tags, (info.Page-1)*12, 12)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, gin.H{"articles": articles, "pages": (s / 12)+1})
}

func getBlogPost(c *gin.Context) {
	id := c.Param("id")
	article, error := getBlogPostData(id)
	if error != nil {
		c.JSON(404, gin.H{"error": error.Error()})
	}
	c.JSON(http.StatusOK, article)
}

func getBlogPostSide(c *gin.Context) {
	side, err := getBlogPostSideData()
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, side)
}

func getSearchHeadDatas(c *gin.Context) {
	cs, err := getCategories()
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
	}
	ts, err := getTags()
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
	}

	data := map[string]interface{}{
		"categories":  cs,
		"tags": ts,
	}
	c.JSON(http.StatusOK, data)
}

func getCategoriesHandle(c *gin.Context) {
	cs, err := getCategories()
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, cs)
}

func getTagsHandle(c *gin.Context) {
	ts, err := getTags()
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, ts)
}

type ArticleSearchRequestInfo struct {
	Term string `form:"term"`
	Category int `form:"category"`
	Tags []int `form:"tag"`
	Page int `form:"page"`
}

func articleSearchHandler(c *gin.Context) {
	var info ArticleSearchRequestInfo
	err := c.ShouldBind(&info)
	if err != nil {
		info.Term = "0"
	}
	if info.Page <= 0 {
		info.Page = 1
	}
	ars, nRows, err := searchArticle(info.Term, info.Category, info.Tags, (info.Page-1)*12, 12)
	if err != nil {
		c.JSON(404, gin.H{"error": err.Error()})
	}
	c.JSON(http.StatusOK, gin.H{"articles": ars, "pages": (nRows / 12)+1} )
}

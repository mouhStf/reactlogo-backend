package main

import "github.com/golang-jwt/jwt/v4"

// Credentials struct for login/signup
type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
// Claims struct for JWT
type Claims struct {
	UserID int `json:"userId"`
	jwt.RegisteredClaims
}

// ---------- Users & Authors ----------
type Author struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Title     string `json:"title"`
	Summary   string `json:"summary"`
}

type User struct {
	ID        int     `json:"id"`
	Prenom    string `json:"prenom"`
	Nom       string `json:"nom"`
	Telephone string `json:"telephone"`
	Email     string  `json:"email"`
	Password  string  `json:"-"`        // never expose in JSON
	AvatarURL string `json:"avatarUrl"`
	AuthorID  int    `json:"authorId"`
	DoesLogin bool    `json:"doesLogin"`
}


// ---------- Articles ----------
type Markup struct {
	Element string      `json:"element"`
	Classe  string      `json:"classe"`
	Data    interface{} `json:"data"` // can be string or []map[string]string
}

type Article struct {
	ID       int      `json:"id"`
	AuthorID int      `json:"authorId"`
	Title    string   `json:"title"`
	Category string   `json:"category"`
	CategoryID int   `json:"categoryId"`
	Image    *string  `json:"image"`
	Date     string   `json:"date"`
	Summary  *string  `json:"summary"`
	Content  []Markup `json:"content"`
}

type Tag struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ArticleCategory struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Comment struct {
	ID        int     `json:"id"`
	UserID    int     `json:"userId"`
	Date      string  `json:"date"`
	Comment   string  `json:"comment"`
}

// ---------- Products & Collections ----------
type Collection struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Description *string `json:"description"`
}

type Product struct {
	SKU         string            `json:"sku"`
	CollectionID *int             `json:"collectionId"`
	Name        string            `json:"name"`
	Note        *float32          `json:"note"`
	Price       int               `json:"price"`
	Description *string           `json:"description"`
	Colors      []string          `json:"colors"`
	Images      map[string][]string `json:"images"` // from JSONB
	Sizes       []string          `json:"sizes"`
	Quantity    int               `json:"quantity"`
	// categories come from product_category_links
	Categories []int `json:"categories"`
}

type ProductCategory struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// ---------- Carts ----------
type Cart struct {
	ID        int               `json:"id"`
	UserID    int               `json:"userId"`
	Content   map[string]int    `json:"content"` // product_sku -> quantity
	CreatedAt string            `json:"createdAt"`
	State     int               `json:"state"`
	Viewed    bool              `json:"viewed"`
}


// ------ Utilities

type UserComment struct {
	User    User    `json:"user"`
	Comment Comment `json:"comment"`
}

type BlogPost struct {
	Article  Article         `json:"article"`
	Category ArticleCategory `json:"category"`
	Tags     []Tag           `json:"tags"`
	Author   Author          `json:"author"`
	User     User            `json:"user"`
	N        int             `json:"n"`
	Comments []UserComment   `json:"comments"`
	Next     Article         `json:"next"`
	Previous Article         `json:"previous"`
	Sims     []Article       `json:"sims"`
}

type BlogPostSide struct {
	Categories []ArticleCategory `json:"categories"`
	Tags       []Tag             `json:"tags"`
	Recents    []Article         `json:"recents"`
}

type CategoriesTags struct {
	Categories []ArticleCategory `json:"categories"`
	Tags       []Tag             `json:"tags"`
}

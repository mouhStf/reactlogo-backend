package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"encoding/json"
	"time"
	"strings"
)

var db *sql.DB
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

	createDatabase()

	fmt.Println("Database connected and table initialized.")
}

func createDatabase() {
createTableQuery := `
CREATE TABLE IF NOT EXISTS authors (
id SERIAL PRIMARY KEY,
name TEXT NOT NULL,
title TEXT,
summary TEXT
);
CREATE TABLE IF NOT EXISTS users (
id SERIAL PRIMARY KEY,
prenom TEXT,
nom TEXT,
telephone TEXT,
email TEXT NOT NULL UNIQUE,
password TEXT NOT NULL,
avatar_url TEXT,
author_id INTEGER REFERENCES authors(id),
does_login BOOLEAN DEFAULT false
);
CREATE TABLE IF NOT EXISTS article_categories (
id SERIAL PRIMARY KEY,
name TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS article_tags (
id SERIAL PRIMARY KEY,
name TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS articles (
id SERIAL PRIMARY KEY,
author_id INTEGER REFERENCES authors(id),
category_id INTEGER REFERENCES article_categories(id),
title TEXT NOT NULL,
image TEXT,
date DATE,
summary TEXT,
content JSONB
);
CREATE TABLE IF NOT EXISTS article_tag_links (
article_id INTEGER REFERENCES articles(id) ON DELETE CASCADE,
tag_id INTEGER REFERENCES article_tags(id) ON DELETE CASCADE,
PRIMARY KEY (article_id, tag_id)
);
CREATE TABLE IF NOT EXISTS comments (
id SERIAL PRIMARY KEY,
user_id INTEGER REFERENCES users(id),
article_id INTEGER REFERENCES articles(id),
date TIMESTAMP DEFAULT now(),
comment TEXT
);
CREATE TABLE IF NOT EXISTS product_categories (
id SERIAL PRIMARY KEY,
name TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS collections (
id SERIAL PRIMARY KEY,
name TEXT,
description TEXT
);
CREATE TABLE IF NOT EXISTS products (
SKU TEXT PRIMARY KEY,
collection_id INTEGER REFERENCES collections(id),
name TEXT,
note DECIMAL(1,1),
price INTEGER,
description TEXT,
colors TEXT[],
images JSONB,
sizes TEXT[],
quantity INTEGER
);
CREATE TABLE IF NOT EXISTS carts (
id SERIAL PRIMARY KEY,
user_id INTEGER REFERENCES users(id),
content JSONB,
created_at TIMESTAMP DEFAULT now(),
state INTEGER,
viewed BOOLEAN DEFAULT false
);
CREATE TABLE IF NOT EXISTS product_category_links (
product_sku TEXT REFERENCES products(SKU) ON DELETE CASCADE,
category_id INTEGER REFERENCES product_categories(id) ON DELETE CASCADE,
PRIMARY KEY (product_sku, category_id)
);
` 
// images JSONB example: {"red": ["1.jpg", "2.jpg"], "green": []}
// content JSONB example: {"12743XF": 100, "DF234H": 0}

_, err := db.Exec(createTableQuery)
if err != nil {
log.Fatal("Failed to create users table:", err)
}
}

func getUserById(id int) (*User, error) {
var user User
row := db.QueryRow("SELECT id, prenom, nom, telephone, email, avatar_url FROM users WHERE id = $1", id)
if err := row.Scan(&user.ID, &user.Prenom, &user.Nom, &user.Telephone, &user.Email, &user.AvatarURL); err != nil {
	return nil, err
}

	return &user, nil
}

func formatFrenchDate(dt string) string {

	t, err := time.Parse("2006-01-02T15:04:05Z", dt)
	if err != nil {
		return dt
	}
	frenchMonths := map[string]string{
		"Jan": "janv.",
		"Feb": "févr.",
		"Mar": "mars",
		"Apr": "avr.",
		"May": "mai",
		"Jun": "juin",
		"Jul": "juil.",
		"Aug": "août",
		"Sep": "sept.",
		"Oct": "oct.",
		"Nov": "nov.",
		"Dec": "déc.",
	}

	formatted := t.Format("02 Jan 2006")
	for eng, fr := range frenchMonths {
		formatted = strings.Replace(formatted, eng, fr, 1)
	}
	return formatted
}

func getArticles() ([]Article, error) {
	rows, err := db.Query(`
		SELECT a.id, a.title, ac."name", a.image, a.date, a.summary
		FROM articles a
		left join article_categories ac  ON a.category_id = ac.id
		group by a.id, ac."name" 
		ORDER BY a.date DESC
		`)
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}
	defer rows.Close()

	var articles []Article
	for rows.Next() {
		var a Article

		if err := rows.Scan(
			&a.ID,
			&a.Title,
			&a.Category,
			&a.Image,
			&a.Date,
			&a.Summary,
		); err != nil {
			return nil, fmt.Errorf("scan failed: %v", err)
		}
		a.Date = formatFrenchDate(a.Date)

		articles = append(articles, a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %v", err)
	}

	return articles, nil
}


func getBlogPostData(id string) (*BlogPost, error) {
	var b BlogPost

	var contentJSON []byte
	err := db.QueryRow( `
		select ar.id, ar.title, ar.image, ar."date", ar.summary, ar."content", ac.id, ac."name", au.id, au."name", au.title, au.summary, u.id, u.avatar_url
		from articles ar
		join article_categories ac on ar.category_id = ac.id
		join authors au on ar.author_id = au.id
		join users u on u.author_id = au.id
		where ar.id = $1
	`, id).Scan(&b.Article.ID, &b.Article.Title, &b.Article.Image, &b.Article.Date, &b.Article.Summary, &contentJSON, &b.Category.ID, &b.Category.Name, &b.Author.ID, &b.Author.Name, &b.Author.Title, &b.Author.Summary, &b.User.ID, &b.User.AvatarURL)
	b.Article.Date = formatFrenchDate(b.Article.Date)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("Article with id %d not found", id)
	}

	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}

	if err := json.Unmarshal(contentJSON, &b.Article.Content); err != nil {
		return nil, fmt.Errorf("unmarshal content failed: %v", err)
	}

	tagRows, err := db.Query(`
		select t.id, t."name" 
		from article_tag_links atl
		join article_tags t on atl.tag_id = t.id
		where atl.article_id = $1
		`, id)
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}
	defer tagRows.Close()

	for tagRows.Next() {
		var t Tag
		if err := tagRows.Scan(&t.ID, &t.Name); err != nil {
			return nil, fmt.Errorf("scan failed: %v", err)
		}

		b.Tags = append(b.Tags, t)
	}

	cRows, err := db.Query(`
		select c.id, c."date", c."comment", u.id, u.prenom, u.nom, u.avatar_url 
		from "comments" c
		join users u on c.user_id = u.id
		where c.article_id = $1
		`, id)

	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}
	defer cRows.Close()

	for cRows.Next() {
		var c UserComment
		if err := cRows.Scan(
			&c.Comment.ID, &c.Comment.Date, &c.Comment.Comment,
			&c.User.ID, &c.User.Prenom, &c.User.Nom, &c.User.AvatarURL); err != nil {
			return nil, fmt.Errorf("scan failed: %v", err)
		}
		c.Comment.Date = formatFrenchDate(c.Comment.Date)

		b.Comments = append(b.Comments, c)
	}

	db.QueryRow(`
		select count(c) 
		from "comments" c
		join users u on c.user_id = u.id
		where c.article_id = $1`, id).Scan(&b.N)

	db.QueryRow(`
		select a.id, a.title 
		from articles a
		where a.id < $1
		order by id desc
		limit 1
		`, id).Scan(&b.Previous.ID, &b.Previous.Title)

	db.QueryRow(`
		select a.id, a.title 
		from articles a
		where a.id > $1
		order by id asc
		limit 1
		`, id).Scan(&b.Next.ID, &b.Next.Title)


	var sm []Article
	sRows, err := db.Query(`
		WITH target_article AS (
		SELECT id, title, summary, content , date
		FROM articles
		WHERE id = $1
		),
		tag_overlap AS (
		SELECT a.id AS article_id,
		COUNT(*) AS shared_tags
		FROM article_tag_links atl
		JOIN article_tag_links target_atl 
		ON atl.tag_id = target_atl.tag_id
		JOIN articles a ON a.id = atl.article_id
		WHERE target_atl.article_id = $1
		AND a.id != $1
		GROUP BY a.id
		),
		text_similarity AS (
		SELECT a.id AS article_id,
		ts_rank_cd(
		to_tsvector('french', a.title || ' ' || a.summary || ' ' || a.content::text),
		plainto_tsquery('french', t.content::text)
		) AS text_rank
		FROM articles a
		CROSS JOIN target_article t
		WHERE a.id != $1
		)
		SELECT a.id, a.image, a.date, a.title, a.summary, ac.name,
		COALESCE(tag_overlap.shared_tags, 0) AS shared_tags,
		COALESCE(text_similarity.text_rank, 0) AS text_rank,
		(COALESCE(tag_overlap.shared_tags, 0) * 2 + COALESCE(text_similarity.text_rank, 0)) AS similarity_score
		FROM articles a
		join article_categories ac on a.category_id = ac.id
		LEFT JOIN tag_overlap ON a.id = tag_overlap.article_id
		LEFT JOIN text_similarity ON a.id = text_similarity.article_id
		WHERE a.id != $1
		ORDER BY similarity_score DESC, a.date DESC
		LIMIT 3;
		`, id)

	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}
	defer sRows.Close()

	for sRows.Next() {
		var a Article
		var tagsShared, textRank, simScore int // Just here for now
		if err := sRows.Scan(
			&a.ID, &a.Image, &a.Date, &a.Title, &a.Summary, &a.Category, &tagsShared, &textRank, &simScore); err != nil {
			return nil, fmt.Errorf("scan failed: %v", err)
		}
		a.Date = formatFrenchDate(a.Date)

		sm = append(sm, a)
	}
	b.Sims = sm

	return &b, nil
}

func getBlogPostSideData() (*BlogPostSide, error){
	var b BlogPostSide


	rows, err := db.Query("select c.id, c.name from article_categories c")
	if err != nil {
		return nil, fmt.Errorf("Error: Could not get categories %v", err)
	}
	for rows.Next() {
		var c ArticleCategory
		rows.Scan(&c.ID, &c.Name)
		b.Categories = append(b.Categories, c)
	}

	rows, err = db.Query("select t.id, t.name from article_tags t")
	if err != nil {
		return nil, fmt.Errorf("Error: Could not get tags %v", err)
	}
	for rows.Next() {
		var t Tag
		rows.Scan(&t.ID, &t.Name)
		b.Tags = append(b.Tags, t)
	}

	rows, err = db.Query("select a.id, a.image, a.date, a.title from articles a order by date desc limit 3")
	if err != nil {
		return nil, fmt.Errorf("Error: Could not get recent articles %v", err)
	}
	for rows.Next() {
		var a Article
		rows.Scan(&a.ID, &a.Image, &a.Date, &a.Title)
		a.Date = formatFrenchDate(a.Date)
		b.Recents = append(b.Recents, a)
	}

	return &b, nil
}

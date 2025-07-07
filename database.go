package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "konstantin"
	// password = "yourpassword"
	dbname   = "blog_db"
)

func DBConnect() *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable",
		host, port, user, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("DB connected successfully")

	return db
}

func CreateTable(db *sql.DB) {
	query := `
		CREATE TABLE IF NOT EXISTS posts (
			id SERIAL PRIMARY KEY,
			title VARCHAR(255) NOT NULL,
			content TEXT NOT NULL,
			category VARCHAR(100) NOT NULL,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS tags (
			id SERIAL PRIMARY KEY,
			name VARCHAR(100) NOT NULL UNIQUE
		);

		CREATE TABLE IF NOT EXISTS post_tags (
			post_id INTEGER REFERENCES posts(id) ON DELETE CASCADE,
			tag_id INTEGER REFERENCES tags(id) ON DELETE CASCADE,
			PRIMARY KEY (post_id, tag_id)
		);
	`

	_, err := db.Exec(query)
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println("Tables created successfully")
}

func CreatePost(db *sql.DB, post Post) (int, error) {
	var id int
	err := db.QueryRow(`
		INSERT INTO posts (title, content, category, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`,
		post.Title, post.Content, post.Category, time.Now(), time.Now(),
		).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func UpdatePost(db *sql.DB, post Post) {
	
}
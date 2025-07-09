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
	user     = "user"
	password = "password"
	dbname   = "dbname"
)

func DBConnect() *sql.DB {
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable",
		host, port, user, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
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
		log.Fatalf("failed to create tables: %v", err)
	}
	fmt.Println("Tables created successfully")
}

func AddOrUpdateTags(db *sql.DB, post Post) error {
	_, err := db.Exec(`DELETE FROM post_tags WHERE post_id = $1`, post.ID)
	if err != nil {
		return fmt.Errorf("failed to delete old tags for post %d: %w", post.ID, err)
	}

	for tag := range post.Tags {
		var tagID int
		err := db.QueryRow(`SELECT id FROM tags WHERE name = $1`, tag).Scan(&tagID)

		if err == sql.ErrNoRows {
			err = db.QueryRow(`INSERT INTO tags (name) VALUES ($1) RETURNING id`, tag).Scan(&tagID)
			if err != nil {
				return fmt.Errorf("failed to insert new tag '%d': %w", tag, err)
			}
		} else if err != nil {
			return fmt.Errorf("failed to query tag '%d': %w", tag, err)
		}

		_, err = db.Exec(`INSERT INTO post_tags VALUES ($1, $2)`, post.ID, tagID)
		if err != nil {
			return fmt.Errorf("failed to link tag %d to post %d: %w", tagID, post.ID, err)
		}
	}

	return nil
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
		return 0, fmt.Errorf("failed to create post: %w", err)
	}

	post.ID = id
	if err := AddOrUpdateTags(db, post); err != nil {
		return 0, fmt.Errorf("failed to add tags to new post: %w", err)
	}

	return id, nil
}

func UpdatePost(db *sql.DB, post Post) error {
	_, err := db.Exec(`
		UPDATE posts
		SET title = $1, content = $2, category = $3, updated_at = $4
		WHERE id = $5`,
		post.Title,
		post.Content,
		post.Category,
		time.Now(),
		post.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update post %d: %w", post.ID, err)
	}

	if err := AddOrUpdateTags(db, post); err != nil {
		return fmt.Errorf("failed to update tags for post %d: %w", post.ID, err)
	}

	return nil
}

func DeletePost(db *sql.DB, post Post) error {
	_, err := db.Exec(`DELETE FROM posts WHERE id = $1`, post.ID)
	if err != nil {
		return fmt.Errorf("failed to delete post %d: %w", post.ID, err)
	}
	return nil
}

func GetTags(db *sql.DB, postID int) ([]string, error) {
	var tags []string

	rows, err := db.Query(`SELECT tag_id FROM post_tags WHERE post_id = $1`, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to query tag IDs for post %d: %w", postID, err)
	}
	defer rows.Close()

	for rows.Next() {
		var tagID int
		if err := rows.Scan(&tagID); err != nil {
			return nil, fmt.Errorf("failed to scan tag ID for post %d: %w", postID, err)
		}

		var tag string
		err := db.QueryRow(`SELECT name FROM tags WHERE id = $1`, tagID).Scan(&tag)
		if err != nil {
			return nil, fmt.Errorf("failed to query tag name for ID %d: %w", tagID, err)
		}

		tags = append(tags, tag)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error for post %d tags: %w", postID, err)
	}

	return tags, nil
}

func GetPost(db *sql.DB, postID int) (Post, error) {
	var post Post

	err := db.QueryRow(`
		SELECT id, title, content, category, created_at, updated_at 
		FROM posts WHERE id = $1`,
		postID,
	).Scan(&post.ID, &post.Title, &post.Content, &post.Category, &post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return Post{}, fmt.Errorf("post with ID %d not found", postID)
		}
		return Post{}, fmt.Errorf("failed to query post %d: %w", postID, err)
	}

	tags, err := GetTags(db, post.ID)
	if err != nil {
		return Post{}, fmt.Errorf("failed to get tags for post %d: %w", post.ID, err)
	}
	post.Tags = tags

	return post, nil
}

func GetAllPosts(db *sql.DB) ([]Post, error) {
	var posts []Post

	rows, err := db.Query(`SELECT id, title, content, category, created_at, updated_at FROM posts`)
	if err != nil {
		return nil, fmt.Errorf("failed to query all posts: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Category, &post.CreatedAt, &post.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan post row: %w", err)
		}

		tags, err := GetTags(db, post.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get tags for post %d: %w", post.ID, err)
		}
		post.Tags = tags

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error for all posts: %w", err)
	}

	return posts, nil
}

func GetPostsByTerm(db *sql.DB, term string) ([]Post, error) {
	var posts []Post

	query := `
		SELECT id, title, content, category, created_at, updated_at 
		FROM posts 
		WHERE title LIKE '%' || $1 || '%'
		OR content LIKE '%' || $1 || '%'
		OR category LIKE '%' || $1 || '%'
	`

	rows, err := db.Query(query, term)
	if err != nil {
		return nil, fmt.Errorf("failed to search posts by term '%s': %w", term, err)
	}
	defer rows.Close()

	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content, &post.Category, &post.CreatedAt, &post.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan search result row: %w", err)
		}

		tags, err := GetTags(db, post.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get tags for post %d in search results: %w", post.ID, err)
		}
		post.Tags = tags

		posts = append(posts, post)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error for search results: %w", err)
	}

	return posts, nil
}

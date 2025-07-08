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

func AddOrUpdateTags(db *sql.DB, post Post) error {
	_, err := db.Exec(`DELETE * FROM post_tags WHERE post_id = $1`, post.ID)
	if err != nil {
		return err
	}

	for tag := range post.Tags {
		tagID, err := db.Exec(`SELECT id tags WHERE name = $1 RETURNING id`, tag)
		
		if err == sql.ErrNoRows {
			tagID, err = db.Exec(`INSERT INTO tags VALUES ($1) RETURNING id`, tag)
		}

		if err != nil {
			return err
		}

		_, err = db.Exec(`INSERT INTO post_tags VALUES ($1, $2)`, post.ID, tagID)

		if err != nil {
			return err
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
		return 0, err
	}

	err = AddOrUpdateTags(db, post)
	if err != nil {
		return 0, nil
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
		return err
	}

	err = AddOrUpdateTags(db, post)
	if err != nil {
		return nil
	}

	return nil
}

func DeletePost(db *sql.DB, post Post) error {
	_, err := db.Exec(`DELETE FROM posts WHERE id = $1`, post.ID)
	return err
}

func GetTags(db *sql.DB, postID int) ([]string, error) {
	var tags []string

	tagIDs, err := db.Query(`SELECT tag_id FROM post_tags WHERE post_id = $1`, postID)
	if err != nil {
		fmt.Printf("`SELECT tag_id FROM post_tags WHERE post_id = $1` query in GetTags() failed: %v.\n", err)
		return nil, err
	}
	defer tagIDs.Close()

	for tagIDs.Next() {
		var tagID int
		if err := tagIDs.Scan(&tagID); err != nil {
			fmt.Printf("tagIDs scan in GetTags() failed: %v.\n", err)
			return nil, err
		}

		var tag string
		err := db.QueryRow(`SELECT name FROM tags WHERE id = $1`, tagID).Scan(&tag)
		if err == sql.ErrNoRows {
			fmt.Printf("`SELECT name FROM tags WHERE id = $1` query in GetTags() failed: %v.\n", err)
			return nil, err
		}

		tags = append(tags, tag)
	}

	return tags, nil
}

func GetPost(db *sql.DB, postID int) (Post, error) {
	var post Post

	err := db.QueryRow(`SELECT * FROM posts WHERE id = $1`, postID).Scan(&post.ID, &post.Title, &post.Content, &post.Category,
		&post.CreatedAt, &post.UpdatedAt)
	if err != nil {
		fmt.Printf("`SELECT * FROM posts WHERE id = $1` query in GetPost() failed: %v.\n", err)
		return Post{}, err
	}

	rows, err := db.Query(`SELECT tag_id FROM post_tags WHERE post_id = $1`, postID)
	if err != nil {
		fmt.Printf("`SELECT tag_id FROM post_tags WHERE post_id = $1` query in GetPost() failed: %v.\n", err)
		return Post{}, err
	}
	defer rows.Close()

	for rows.Next() {
		post.Tags, err = GetTags(db, post.ID)
		if err != nil {
			return Post{}, err
		}
	}

	return post, nil
}

func GetAllPosts(db *sql.DB) ([]Post, error) {
	var posts []Post

	rows, err := db.Query(`SELECT * FROM posts`)
	if err != nil {
		fmt.Printf("'SELECT * FROM posts' query in GetAllPosts() failed: %v.\n", err)
		return nil, err
	}

	for rows.Next() {
		var post Post
		if err := rows.Scan(&post.ID, &post.Title, &post.Content,
			&post.Category, &post.CreatedAt, &post.UpdatedAt);
			err != nil {
			fmt.Printf("rows.Scan in GetAllPosts() failed: %v.\n", err)
			return nil, err
		}

		post.Tags, err = GetTags(db, post.ID)
		if err != nil {
			return nil, err
		}

		posts = append(posts, post)
	}

	return posts, nil
}

func GetPostsByTerm(db *sql.DB, term string) ([]Post, error) {
	var posts []Post

	query := `
		SELECT * FROM posts WHERE title LIKE '%' || ? || '%'
		OR content LIKE '%' || ? || '%'
		OR category LIKE '%' || ? || '%'
	`

	rows, err := db.Query(query, term)
	if err != nil {
		fmt.Printf("Search query in GetPostsByTerm() failed: %v.\n", err)
		return nil, err
	}

	for rows.Next() {
		var post Post
		if err := rows.Scan(&post); err != nil {
			fmt.Printf("rows scan in GetPostsByTerm() failed: %v.\n", err)
			return nil, err
		}

		post.Tags, err = GetTags(db, post.ID)
		if err != nil {
			return nil, err
		}

		posts = append(posts, post)
	}

	return posts, nil
}
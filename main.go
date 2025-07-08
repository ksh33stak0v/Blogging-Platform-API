package main

import (
	"fmt"
	"log"
	"database/sql"
)

func TestCRUDOperations(db *sql.DB) {
	// Тестирование создания поста
	fmt.Println("=== Testing CreatePost ===")
	post := Post{
		Title:    "First Post",
		Content:  "This is the content of the first post",
		Category: "Technology",
		Tags:     []string{"golang", "database", "web"},
	}
	id, err := CreatePost(db, post)
	if err != nil {
		log.Fatalf("CreatePost failed: %v", err)
	}
	fmt.Printf("Created post with ID: %d\n", id)

	// Тестирование получения поста
	fmt.Println("\n=== Testing GetPost ===")
	retrievedPost, err := GetPost(db, id)
	if err != nil {
		log.Fatalf("GetPost failed: %v", err)
	}
	fmt.Printf("Retrieved post: %+v\n", retrievedPost)

	// Тестирование обновления поста
	fmt.Println("\n=== Testing UpdatePost ===")
	updatedPost := Post{
		ID:       id,
		Title:    "Updated First Post",
		Content:  "Updated content with more details",
		Category: "Programming",
		Tags:     []string{"golang", "backend", "postgres"},
	}
	err = UpdatePost(db, updatedPost)
	if err != nil {
		log.Fatalf("UpdatePost failed: %v", err)
	}
	fmt.Println("Post updated successfully")

	// Проверка обновлений
	updatedRetrievedPost, err := GetPost(db, id)
	if err != nil {
		log.Fatalf("GetPost after update failed: %v", err)
	}
	fmt.Printf("Updated post details: %+v\n", updatedRetrievedPost)

	// Тестирование поиска по термину
	fmt.Println("\n=== Testing GetPostsByTerm ===")
	posts, err := GetPostsByTerm(db, "Updated")
	if err != nil {
		log.Fatalf("GetPostsByTerm failed: %v", err)
	}
	fmt.Println("Found posts with term 'Updated':")
	for _, p := range posts {
		fmt.Printf("- %s (ID: %d)\n", p.Title, p.ID)
	}

	// Тестирование получения всех постов
	fmt.Println("\n=== Testing GetAllPosts ===")
	allPosts, err := GetAllPosts(db)
	if err != nil {
		log.Fatalf("GetAllPosts failed: %v", err)
	}
	fmt.Println("All posts in database:")
	for _, p := range allPosts {
		fmt.Printf("- %s (ID: %d, Tags: %v)\n", p.Title, p.ID, p.Tags)
	}

	// Тестирование удаления поста
	fmt.Println("\n=== Testing DeletePost ===")
	err = DeletePost(db, Post{ID: id})
	if err != nil {
		log.Fatalf("DeletePost failed: %v", err)
	}
	fmt.Printf("Post with ID %d deleted successfully\n", id)

	// Проверка, что пост удален
	_, err = GetPost(db, id)
	if err == sql.ErrNoRows {
		fmt.Println("Post not found (as expected after deletion)")
	} else if err != nil {
		log.Fatalf("Unexpected error checking deleted post: %v", err)
	} else {
		fmt.Println("WARNING: Post still exists after deletion!")
	}
}

func TestTagOperations(db *sql.DB) {
	fmt.Println("\n=== Testing Tag Operations ===")
	
	// Создаем пост с тегами
	post := Post{
		Title:    "Tag Test Post",
		Content:  "This post tests tag functionality",
		Category: "Testing",
		Tags:     []string{"unit-test", "integration"},
	}
	id, err := CreatePost(db, post)
	if err != nil {
		log.Fatalf("CreatePost for tag test failed: %v", err)
	}
	fmt.Printf("Created post with ID: %d and tags: %v\n", id, post.Tags)

	// Проверяем теги
	retrievedPost, err := GetPost(db, id)
	if err != nil {
		log.Fatalf("GetPost for tag test failed: %v", err)
	}
	fmt.Printf("Retrieved tags: %v\n", retrievedPost.Tags)

	// Обновляем теги
	updatedPost := Post{
		ID:      id,
		Title:   retrievedPost.Title,
		Content: retrievedPost.Content,
		Tags:    []string{"unit-test", "regression", "qa"},
	}
	err = UpdatePost(db, updatedPost)
	if err != nil {
		log.Fatalf("UpdatePost for tag test failed: %v", err)
	}
	fmt.Println("Updated post tags to:", updatedPost.Tags)

	// Проверяем обновленные теги
	updatedRetrievedPost, err := GetPost(db, id)
	if err != nil {
		log.Fatalf("GetPost after tag update failed: %v", err)
	}
	fmt.Printf("Now has tags: %v\n", updatedRetrievedPost.Tags)

	// Удаляем тестовый пост
	err = DeletePost(db, Post{ID: id})
	if err != nil {
		log.Fatalf("DeletePost for tag test failed: %v", err)
	}
	fmt.Printf("Post with ID %d deleted\n", id)
}

func main() {
	db := DBConnect()
	CreateTable(db)
	defer db.Close()

	// Запуск тестов
	TestCRUDOperations(db)
	TestTagOperations(db)
}
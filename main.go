package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"strconv"
	"database/sql"
)

func main() {
	db := DBConnect()
	defer db.Close()
	CreateTable(db)

	http.HandleFunc("/posts", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handleGetPosts(w, r, db)
		case http.MethodPost:
			handleCreatePost(w, r, db)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/posts/", func(w http.ResponseWriter, r *http.Request) {
		idStr := strings.TrimPrefix(r.URL.Path, "/posts/")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			http.Error(w, "Invalid post ID", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodGet:
			handleGetPost(w, r, db, id)
		case http.MethodPut:
			handleUpdatePost(w, r, db, id)
		case http.MethodDelete:
			handleDeletePost(w, r, db, id)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	fmt.Println("Server starting on port 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func handleGetPosts(w http.ResponseWriter, r *http.Request, db *sql.DB) {
    term := r.URL.Query().Get("term")
    
    var posts []Post
    var err error
    
    if term != "" {
        posts, err = GetPostsByTerm(db, term)
    } else {
        posts, err = GetAllPosts(db)
    }
    
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(posts)
}

func handleGetPost(w http.ResponseWriter, r *http.Request, db *sql.DB, id int) {
	post, err := GetPost(db, id)
	if err != nil {
		if err == sql.ErrNoRows {
			http.Error(w, "Post not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

func handleCreatePost(w http.ResponseWriter, r *http.Request, db *sql.DB) {
	var post Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id, err := CreatePost(db, post)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	post.ID = id
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(post)
}

func handleUpdatePost(w http.ResponseWriter, r *http.Request, db *sql.DB, id int) {
	var post Post
	if err := json.NewDecoder(r.Body).Decode(&post); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	post.ID = id
	if err := UpdatePost(db, post); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(post)
}

func handleDeletePost(w http.ResponseWriter, r *http.Request, db *sql.DB, id int) {
	post := Post{ID: id}
	if err := DeletePost(db, post); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
)

func HandleGetPosts(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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

func HandleGetPost(w http.ResponseWriter, r *http.Request, db *sql.DB, id int) {
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

func HandleCreatePost(w http.ResponseWriter, r *http.Request, db *sql.DB) {
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

func HandleUpdatePost(w http.ResponseWriter, r *http.Request, db *sql.DB, id int) {
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

func HandleDeletePost(w http.ResponseWriter, r *http.Request, db *sql.DB, id int) {
	post := Post{ID: id}
	if err := DeletePost(db, post); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

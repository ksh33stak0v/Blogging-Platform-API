package main

import "time"

type Post struct {
	ID int
	Title string
	Content string
	Category string
	Tags []string
	CreatedAt time.Time
	UpdatedAt time.Time
}
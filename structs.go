package todo

import (
	"mime/multipart"
)

type User struct {
	Email    string `json:"email" binding:"required"`
	IsEmail  bool
	Password string `json:"password_hash" binding:"required"`
	Role     string `json:"role" binding:"required"`
}

type Post struct {
	Title           string `json:"title" binding:"required"`
	Content         string `json:"content" binding:"required"`
	Idempotency_key string `json:"idempotency_key" binding:"required"`
	Status          string `json:"status" binding:"required"`
}

type Image struct {
	FileHeader multipart.FileHeader `json:"FileHeader" binding:"required"`
	File       multipart.File       `json:"file" binding:"required"`
}

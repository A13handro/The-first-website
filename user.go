package todo

import "github.com/google/uuid"

type User struct {
	Id       uuid.UUID `json:"user_id" binding:"required"`
	Email    string    `json:"email" binding:"required"`
	IsEmail  bool
	Password string `json:"password_hash" binding:"required"`
	Role     string `json:"role" binding:"required"`
}

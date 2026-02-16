package repository

import (
	todo "the-first-website"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

type Authorization interface {
	CreateUser(user todo.User) (uuid.UUID, error)
	CheckUser(user todo.User) (uuid.UUID, string, error)
	GetID(string) (uuid.UUID, error)
}

type Posts interface {
}

type Picture interface {
}

type Repository struct {
	Authorization
	Posts
	Picture
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{
		Authorization: NewAuthPostgres(db),
	}
}

package service

import (
	todo "the-first-website"
	"the-first-website/pkg/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Authorization interface {
	CreateUser(user todo.User) (uuid.UUID, error)
	CheckUser(user todo.User) (uuid.UUID, error)
	GenerateToken(uuid.UUID, int) (string, error)
	ParseToken(string) (*jwt.Token, error)
	GetID(string) (uuid.UUID, error)
}

type Posts interface {
}

type Picture interface {
}

type Service struct {
	Authorization
	Posts
	Picture
}

func NewService(repos *repository.Repository) *Service {
	return &Service{
		Authorization: NewAuthService(repos.Authorization),
	}
}

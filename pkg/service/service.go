package service

import (
	todo "the-first-website"
	"the-first-website/pkg/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type Authorization interface {
	CreateUser(RefreshToken string, user todo.User) error
	CheckUser(user todo.User) (uuid.UUID, error)
	GenerateToken(uuid.UUID, int) (string, error)
	ParseToken(string) (*jwt.Token, error)
	GetID(string) (uuid.UUID, error)
	UpdateUser(string, uuid.UUID) error
}

type Posts interface {
	CreatePost(RefreshToken string, post todo.Post) error
	EditPost(RefreshToken string, post todo.Post, PostId string) error
	PublishPost(RefreshToken string, post todo.Post, PostId string) error
	ViewingPosts(RefreshToken string) ([]byte, error)
}

type Picture interface {
	AddImage(RefreshToken string, img todo.Image, PostId string) error
	DeleteImage(RefreshToken string, ImageId string, PostId string) error
}

type Service struct {
	Authorization
	Posts
	Picture
}

func NewService(repos *repository.Repository) *Service {
	return &Service{
		Authorization: NewAuthService(repos.Authorization),
		Posts:         NewPostService(repos.Posts),
		Picture:       NewPicturesService(repos.Picture),
	}
}

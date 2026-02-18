package repository

import (
	todo "the-first-website"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/minio/minio-go/v7"
)

type Authorization interface {
	CreateUser(RefreshToken string, user todo.User) error
	CheckUser(user todo.User) (uuid.UUID, string, error)
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

type Repository struct {
	Authorization
	Posts
	Picture
}

func NewRepository(db *sqlx.DB, minioClient *minio.Client) *Repository {
	return &Repository{
		Authorization: NewAuthPostgres(db),
		Posts:         NewPostPostgres(db),
		Picture:       NewPicturePostgres(db, minioClient),
	}
}

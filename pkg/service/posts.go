package service

import (
	todo "the-first-website"
	"the-first-website/pkg/repository"
)

type PostService struct {
	repo repository.Posts
}

func NewPostService(repo repository.Posts) *PostService {
	return &PostService{repo: repo}
}

func (s *PostService) CreatePost(RefreshToken string, post todo.Post) error {
	return s.repo.CreatePost(RefreshToken, post)
}

func (s *PostService) EditPost(RefreshToken string, post todo.Post, PostId string) error {
	return s.repo.EditPost(RefreshToken, post, PostId)
}

func (s *PostService) PublishPost(RefreshToken string, post todo.Post, PostId string) error {
	return s.repo.PublishPost(RefreshToken, post, PostId)
}

func (s *PostService) ViewingPosts(RefreshToken string) ([]byte, error) {
	return s.repo.ViewingPosts(RefreshToken)
}

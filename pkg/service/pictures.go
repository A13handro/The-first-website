package service

import (
	todo "the-first-website"
	"the-first-website/pkg/repository"
)

type PicturesService struct {
	repo repository.Picture
}

func NewPicturesService(repo repository.Picture) *PicturesService {
	return &PicturesService{repo: repo}
}

func (s *PicturesService) AddImage(RefreshToken string, img todo.Image, PostId string) error {
	return s.repo.AddImage(RefreshToken, img, PostId)
}

func (s *PicturesService) DeleteImage(RefreshToken string, ImageId string, PostId string) error {
	return s.repo.DeleteImage(RefreshToken, ImageId, PostId)
}

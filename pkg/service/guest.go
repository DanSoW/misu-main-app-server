package service

import (
	articleModel "main-server/pkg/model/article"
	repository "main-server/pkg/repository"
)

/* Structure for this service */
type GuestService struct {
	repo repository.Guest
}

/* Function for create new service */
func NewGuestService(repo repository.Guest) *GuestService {
	return &GuestService{
		repo: repo,
	}
}

/* Get all articles */
func (s *GuestService) GetArticles() (articleModel.ArticlesModel, error) {
	return s.repo.GetArticles()
}

package service

import (
	articleModel "main-server/pkg/model/article"
	repository "main-server/pkg/repository"

	"github.com/gin-gonic/gin"
)

/* Structure for this service */
type ModeratorService struct {
	repo repository.Moderator
}

/* Function for create new service */
func NewModeratorService(repo repository.Moderator) *ModeratorService {
	return &ModeratorService{
		repo: repo,
	}
}

/* Method for get unchecked article */
func (s *ModeratorService) GetUncheckedArticle(uuid articleModel.ArticleUuidModel, c *gin.Context) (articleModel.ArticleModel, error) {
	return s.repo.GetUncheckedArticle(uuid, c)
}

/* Method for get all unchecked articles */
func (s *ModeratorService) GetUncheckedArticles(c *gin.Context) (articleModel.ArticlesModel, error) {
	return s.repo.GetUncheckedArticles(c)
}

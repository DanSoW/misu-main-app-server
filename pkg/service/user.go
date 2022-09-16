package service

import (
	articleModel "main-server/pkg/model/article"
	userModel "main-server/pkg/model/user"
	repository "main-server/pkg/repository"

	"github.com/gin-gonic/gin"
)

/* Structure for this service */
type UserService struct {
	repo repository.User
}

/* Function for create new service */
func NewUserService(repo repository.User) *UserService {
	return &UserService{
		repo: repo,
	}
}

/* ********** */
/* Methods for articles */

/* Create new article */
func (s *UserService) CreateArticle(c *gin.Context, data articleModel.ArticleCreateRequestModel) (bool, error) {
	return s.repo.CreateArticle(c, data)
}

/* Update article */
func (s *UserService) UpdateArticle(c *gin.Context, data articleModel.ArticleUpdateRequestModel) (bool, error) {
	return s.repo.UpdateArticle(c, data)
}

/* Delete article for user */
func (s *UserService) DeleteArticle(uuid articleModel.ArticleUuidModel, c *gin.Context) (articleModel.ArticleSuccessModel, error) {
	return s.repo.DeleteArticle(uuid, c)
}

/* Get information about article */
func (s *UserService) GetArticle(uuid articleModel.ArticleUuidModel, c *gin.Context) (articleModel.ArticleModel, error) {
	return s.repo.GetArticle(uuid, c)
}

/* Get information about all article for user */
func (s *UserService) GetArticles(c *gin.Context) (articleModel.ArticlesModel, error) {
	return s.repo.GetArticles(c)
}

/* ********** */

/* ********** */
/* Methods for profile */

/* Get information about profile user */
func (s *UserService) GetProfile(c *gin.Context) (userModel.UserProfileModel, error) {
	return s.repo.GetProfile(c)
}

func (s *UserService) UpdateProfile(c *gin.Context, data userModel.UserProfileDataModel) (userModel.UserProfileDataModel, error) {
	return s.repo.UpdateProfile(c, data)
}

/* ********** */

package service

import (
	articleModel "main-server/pkg/model/article"
	rbacModel "main-server/pkg/model/rbac"
	userModel "main-server/pkg/model/user"
	repository "main-server/pkg/repository"

	"github.com/gin-gonic/gin"
)

type Authorization interface {
	CreateUser(user userModel.UserRegisterModel) (userModel.UserAuthDataModel, error)
	LoginUser(user userModel.UserLoginModel) (userModel.UserAuthDataModel, error)
	LoginUserOAuth2(code string) (userModel.UserAuthDataModel, error)
	Refresh(data userModel.TokenLogoutDataModel, refreshToken string) (userModel.UserAuthDataModel, error)
	Logout(tokens userModel.TokenLogoutDataModel) (bool, error)
	Activate(link string) (bool, error)

	// Recover password
	RecoveryPassword(email string) (bool, error)
	ResetPassword(data userModel.ResetPasswordModel) (bool, error)
}

type Token interface {
	ParseToken(token, signingKey string) (userModel.TokenOutputParse, error)
	ParseTokenWithoutValid(token, signingKey string) (userModel.TokenOutputParse, error)
	ParseResetToken(pToken, signingKey string) (userModel.ResetTokenOutputParse, error)
}

type AuthType interface {
	GetAuthType(column, value string) (userModel.AuthTypeModel, error)
}

type User interface {
	// Article
	CreateArticle(c *gin.Context, data articleModel.ArticleCreateRequestModel) (bool, error)
	UpdateArticle(c *gin.Context, data articleModel.ArticleUpdateRequestModel) (bool, error)
	DeleteArticle(uuid articleModel.ArticleUuidModel, c *gin.Context) (articleModel.ArticleSuccessModel, error)
	GetArticle(uuid articleModel.ArticleUuidModel, c *gin.Context) (articleModel.ArticleModel, error)
	GetArticles(c *gin.Context) (articleModel.ArticlesModel, error)

	// Profile
	GetProfile(c *gin.Context) (userModel.UserProfileModel, error)
	UpdateProfile(c *gin.Context, data userModel.UserProfileDataModel) (userModel.UserProfileDataModel, error)
}

type Moderator interface {
	GetUncheckedArticle(uuid articleModel.ArticleUuidModel, c *gin.Context) (articleModel.ArticleModel, error)
	GetUncheckedArticles(c *gin.Context) (articleModel.ArticlesModel, error)
}

type Guest interface {
	GetArticles() (articleModel.ArticlesModel, error)
}

type Domain interface {
	GetDomain(column, value interface{}) (rbacModel.DomainModel, error)
}

type Role interface {
	GetRole(column, value interface{}) (rbacModel.RoleModel, error)
	HasRole(usersId, domainsId int, roleValue string) (bool, error)
}

type Service struct {
	Authorization
	Token
	User
	Moderator
	Domain
	Role
	Guest
}

func NewService(repos *repository.Repository) *Service {
	tokenService := NewTokenService(repos.Role, repos.User, repos.AuthType)

	return &Service{
		Token:         tokenService,
		Authorization: NewAuthService(repos.Authorization, *tokenService),
		User:          NewUserService(repos.User),
		Moderator:     NewModeratorService(repos.Moderator),
		Guest:         NewGuestService(repos.Guest),
		Domain:        NewDomainService(repos.Domain),
		Role:          NewRoleService(repos.Role),
	}
}

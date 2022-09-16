package handler

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	_ "main-server/docs"

	route "main-server/pkg/constant/route"
	service "main-server/pkg/service"

	_ "github.com/swaggo/files"
	swaggerFiles "github.com/swaggo/files"
	_ "github.com/swaggo/gin-swagger"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Handler struct {
	services *service.Service
}

func NewHandler(services *service.Service) *Handler {
	return &Handler{services: services}
}

/* Инициализация маршрутов */
func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()

	router.MaxMultipartMemory = 50 << 20 // 50 MiB
	router.Static("/public", "./public")

	router.LoadHTMLGlob("pkg/template/*")

	// Настройка CORS-политики
	router.Use(cors.New(cors.Config{
		//AllowAllOrigins: true,
		AllowOrigins:     []string{viper.GetString("client_url"), viper.GetString("crm_url")},
		AllowMethods:     []string{"POST", "GET"},
		AllowHeaders:     []string{"Origin", "Content-type", "Authorization"},
		AllowCredentials: true,
	}))

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Запросы
	auth := router.Group(route.AUTH_MAIN_ROUTE)
	{
		auth.POST(route.AUTH_SIGN_UP_ROUTE, h.signUp)
		auth.POST(route.AUTH_SIGN_IN_ROUTE, h.signIn)
		auth.POST(route.AUTH_SIGN_IN_GOOGLE_ROUTE, h.signInOAuth2)
		auth.GET(route.AUTH_ACTIVATE_ROUTE, h.activate)

		// With middlewares (for get data from access token)
		auth.POST(route.AUTH_REFRESH_TOKEN_ROUTE, h.userIdentityLogout, h.refresh)
		auth.POST(route.AUTH_LOGOUT_ROUTE, h.userIdentity, h.logout)

		// Recover password
		auth.POST(route.AUTH_RECOVERY_PASSWORD, h.recoveryPassword)
		auth.POST(route.AUTH_RESET_PASSWORD, h.resetPassword)
	}

	// Группа запросов, связанных с пользователями
	user := router.Group(route.USER_MAIN_ROUTE, h.userIdentity)
	{
		// Группа запросов, связанных со статьями
		article := user.Group(route.USER_ARTICLE_ROUTE, h.userIdentityHasRoleUser)
		{
			// URL: /user/article/create
			article.POST(route.CREATE_ROUTE, h.createArticle)

			// URL: /user/article/update
			article.POST(route.UPDATE_ROUTE, h.updateArticle)

			// URL: /user/article/delete
			article.POST(route.DELETE_ROUTE, h.deleteArticle)

			// URL: /user/article/get
			article.POST(route.GET_ROUTE, h.getArticle)

			// URL: /user/article/get/all
			article.POST(route.GET_ALL_ROUTE, h.getArticles)
		}

		// Группа запросов, связанных с профилем пользователя
		profile := user.Group(route.USER_PROFILE_ROUTE)
		{
			// URL: /user/profile/get
			profile.POST(route.GET_ROUTE, h.getProfile)

			// URL: /user/profile/update
			profile.POST(route.UPDATE_ROUTE, h.updateProfile)
		}
	}

	// Route group for the moderator
	moderator := router.Group(route.MODERATOR_MAIN_ROUTE, h.userIdentity, h.userIdentityHasRoleModerator)
	{
		unchecked := moderator.Group(route.MODERATOR_UNCHECKED_ROUTE)
		{
			article := unchecked.Group(route.MODERATOR_ARTICLE_ROUTE)
			{
				article.POST(route.GET_ROUTE, h.getUncheckedArticle)
				article.POST(route.GET_ALL_ROUTE, h.getUncheckedArticles)
			}
		}
	}

	// Route group for the guest
	guest := router.Group(route.GUEST_MAIN_ROUTE)
	{
		article := guest.Group(route.GUEST_ARTICLE_ROUTE)
		{
			article.POST(route.GET_ALL_ROUTE, h.guestGetArticles)
		}
	}

	/*api := router.Group("/api", h.userIdentity)
	{
		lists := api.Group("/lists")
		{
			lists.POST("/", h.createList)
			lists.GET("/", h.getAllLists)
			lists.GET("/:id", h.getListById)
			lists.PUT("/:id", h.updateList)
			lists.DELETE("/:id", h.deleteList)

			items := lists.Group(":id/items")
			{
				items.POST("/", h.createItem)
				items.GET("/", h.getAllItems)
				items.GET("/:item_id", h.getItemById)
				items.PUT("/:item_id", h.updateItem)
				items.DELETE("/:item_id", h.deleteItem)
			}
		}
	}*/

	return router
}

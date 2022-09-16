package handler

import (
	articleModel "main-server/pkg/model/article"
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary GetArticle
// @Tags article
// @Description Get information about article
// @ID get-article
// @Accept  json
// @Produce  json
// @Param input body articleModel.ArticleUuidModel true "credentials"
// @Success 200 {object} articleModel.ArticleModel "data"
// @Failure 400,404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /moderator/unchecked/article/get [post]
func (h *Handler) getUncheckedArticle(c *gin.Context) {
	var input articleModel.ArticleUuidModel

	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	data, err := h.services.Moderator.GetUncheckedArticle(input, c)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, data)
}

// @Summary GetArticles
// @Tags article
// @Description Получение списка статей
// @ID get-articles
// @Accept  json
// @Produce  json
// @Success 200 {object} articleModel.ArticlesModel "data"
// @Failure 400,404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /moderator/unchecked/article/get/all [post]
func (h *Handler) getUncheckedArticles(c *gin.Context) {
	data, err := h.services.Moderator.GetUncheckedArticles(c)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, data)
}

package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary guestGetArticles
// @Tags guest
// @Description Получение списка статей
// @ID get-articles
// @Accept  json
// @Produce  json
// @Success 200 {object} articleModel.ArticlesModel "data"
// @Failure 400,404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /guest/article/get/all [post]
func (h *Handler) guestGetArticles(c *gin.Context) {
	data, err := h.services.Guest.GetArticles()
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, data)
}

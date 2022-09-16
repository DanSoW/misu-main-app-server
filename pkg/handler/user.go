package handler

import (
	"encoding/json"
	userModel "main-server/pkg/model/user"
	"net/http"

	"github.com/gin-gonic/gin"
)

// @Summary GetProfile
// @Tags profile
// @Description Get user profile
// @ID get-profile
// @Accept  json
// @Produce  json
// @Success 200 {object} user.UserProfileDataModel "data"
// @Failure 400,404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /user/profile/get [post]
func (h *Handler) getProfile(c *gin.Context) {
	data, err := h.services.User.GetProfile(c)

	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	var userProfile userModel.UserProfileDataModel

	err = json.Unmarshal([]byte(data.Data), &userProfile)

	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	userProfile.Email = data.Email

	c.JSON(http.StatusOK, userProfile)
}

// @Summary UpdateProfile
// @Tags profile
// @Description Update user profile
// @ID update-profile
// @Accept  json
// @Produce  json
// @Param input body user.UserProfileDataModel true "credentials"
// @Success 200 {object} user.UserProfileDataModel "data"
// @Failure 400,404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /user/profile/update [post]
func (h *Handler) updateProfile(c *gin.Context) {
	var input userModel.UserProfileDataModel

	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	data, err := h.services.User.UpdateProfile(c, input)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, data)
}

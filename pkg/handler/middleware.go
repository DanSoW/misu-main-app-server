package handler

import (
	"errors"
	middlewareConstants "main-server/pkg/constant/middleware"
	roleConstant "main-server/pkg/constant/role"
	authService "main-server/pkg/service/auth"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

/* Обработчик для проверки токена доступа пользователя */
func (h *Handler) userIdentity(c *gin.Context) {
	// Получение данных из заголовка Authorization
	header := c.GetHeader(middlewareConstants.AUTHORIZATION_HEADER)

	if header == "" {
		newErrorResponse(c, http.StatusUnauthorized, "Пустой заголовок авторизации!")
		return
	}

	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 {
		newErrorResponse(c, http.StatusUnauthorized, "Не корректный авторизационный заголовок!")
		return
	}

	// Парсинг токена доступа
	data, err := h.services.Token.ParseToken(headerParts[1], viper.GetString("token.signing_key_access"))

	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	// Получение текущего домена серверного приложения
	domain, err := h.services.Domain.GetDomain("value", viper.GetString("domain"))

	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	// Проверка токена доступа по принадлежности к определённому типу авторизации пользователя
	switch data.AuthType.Value {
	case "GOOGLE":
		if result, err := authService.VerifyAccessToken(*data.TokenApi); err != nil || result != true {
			newErrorResponse(c, http.StatusUnauthorized, "Не действительный токен доступа")
			return
		}
		break

	case "LOCAL":
		break
	}

	// Добавление к контексту дополнительных данных о пользователе
	c.Set(middlewareConstants.USER_CTX, data.UsersId)
	c.Set(middlewareConstants.AUTH_TYPE_VALUE_CTX, data.AuthType.Value)
	c.Set(middlewareConstants.TOKEN_API_CTX, data.TokenApi)
	c.Set(middlewareConstants.ACCESS_TOKEN_CTX, headerParts[1])
	c.Set(middlewareConstants.DOMAINS_ID, domain.Id)
}

func (h *Handler) userIdentityLogout(c *gin.Context) {
	header := c.GetHeader(middlewareConstants.AUTHORIZATION_HEADER)

	if header == "" {
		newErrorResponse(c, http.StatusUnauthorized, "Пустой заголовок авторизации!")
		return
	}

	headerParts := strings.Split(header, " ")
	if len(headerParts) != 2 {
		newErrorResponse(c, http.StatusUnauthorized, "Не корректный авторизационный заголовок!")
		return
	}

	data, err := h.services.Token.ParseTokenWithoutValid(headerParts[1], viper.GetString("token.signing_key_access"))

	if err != nil {
		newErrorResponse(c, http.StatusUnauthorized, err.Error())
		return
	}

	// Добавление к контексту дополнительных данных о пользователе
	c.Set(middlewareConstants.USER_CTX, data.UsersId)
	c.Set(middlewareConstants.AUTH_TYPE_VALUE_CTX, data.AuthType.Value)
	c.Set(middlewareConstants.TOKEN_API_CTX, data.TokenApi)
	c.Set(middlewareConstants.ACCESS_TOKEN_CTX, headerParts[1])
}

func (h *Handler) userIdentityHasRoleUser(c *gin.Context) {
	usersId, _ := c.Get(middlewareConstants.USER_CTX)
	domainsId, _ := c.Get(middlewareConstants.DOMAINS_ID)

	has, err := h.services.Role.HasRole(usersId.(int), domainsId.(int), roleConstant.ROLE_USER)

	if (err != nil) || (!has) {
		newErrorResponse(c, http.StatusForbidden, "Нет доступа!")
		return
	}
}

func (h *Handler) userIdentityHasRoleModerator(c *gin.Context) {
	usersId, _ := c.Get(middlewareConstants.USER_CTX)
	domainsId, _ := c.Get(middlewareConstants.DOMAINS_ID)

	has, err := h.services.Role.HasRole(usersId.(int), domainsId.(int), roleConstant.ROLE_MODERATOR)

	if (err != nil) || (!has) {
		newErrorResponse(c, http.StatusForbidden, "Нет доступа!")
		return
	}
}

func getUserId(c *gin.Context) (int, error) {
	id, ok := c.Get(middlewareConstants.USER_CTX)
	if !ok {
		return 0, errors.New("user id not found")
	}

	idInt, ok := id.(int)
	if !ok {
		return 0, errors.New("user id is of invalid type")
	}

	return idInt, nil
}

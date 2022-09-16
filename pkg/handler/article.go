package handler

import (
	articleModel "main-server/pkg/model/article"
	util "main-server/pkg/util"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	uuid "github.com/satori/go.uuid"
)

// @Summary CreateArticle
// @Tags article
// @Description Create article
// @ID create-article
// @Accept  json
// @Produce  json
// @Success 200 {object} articleModel.ArticleSuccessModel "data"
// @Failure 400,404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /user/article/create [post]
func (h *Handler) createArticle(c *gin.Context) {
	// Получение данных в формате multipart/form-data
	form, err := c.MultipartForm()
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	// Получение данных из полей формы
	files := form.File["files"]
	text := c.PostForm("text")
	tags := c.PostForm("tags")
	title := c.PostForm("title")
	image := form.File["title_file"]
	titleImage := image[len(image)-1]

	// Массив файлов, представляющих собой размеченные изображения
	var arrayFiles []articleModel.ArticlesFilesDBModel

	for _, file := range files {
		newFilename := uuid.NewV4().String()
		filepath := "public/" + newFilename
		index, err := strconv.Atoi(strings.Split(file.Filename, ".")[0])

		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, err.Error())
			return
		}

		arrayFiles = append(arrayFiles, articleModel.ArticlesFilesDBModel{
			Filename: newFilename,
			Filepath: filepath,
			Index:    index,
		})

		// Сохранение всех загруженных изображений
		c.SaveUploadedFile(file, filepath)
	}

	newFilename := uuid.NewV4().String()
	filepath := "public/" + newFilename

	// Сохранение главного изображения статьи
	c.SaveUploadedFile(titleImage, filepath)

	data, err := h.services.User.CreateArticle(c, articleModel.ArticleCreateRequestModel{
		Title:    title,
		Text:     text,
		Tags:     tags,
		Files:    &arrayFiles,
		Filename: &newFilename,
		Filepath: &filepath,
	})

	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, articleModel.ArticleSuccessModel{
		Success: data,
	})
}

// @Summary UpdateArticle
// @Tags article
// @Description Update article
// @ID update-article
// @Accept  json
// @Produce  json
// @Success 200 {object} articleModel.ArticleSuccessModel "data"
// @Failure 400,404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /user/article/create [post]
func (h *Handler) updateArticle(c *gin.Context) {
	form, err := c.MultipartForm()
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	articleUuid := c.PostForm("uuid")

	var files = form.File["files"]
	if len(files) <= 0 {
		files = nil
	}

	// Удаляемые файлы
	deletedFiles := c.PostFormArray("files_deleted")

	text := c.PostForm("text")
	tags := c.PostForm("tags")
	title := c.PostForm("title")
	image := form.File["title_file"]

	var titleImage *multipart.FileHeader = nil
	if len(image) > 0 {
		titleImage = image[len(image)-1]
	}

	var arrayFiles []articleModel.ArticlesFilesDBModel
	var filesDeletedArray []int

	for _, index := range deletedFiles {
		value, _ := strconv.Atoi(index)
		filesDeletedArray = append(filesDeletedArray, value)
	}

	var pointerArrayDeleteFiles *[]int

	if len(filesDeletedArray) <= 0 {
		pointerArrayDeleteFiles = nil
	} else {
		pointerArrayDeleteFiles = &filesDeletedArray
	}

	for _, file := range files {
		newFilename := uuid.NewV4().String()
		filepath := "public/" + newFilename
		index, err := strconv.Atoi(strings.Split(file.Filename, ".")[0])

		exists, _ := util.InArray(index, filesDeletedArray)

		if exists {
			newErrorResponse(c, http.StatusInternalServerError, "Ошибка! В массиве удаляемых файлов найдена ссылка на добавляемый")
			return
		}

		if err != nil {
			newErrorResponse(c, http.StatusInternalServerError, err.Error())
			return
		}

		arrayFiles = append(arrayFiles, articleModel.ArticlesFilesDBModel{
			Filename: newFilename,
			Filepath: filepath,
			Index:    index,
		})

		c.SaveUploadedFile(file, filepath)
	}

	textFilename := uuid.NewV4().String()
	textFilepath := "public/" + textFilename

	var pointerFilename *string = &textFilename
	var pointerFilepath *string = &textFilepath
	var pointerArrayFiles *[]articleModel.ArticlesFilesDBModel = &arrayFiles

	if titleImage == nil {
		pointerFilename = nil
		pointerFilepath = nil
	} else {
		c.SaveUploadedFile(titleImage, *pointerFilepath)
	}

	if files == nil {
		pointerArrayFiles = nil
	}

	data, err := h.services.User.UpdateArticle(c, articleModel.ArticleUpdateRequestModel{
		Uuid:        articleUuid,
		Title:       title,
		Text:        text,
		Tags:        tags,
		Files:       pointerArrayFiles,
		FilesDelete: pointerArrayDeleteFiles,
		Filename:    pointerFilename,
		Filepath:    pointerFilepath,
	})

	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, articleModel.ArticleSuccessModel{
		Success: data,
	})
}

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
// @Router /user/article/get [post]
func (h *Handler) getArticle(c *gin.Context) {
	var input articleModel.ArticleUuidModel

	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	data, err := h.services.User.GetArticle(input, c)
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
// @Router /user/article/get/all [post]
func (h *Handler) getArticles(c *gin.Context) {
	data, err := h.services.User.GetArticles(c)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, data)
}

// @Summary DeleteArticle
// @Tags article
// @Description Удаление статьи
// @ID delete-article
// @Accept  json
// @Produce  json
// @Param input body articleModel.ArticleUuidModel true "credentials"
// @Success 200 {object} articleModel.ArticleSuccessModel "data"
// @Failure 400,404 {object} errorResponse
// @Failure 500 {object} errorResponse
// @Failure default {object} errorResponse
// @Router /user/article/delete [post]
func (h *Handler) deleteArticle(c *gin.Context) {
	var input articleModel.ArticleUuidModel

	if err := c.BindJSON(&input); err != nil {
		newErrorResponse(c, http.StatusBadRequest, "invalid input body")
		return
	}

	data, err := h.services.User.DeleteArticle(input, c)
	if err != nil {
		newErrorResponse(c, http.StatusInternalServerError, err.Error())
		return
	}

	c.JSON(http.StatusOK, data)
}

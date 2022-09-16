package repository

import (
	"encoding/json"
	"fmt"
	actionConstant "main-server/pkg/constant/action"
	middlewareConstants "main-server/pkg/constant/middleware"
	objectConstant "main-server/pkg/constant/object"
	tableConstants "main-server/pkg/constant/table"
	articleModel "main-server/pkg/model/article"
	rbacModel "main-server/pkg/model/rbac"
	userModel "main-server/pkg/model/user"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	uuid "github.com/satori/go.uuid"
)

type UserPostgres struct {
	db       *sqlx.DB
	enforcer *casbin.Enforcer
	domain   *DomainPostgres
}

/*
* Функция создания экземпляра сервиса
 */
func NewUserPostgres(db *sqlx.DB, enforcer *casbin.Enforcer, domain *DomainPostgres) *UserPostgres {
	return &UserPostgres{
		db:       db,
		enforcer: enforcer,
		domain:   domain,
	}
}

func (r *UserPostgres) GetUser(column, value interface{}) (userModel.UserModel, error) {
	var user userModel.UserModel
	query := fmt.Sprintf("SELECT * FROM %s WHERE %s=$1", tableConstants.USERS_TABLE, column.(string))

	var err error

	switch value.(type) {
	case int:
		err = r.db.Get(&user, query, value.(int))
		break
	case string:
		err = r.db.Get(&user, query, value.(string))
		break
	}

	return user, err
}

/* Создание новой статьи */
func (r *UserPostgres) CreateArticle(c *gin.Context, data articleModel.ArticleCreateRequestModel) (bool, error) {
	// Получение данных из контекста запроса (пользовательский и доменный ID)
	usersId, _ := c.Get(middlewareConstants.USER_CTX)
	domainsId, _ := c.Get(middlewareConstants.DOMAINS_ID)

	tx, err := r.db.Begin()
	if err != nil {
		return false, err
	}

	// Добавление общей информации о статье
	query := fmt.Sprintf("INSERT INTO %s (uuid, users_id, title, filename, filepath, text, tags, created_at, updated_at) values ($1, $2, $3, $4, $5, $6, $7, $8, $9) RETURNING id", tableConstants.ARTICLES_TABLE)
	var articleId int
	currentDate := time.Now()
	articleUuid := uuid.NewV4()

	row := tx.QueryRow(query, articleUuid, usersId, data.Title, data.Filename, data.Filepath, data.Text, data.Tags, currentDate, currentDate)
	if err := row.Scan(&articleId); err != nil {
		tx.Rollback()
		return false, err
	}

	// Добавление файлов статьи
	query = fmt.Sprintf("INSERT INTO %s (filename, filepath) values ($1, $2) RETURNING id", tableConstants.FILES_TABLE)
	var filesId []articleModel.FileArticleExModel

	for _, element := range *data.Files {
		var fileId int
		row := tx.QueryRow(query, element.Filename, element.Filepath)
		if err := row.Scan(&fileId); err != nil {
			tx.Rollback()
			return false, err
		}

		filesId = append(filesId, articleModel.FileArticleExModel{
			Filename: element.Filename,
			Filepath: element.Filepath,
			Index:    element.Index,
			Id:       fileId,
		})
	}

	// Связывание статьи с файлами
	query = fmt.Sprintf("INSERT INTO %s (articles_id, files_id, index) values ($1, $2, $3)", tableConstants.ARTICLES_FILES_TABLE)

	for _, element := range filesId {
		_, err = tx.Exec(query, articleId, element.Id, element.Index)
		if err != nil {
			tx.Rollback()
			return false, err
		}
	}

	// Добавление информации о новом ресурсе
	var typesObjects rbacModel.TypesObjectsModel

	query = fmt.Sprintf("SELECT * FROM %s WHERE value=$1", tableConstants.TYPES_OBJECTS_TABLE)

	err = r.db.Get(&typesObjects, query, objectConstant.TYPE_ARTICLE)
	if err != nil {
		tx.Rollback()
		return false, err
	}

	query = fmt.Sprintf("INSERT INTO %s (value, types_objects_id) values ($1, $2)", tableConstants.OBJECTS_TABLE)

	_, err = tx.Exec(query, articleUuid, typesObjects.Id)
	if err != nil {
		tx.Rollback()
		return false, err
	}

	var userId string = strconv.Itoa(usersId.(int))
	var domainId string = strconv.Itoa(domainsId.(int))

	// Обновление политик доступа для текущего пользователя
	_, err = r.enforcer.AddPolicies([][]string{
		{userId, domainId, articleUuid.String(), actionConstant.DELETE},
		{userId, domainId, articleUuid.String(), actionConstant.MODIFY},
		{userId, domainId, articleUuid.String(), actionConstant.READ},
	})

	if err != nil {
		tx.Rollback()
		return false, err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return false, err
	}

	return true, nil
}

/* Обновление информации о статье */
func (r *UserPostgres) UpdateArticle(c *gin.Context, data articleModel.ArticleUpdateRequestModel) (bool, error) {
	usersId, _ := c.Get(middlewareConstants.USER_CTX)

	var article articleModel.ArticleDBModel

	query := fmt.Sprintf("SELECT * FROM %s WHERE uuid=$1 AND users_id = $2", tableConstants.ARTICLES_TABLE)

	err := r.db.Get(&article, query, data.Uuid, usersId)
	if err != nil {
		return false, err
	}

	tx, err := r.db.Begin()
	if err != nil {
		return false, err
	}

	setValues := make([]string, 0)
	args := make([]interface{}, 0)
	argId := 1

	setValues = append(setValues, fmt.Sprintf("title=$%d", argId))
	args = append(args, data.Title)
	argId++

	if data.Filename != nil && data.Filepath != nil {
		setValues = append(setValues, fmt.Sprintf("filename=$%d", argId))
		args = append(args, *data.Filename)
		argId++

		setValues = append(setValues, fmt.Sprintf("filepath=$%d", argId))
		args = append(args, *data.Filepath)
		argId++

		// Удалениие предыдущего изображения статьи
		os.Remove(article.Filepath)
	}

	setValues = append(setValues, fmt.Sprintf("text=$%d", argId))
	args = append(args, data.Text)
	argId++

	setValues = append(setValues, fmt.Sprintf("tags=$%d", argId))
	args = append(args, data.Tags)
	argId++

	// Изменение времени обновления статья
	setValues = append(setValues, fmt.Sprintf("updated_at=$%d", argId))
	args = append(args, time.Now())
	argId++

	setQuery := strings.Join(setValues, ", ")

	query = fmt.Sprintf("UPDATE %s tl SET %s WHERE tl.uuid = $%d AND tl.users_id = $%d",
		tableConstants.ARTICLES_TABLE, setQuery, argId, argId+1)

	args = append(args, article.Uuid)
	args = append(args, usersId)

	// Обновления данных о статье
	_, err = r.db.Exec(query, args...)
	if err != nil {
		tx.Rollback()
		return false, err
	}

	// Добавление информации о новых изображениях
	query = fmt.Sprintf("INSERT INTO %s (filename, filepath) values ($1, $2) RETURNING id", tableConstants.FILES_TABLE)
	var filesId []articleModel.FileArticleExModel

	if data.Files != nil {
		for _, element := range *data.Files {
			var fileId int
			row := tx.QueryRow(query, element.Filename, element.Filepath)
			if err := row.Scan(&fileId); err != nil {
				tx.Rollback()
				return false, err
			}

			filesId = append(filesId, articleModel.FileArticleExModel{
				Filename: element.Filename,
				Filepath: element.Filepath,
				Index:    element.Index,
				Id:       fileId,
			})
		}
	}

	// Добавление связывания новых изображений со статьей
	query = fmt.Sprintf("INSERT INTO %s (articles_id, files_id, index) values ($1, $2, $3)", tableConstants.ARTICLES_FILES_TABLE)

	for _, element := range filesId {
		_, err = tx.Exec(query, article.Id, element.Id, element.Index)
		if err != nil {
			tx.Rollback()
			return false, err
		}
	}

	// Запросы на удаление
	query = fmt.Sprintf(`SELECT * FROM %s tl WHERE tl.index=$1 AND tl.articles_id=$2 LIMIT 1`, tableConstants.ARTICLES_FILES_TABLE)
	queryDelete := fmt.Sprintf(`DELETE FROM %s tl WHERE tl.index=$1 AND tl.files_id=$2`, tableConstants.ARTICLES_FILES_TABLE)
	queryDeleteFiles := fmt.Sprintf(`DELETE FROM %s tl WHERE tl.id=$1 RETURNING filepath`, tableConstants.FILES_TABLE)

	// Удаление старых файлов
	if data.FilesDelete != nil {
		for _, element := range *data.FilesDelete {
			var articleFile []articleModel.ArticlesFilesModel

			err := r.db.Select(&articleFile, query, element, article.Id)
			if err != nil {
				tx.Rollback()
				return false, err
			}

			if len(articleFile) <= 0 {
				continue
			}

			_, err = tx.Exec(queryDelete, element, articleFile[0].FilesId)
			if err != nil {
				tx.Rollback()
				return false, err
			}

			var filePath string
			row := tx.QueryRow(queryDeleteFiles, articleFile[0].FilesId)
			if err := row.Scan(&filePath); err != nil {
				tx.Rollback()
				return false, err
			}

			os.Remove(filePath)
		}
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return false, err
	}

	return true, nil
}

/* Получение информации о статье */
func (r *UserPostgres) GetArticle(uuid articleModel.ArticleUuidModel, c *gin.Context) (articleModel.ArticleModel, error) {
	usersId, _ := c.Get(middlewareConstants.USER_CTX)

	var article articleModel.ArticleDBModel

	query := fmt.Sprintf("SELECT * FROM %s WHERE %s.uuid = $1 AND %s.users_id=$2 LIMIT 1",
		tableConstants.ARTICLES_TABLE,
		tableConstants.ARTICLES_TABLE,
		tableConstants.ARTICLES_TABLE,
	)

	err := r.db.Get(&article, query, uuid.Uuid, usersId)
	if err != nil {
		return articleModel.ArticleModel{}, err
	}

	var articlesFiles []articleModel.ArticlesFilesDBModel

	query = fmt.Sprintf(`SELECT index, filename, filepath FROM %s JOIN %s ON %s.files_id = %s.id WHERE %s.articles_id=$1;`,
		tableConstants.ARTICLES_FILES_TABLE, tableConstants.FILES_TABLE,
		tableConstants.ARTICLES_FILES_TABLE, tableConstants.FILES_TABLE,
		tableConstants.ARTICLES_FILES_TABLE,
	)

	err = r.db.Select(&articlesFiles, query, article.Id)
	if err != nil {
		return articleModel.ArticleModel{}, err
	}

	return articleModel.ArticleModel{
		Uuid:      article.Uuid,
		Filepath:  article.Filepath,
		Title:     article.Title,
		Text:      article.Text,
		Tags:      article.Tags,
		Files:     articlesFiles,
		CreatedAt: article.CreatedAt,
		UpdatedAt: article.UpdatedAt,
	}, nil
}

func (r *UserPostgres) GetArticles(c *gin.Context) (articleModel.ArticlesModel, error) {
	usersId, _ := c.Get(middlewareConstants.USER_CTX)

	query := fmt.Sprintf("SELECT * FROM %s WHERE users_id = $1", tableConstants.ARTICLES_TABLE)

	var articlesDb []articleModel.ArticleDBModel
	err := r.db.Select(&articlesDb, query, usersId)

	if err != nil {
		return articleModel.ArticlesModel{}, err
	}

	var articles articleModel.ArticlesModel

	query = fmt.Sprintf(`SELECT index, filename, filepath FROM %s JOIN %s ON %s.files_id = %s.id WHERE %s.articles_id=$1;`,
		tableConstants.ARTICLES_FILES_TABLE, tableConstants.FILES_TABLE,
		tableConstants.ARTICLES_FILES_TABLE, tableConstants.FILES_TABLE,
		tableConstants.ARTICLES_FILES_TABLE,
	)

	for _, element := range articlesDb {
		var files []articleModel.ArticlesFilesDBModel
		err := r.db.Select(&files, query, element.Id)

		if err != nil {
			return articleModel.ArticlesModel{}, err
		}

		articles.Articles = append(articles.Articles, articleModel.ArticleModel{
			Uuid:      element.Uuid,
			Filepath:  element.Filepath,
			Title:     element.Title,
			Text:      element.Text,
			Tags:      element.Tags,
			Files:     files,
			CreatedAt: element.CreatedAt,
			UpdatedAt: element.UpdatedAt,
		})
	}

	return articles, nil
}

/* Удаление статьи */
func (r *UserPostgres) DeleteArticle(uuid articleModel.ArticleUuidModel, c *gin.Context) (articleModel.ArticleSuccessModel, error) {
	usersId, _ := c.Get(middlewareConstants.USER_CTX)

	var article articleModel.ArticleDBModel

	query := fmt.Sprintf("SELECT * FROM %s WHERE %s.uuid = $1 AND %s.users_id=$2 LIMIT 1",
		tableConstants.ARTICLES_TABLE,
		tableConstants.ARTICLES_TABLE,
		tableConstants.ARTICLES_TABLE,
	)

	err := r.db.Get(&article, query, uuid.Uuid, usersId)
	if err != nil {
		return articleModel.ArticleSuccessModel{}, err
	}

	var articlesFiles []articleModel.ArticlesFilesDBModel

	query = fmt.Sprintf(`SELECT files_id, index, filename, filepath FROM %s JOIN %s ON %s.files_id = %s.id WHERE %s.articles_id=$1;`,
		tableConstants.ARTICLES_FILES_TABLE, tableConstants.FILES_TABLE,
		tableConstants.ARTICLES_FILES_TABLE, tableConstants.FILES_TABLE,
		tableConstants.ARTICLES_FILES_TABLE,
	)

	err = r.db.Select(&articlesFiles, query, article.Id)
	if err != nil {
		return articleModel.ArticleSuccessModel{}, err
	}

	tx, err := r.db.Begin()
	if err != nil {
		return articleModel.ArticleSuccessModel{}, err
	}

	query = fmt.Sprintf(`DELETE FROM %s tl WHERE tl.files_id=$1`, tableConstants.ARTICLES_FILES_TABLE)
	queryFiles := fmt.Sprintf(`DELETE FROM %s tl WHERE tl.id=$1`, tableConstants.FILES_TABLE)

	// Удаление файлов
	for _, element := range articlesFiles {
		_, err = r.db.Query(query, element.FilesId)
		if err != nil {
			tx.Rollback()
			return articleModel.ArticleSuccessModel{}, err
		}

		_, err = r.db.Query(queryFiles, element.FilesId)
		if err != nil {
			tx.Rollback()
			return articleModel.ArticleSuccessModel{}, err
		}

		err = os.Remove(element.Filepath)
		if err != nil {
			tx.Rollback()
			return articleModel.ArticleSuccessModel{}, err
		}
	}

	err = os.Remove(article.Filepath)
	if err != nil {
		tx.Rollback()
		return articleModel.ArticleSuccessModel{}, err
	}

	query = fmt.Sprintf(`DELETE FROM %s tl WHERE tl.uuid=$1`, tableConstants.ARTICLES_TABLE)
	_, err = r.db.Query(query, article.Uuid)
	if err != nil {
		tx.Rollback()
		return articleModel.ArticleSuccessModel{}, err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return articleModel.ArticleSuccessModel{}, err
	}

	return articleModel.ArticleSuccessModel{
		Success: true,
	}, nil
}

func (r *UserPostgres) GetProfile(c *gin.Context) (userModel.UserProfileModel, error) {
	usersId, _ := c.Get(middlewareConstants.USER_CTX)

	var profile userModel.UserProfileModel
	var email userModel.UserEmailModel

	query := fmt.Sprintf("SELECT data FROM %s tl WHERE tl.users_id = $1 LIMIT 1",
		tableConstants.USERS_DATA_TABLE,
	)

	err := r.db.Get(&profile, query, usersId)
	if err != nil {
		return userModel.UserProfileModel{}, err
	}

	query = fmt.Sprintf("SELECT email FROM %s tl WHERE tl.id = $1 LIMIT 1", tableConstants.USERS_TABLE)

	err = r.db.Get(&email, query, usersId)
	if err != nil {
		return userModel.UserProfileModel{}, err
	}

	return userModel.UserProfileModel{
		Email: email.Email,
		Data:  profile.Data,
	}, nil
}

func (r *UserPostgres) UpdateProfile(c *gin.Context, data userModel.UserProfileDataModel) (userModel.UserProfileDataModel, error) {
	usersId, _ := c.Get(middlewareConstants.USER_CTX)

	userJsonb, err := json.Marshal(data)
	if err != nil {
		return userModel.UserProfileDataModel{}, err
	}

	tx, err := r.db.Begin()
	if err != nil {
		return userModel.UserProfileDataModel{}, err
	}

	query := fmt.Sprintf("UPDATE %s tl SET data=$1 WHERE tl.users_id = $2", tableConstants.USERS_DATA_TABLE)

	// Update data about user profile
	_, err = r.db.Exec(query, userJsonb, usersId)
	if err != nil {
		tx.Rollback()
		return userModel.UserProfileDataModel{}, err
	}

	err = tx.Commit()

	if err != nil {
		tx.Rollback()
		return userModel.UserProfileDataModel{}, err
	}

	return data, nil
}

package repository

import (
	"fmt"
	tableConstant "main-server/pkg/constant/table"
	tableConstants "main-server/pkg/constant/table"
	articleModel "main-server/pkg/model/article"

	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
)

/* Structure for this repository */
type ModeratorPostgres struct {
	db       *sqlx.DB
	enforcer *casbin.Enforcer
	domain   *DomainPostgres
}

/* Function for create repository */
func NewModeratorPostgres(db *sqlx.DB, enforcer *casbin.Enforcer, domain *DomainPostgres) *ModeratorPostgres {
	return &ModeratorPostgres{
		db:       db,
		enforcer: enforcer,
		domain:   domain,
	}
}

func (r *ModeratorPostgres) GetUncheckedArticle(uuid articleModel.ArticleUuidModel, c *gin.Context) (articleModel.ArticleModel, error) {
	var article articleModel.ArticleDBModel

	query := fmt.Sprintf("SELECT * FROM %s as tl WHERE tl.uuid = $1 LIMIT 1",
		tableConstant.ARTICLES_TABLE,
	)

	err := r.db.Get(&article, query, uuid.Uuid)
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

func (r *ModeratorPostgres) GetUncheckedArticles(c *gin.Context) (articleModel.ArticlesModel, error) {
	query := fmt.Sprintf(
		"SELECT * FROM %s AS a1 WHERE not exists(SELECT * FROM %s AS a2 WHERE a2.articles_id = a1.id);",
		tableConstant.ARTICLES_TABLE,
		tableConstant.ARTICLES_CHECKED_TABLE,
	)

	var articlesDb []articleModel.ArticleDBModel
	err := r.db.Select(&articlesDb, query)

	if err != nil {
		return articleModel.ArticlesModel{}, err
	}

	var articles articleModel.ArticlesModel

	query = fmt.Sprintf(`SELECT index, filename, filepath FROM %s JOIN %s ON %s.files_id = %s.id WHERE %s.articles_id=$1;`,
		tableConstant.ARTICLES_FILES_TABLE, tableConstant.FILES_TABLE,
		tableConstant.ARTICLES_FILES_TABLE, tableConstant.FILES_TABLE,
		tableConstant.ARTICLES_FILES_TABLE,
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

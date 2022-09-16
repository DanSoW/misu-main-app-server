package repository

import (
	"fmt"
	tableConstants "main-server/pkg/constant/table"
	articleModel "main-server/pkg/model/article"

	"github.com/jmoiron/sqlx"
)

type GuestPostgres struct {
	db *sqlx.DB
}

/*
* Функция создания экземпляра сервиса
 */
func NewGuestPostgres(db *sqlx.DB) *GuestPostgres {
	return &GuestPostgres{db: db}
}

/*
* Функция получения данных о роли
 */
func (r *GuestPostgres) GetArticles() (articleModel.ArticlesModel, error) {
	query := fmt.Sprintf("SELECT * FROM %s", tableConstants.ARTICLES_TABLE)

	var articlesDb []articleModel.ArticleDBModel
	err := r.db.Select(&articlesDb, query)

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

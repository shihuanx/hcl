package article_repository

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"huancuilou/internal/article/article_model"
)

type ArticleRepository struct {
	DB *gorm.DB
}

func NewArticleRepository(db *gorm.DB) *ArticleRepository {
	return &ArticleRepository{
		DB: db,
	}
}

func (a *ArticleRepository) AddArticle(tx *gorm.DB, article *article_model.Article) error {
	result := tx.Create(article)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (a *ArticleRepository) GetArticleByID(id int) (*article_model.ArticleWithNoLike, error) {
	var article article_model.Article
	result := a.DB.First(&article, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &article_model.ArticleWithNoLike{
		ID:        article.ID,
		Title:     article.Title,
		Content:   article.Content,
		Kind:      article.Kind,
		ManagerID: article.ManagerID,
		CreateAt:  article.CreateAt,
	}, nil
}

func (a *ArticleRepository) WriteLikesToMySQL(tx *gorm.DB, slice []int) error {
	result := tx.Model(&article_model.Article{}).Where("id = ?", slice[0]).Update("like", slice[1])
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (a *ArticleRepository) UpdateArticle(article *article_model.Article) (*article_model.Article, error) {
	newArticle := article_model.Article{}

	result := a.DB.Model(article).Where("id =?", article.ID).Updates(map[string]interface{}{
		"title":   article.Title,
		"content": article.Content,
		"like":    article.Like,
	})

	if result.Error != nil {
		return nil, result.Error
	}

	if result.RowsAffected == 0 {
		return nil, fmt.Errorf("未找到要更新的文章记录，ID: %d", article.ID)
	}

	result = a.DB.First(&newArticle, article.ID)
	if result.Error != nil {
		return nil, result.Error
	}

	return &article_model.Article{
		ID:        newArticle.ID,
		Title:     newArticle.Title,
		Content:   newArticle.Content,
		Kind:      newArticle.Kind,
		Like:      newArticle.Like,
		ManagerID: newArticle.ManagerID,
		CreateAt:  newArticle.CreateAt,
	}, nil
}

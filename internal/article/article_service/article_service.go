package article_service

import (
	"fmt"
	"huancuilou/common/utils"
	"huancuilou/internal/article/article_model"
	"huancuilou/internal/article/article_repository"
	"log"
	"time"
)

type ArticleService struct {
	articleRepository      *article_repository.ArticleRepository
	articleCacheRepository *article_repository.ArticleCacheRepository
}

func NewArticleService(articleRepository *article_repository.ArticleRepository, articleCacheRepository *article_repository.ArticleCacheRepository) *ArticleService {
	return &ArticleService{
		articleRepository:      articleRepository,
		articleCacheRepository: articleCacheRepository,
	}
}

func (a *ArticleService) AddArticle(article *article_model.Article, managerID int) error {
	article.ManagerID = managerID
	article.CreateAt = time.Now()
	article.Like = 0

	tx := a.articleRepository.DB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("ArticleService.AddArticle err: %w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("事务已回滚")
		}
	}()

	if err := a.articleRepository.AddArticle(tx, article); err != nil {
		tx.Rollback()
		return fmt.Errorf("ArticleService.AddArticle err: %w", err)
	}

	basicArticle := &article_model.BasicArticle{
		ID:        article.ID,
		Title:     article.Title,
		Content:   utils.Substring(article.Content, 5),
		Kind:      article.Kind,
		ManagerID: managerID,
		Like:      0,
	}

	if err := a.articleCacheRepository.AddBasicArticle(basicArticle); err != nil {
		tx.Rollback()
		return fmt.Errorf("ArticleService.AddArticle err: %w", err)
	}

	go func() {
		if err := a.articleCacheRepository.AddArticle(article); err != nil {
			log.Printf("缓存文章添加失败: %v", err)
		}
	}()

	if err := tx.Commit().Error; err != nil {
		log.Printf("事务提交失败: %v", err)
		return fmt.Errorf("ArticleService.AddArticle err: %w", err)
	}

	return nil
}

func (a *ArticleService) GetAllArticleByKind(kind string) ([]*article_model.BasicArticle, error) {
	articles, err := a.articleCacheRepository.GetAllArticleByKind(kind)
	if err != nil {
		return nil, fmt.Errorf("ArticleService.GetAllArticleByKind err: %w", err)
	}
	return articles, nil
}

func (a *ArticleService) GetArticle(id int) (*article_model.Article, error) {
	article, err := a.articleCacheRepository.GetArticleByID(id)
	if err != nil {
		return nil, fmt.Errorf("ArticleService.GetArticle err: %w", err)
	}
	if article == nil {
		log.Printf("从缓存中获取文章失败, 从数据库中获取文章")
		articleWithNoLike, err := a.articleRepository.GetArticleByID(id)
		if err != nil {
			return nil, fmt.Errorf("ArticleService.GetArticle err: %w", err)
		}
		if articleWithNoLike == nil {
			return nil, fmt.Errorf("ArticleService.GetArticle err: 文章不存在")
		}
		basicArticle, err := a.articleCacheRepository.GetBasicArticleByID(id)
		if err != nil {
			return nil, fmt.Errorf("ArticleService.GetArticle err: %w", err)
		}
		article = &article_model.Article{
			ID:        articleWithNoLike.ID,
			Title:     articleWithNoLike.Title,
			Content:   articleWithNoLike.Content,
			Kind:      articleWithNoLike.Kind,
			ManagerID: articleWithNoLike.ManagerID,
			CreateAt:  articleWithNoLike.CreateAt,
			Like:      basicArticle.Like,
		}
		go func() {
			log.Printf("向缓存中添加文章")
			if err := a.articleCacheRepository.AddArticle(article); err != nil {
				log.Printf("缓存文章添加失败: %v", err)
			}
		}()
		return article, nil
	}
	return article, nil
}

// PeriodicUpdateLikes 周期性更新文章点赞数据到 MySQL
func (a *ArticleService) PeriodicUpdateLikes(interval time.Duration) {
	ticker := time.NewTicker(interval)
	for {
		select {
		case <-ticker.C:
			// 从 Redis 获取文章点赞数据
			results, err := a.articleCacheRepository.GetAllArticlesFromHash()
			if err != nil || len(results) == 0 {
				log.Printf("从 Redis 获取文章点赞数据出错: %v", err)
				continue
			}
			log.Printf("从 Redis 获取文章点赞数据")

			// 将点赞数据回写到 MySQL
			if err := a.updateLikesInTransaction(results); err != nil {
				log.Printf("更新点赞数据到 MySQL 出错: %v", err)
				continue
			}
			log.Printf("更新点赞数据到 MySQL 成功")
		}
	}
}

// updateLikesInTransaction 在事务中更新点赞数据到 MySQL
func (a *ArticleService) updateLikesInTransaction(results [][]int) error {
	tx := a.articleRepository.DB.Begin()
	if tx.Error != nil {
		return tx.Error
	}

	for _, result := range results {

		err := a.articleRepository.WriteLikesToMySQL(tx, result)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

func (a *ArticleService) AddLikes(articleID int, userID int) error {
	err := a.articleCacheRepository.AddLikes(articleID, userID, 1)
	if err != nil {
		return fmt.Errorf("ArticleService.AddLikes err: %w", err)
	}
	return nil
}

func (a *ArticleService) RemoveLikes(articleID int, userID int) error {
	err := a.articleCacheRepository.RemoveLikes(articleID, userID, -1)
	if err != nil {
		return fmt.Errorf("ArticleService.RemoveLikes err: %w", err)
	}
	return nil
}

func (a *ArticleService) UpdateArticle(article *article_model.Article) error {
	//为防止文章点赞量丢失，先从缓存中获取点赞量
	like, err := a.articleCacheRepository.GetArticleByID(article.ID)
	if err != nil {
		return fmt.Errorf("ArticleService.UpdateArticle err: %w", err)
	}
	article.Like = like.Like

	//第一次删除缓存
	if err = a.articleCacheRepository.DeleteArticleForUpdate(article.ID); err != nil {
		return fmt.Errorf("ArticleService.UpdateArticle err: %w", err)
	}

	//更新数据库
	newArticle, err := a.articleRepository.UpdateArticle(article)
	if err != nil {
		return fmt.Errorf("ArticleService.UpdateArticle err: %w", err)
	}

	//第二次删除缓存
	err = a.articleCacheRepository.DeleteArticleForUpdate(article.ID)
	if err != nil {
		return fmt.Errorf("ArticleService.UpdateArticle err: %w", err)
	}

	//将更新后的基本文章重新加入到缓存
	go func() {
		if err := a.articleCacheRepository.AddBasicArticleWithoutAddList(newArticle); err != nil {
			log.Printf("ArticleService.UpdateArticle err:重新添加基本文章失败：%v", err)
		}
	}()

	return nil

}

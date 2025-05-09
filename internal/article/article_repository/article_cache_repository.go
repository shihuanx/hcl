package article_repository

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"huancuilou/common/utils"
	"huancuilou/internal/article/article_model"
	"log"
	"strconv"
	"time"
)

type ArticleCacheRepository struct {
	client *redis.Client
}

var prefix = "hcl:article"

func NewArticleCacheRepository(client *redis.Client) *ArticleCacheRepository {
	return &ArticleCacheRepository{client: client}
}

func (a *ArticleCacheRepository) AddBasicArticle(article *article_model.BasicArticle) error {
	ctx := context.Background()
	mapKey := fmt.Sprintf("%s:basic:map:%d", prefix, article.ID)
	basicArticleMap := map[string]interface{}{
		"id":         article.ID,
		"title":      article.Title,
		"content":    article.Content,
		"kind":       article.Kind,
		"like":       article.Like,
		"manager_id": article.ManagerID,
	}
	if err := a.client.HSet(ctx, mapKey, basicArticleMap).Err(); err != nil {
		return fmt.Errorf("ArticleCacheRepository.AddBasicArticle err: %w", err)
	}

	listKey := fmt.Sprintf("%s:basic:list:%s", prefix, article.Kind)
	if err := a.client.LPush(ctx, listKey, article.ID).Err(); err != nil {
		return fmt.Errorf("ArticleCacheRepository.AddBasicArticle err: %w", err)
	}

	return nil
}

func (a *ArticleCacheRepository) AddArticle(article *article_model.Article) error {
	ctx := context.Background()
	key := fmt.Sprintf("%s:full:%d", prefix, article.ID)
	articleMap := map[string]interface{}{
		"id":         article.ID,
		"title":      article.Title,
		"content":    article.Content,
		"kind":       article.Kind,
		"like":       article.Like,
		"manager_id": article.ManagerID,
		"create_at":  article.CreateAt,
	}
	if err := a.client.HSet(ctx, key, articleMap).Err(); err != nil {
		return fmt.Errorf("ArticleCacheRepository.AddArticle err: %w", err)
	}

	// 设置哈希表的过期时间，这里设置为 2 小时
	expiration := 2 * time.Hour
	if err := a.client.Expire(ctx, key, expiration).Err(); err != nil {
		return fmt.Errorf("设置哈希表过期时间时出错: %w", err)
	}

	return nil
}

func (a *ArticleCacheRepository) GetAllArticleByKind(kind string) ([]*article_model.BasicArticle, error) {
	var articles []*article_model.BasicArticle
	ctx := context.Background()
	listKey := fmt.Sprintf("%s:basic:list:%s", prefix, kind)
	articleIDs, err := a.client.LRange(ctx, listKey, 0, -1).Result()
	if err != nil {
		return nil, fmt.Errorf("获取列表时出错: %w", err)
	}

	for _, idStr := range articleIDs {
		id, err := strconv.Atoi(idStr)
		if err != nil {
			return nil, fmt.Errorf("转换ID时出错: %w", err)
		}
		mapKey := fmt.Sprintf("%s:basic:map:%d", prefix, id)
		articleMap, err := a.client.HGetAll(ctx, mapKey).Result()
		if err != nil {
			return nil, fmt.Errorf("获取哈希表时出错: %w", err)
		}
		if len(articleMap) == 0 {
			log.Printf("文章ID为: %d的基本文章在缓存中不存在", id)
			continue
		}

		like, err := strconv.Atoi(articleMap["like"])
		if err != nil {
			return nil, fmt.Errorf("转换点赞数时出错: %w", err)
		}
		managerID, err := strconv.Atoi(articleMap["manager_id"])
		if err != nil {
			return nil, fmt.Errorf("转换管理员id时出错: %w", err)
		}
		article := &article_model.BasicArticle{
			ID:        id,
			Title:     articleMap["title"],
			Content:   articleMap["content"],
			Kind:      articleMap["kind"],
			Like:      like,
			ManagerID: managerID,
		}
		articles = append(articles, article)
	}

	return articles, nil
}

func (a *ArticleCacheRepository) GetArticleByID(id int) (*article_model.Article, error) {
	ctx := context.Background()
	key := fmt.Sprintf("%s:full:%d", prefix, id)

	articleMap, err := a.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, fmt.Errorf("获取哈希表时出错: %w", err)
	}
	if len(articleMap) == 0 {
		return nil, nil
	}

	like, err := strconv.Atoi(articleMap["like"])
	if err != nil {
		return nil, fmt.Errorf("转换点赞数时出错: %w", err)
	}
	managerID, err := strconv.Atoi(articleMap["manager_id"])
	if err != nil {
		return nil, fmt.Errorf("转换管理员id时出错: %w", err)
	}

	createAtStr := articleMap["create_at"]
	// 定义时间格式，根据实际存储的时间格式进行调整
	layout := time.RFC3339Nano
	createAt, err := time.Parse(layout, createAtStr)
	if err != nil {
		return nil, fmt.Errorf("解析创建时间时出错: %w", err)
	}

	article := &article_model.Article{
		ID:        id,
		Title:     articleMap["title"],
		Content:   articleMap["content"],
		Kind:      articleMap["kind"],
		Like:      like,
		ManagerID: managerID,
		CreateAt:  createAt,
	}

	return article, nil
}

func (a *ArticleCacheRepository) GetBasicArticleByID(id int) (*article_model.BasicArticle, error) {
	ctx := context.Background()
	key := fmt.Sprintf("%s:basic:map:%d", prefix, id)
	articleMap, err := a.client.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if len(articleMap) == 0 {
		return nil, fmt.Errorf("基本文章类型在缓存不存在")
	}
	like, err := strconv.Atoi(articleMap["like"])
	if err != nil {
		return nil, err
	}
	managerID, err := strconv.Atoi(articleMap["manager_id"])
	if err != nil {
		return nil, err
	}
	return &article_model.BasicArticle{
		ID:        id,
		Title:     articleMap["title"],
		Content:   articleMap["content"],
		Kind:      articleMap["kind"],
		Like:      like,
		ManagerID: managerID,
	}, nil
}

func (a *ArticleCacheRepository) AddLikes(articleID int, userID int, i int) error {
	ctx := context.Background()
	articleLikeKey := fmt.Sprintf("%s:like:%d", prefix, articleID)
	exists, err := a.client.SIsMember(ctx, articleLikeKey, userID).Result()
	if err != nil {
		return fmt.Errorf("判断用户点赞时出错，文章 ID: %d, 用户 ID: %d, 错误信息: %w", articleID, userID, err)
	}
	if exists {
		return fmt.Errorf("用户已经点赞，无法重复点赞，文章 ID: %d, 用户 ID: %d", articleID, userID)
	}

	basicKey := fmt.Sprintf("%s:basic:map:%d", prefix, articleID)
	if err := a.client.HIncrBy(ctx, basicKey, "like", int64(i)).Err(); err != nil {
		return fmt.Errorf("增加基本信息哈希表点赞数时出错，文章 ID: %d, 错误信息: %w", articleID, err)
	}

	fullKey := fmt.Sprintf("%s:full:%d", prefix, articleID)
	if err := a.client.HIncrBy(ctx, fullKey, "like", int64(i)).Err(); err != nil {
		// 尝试回滚基本信息哈希表中的点赞数
		if rollbackErr := a.client.HIncrBy(ctx, basicKey, "like", int64(-i)).Err(); rollbackErr != nil {
			log.Printf("回滚基本文章点赞数时出错，文章 ID: %d, 错误信息: %v", articleID, rollbackErr)
			return fmt.Errorf("增加完整信息哈希表点赞数出错且回滚基本信息哈希表点赞数失败，文章 ID: %d, 原始错误: %w, 回滚错误: %v", articleID, err, rollbackErr)
		}
		return fmt.Errorf("增加完整文章表点赞数时出错，文章 ID: %d, 错误信息: %w", articleID, err)
	}
	if err := a.UseLikeSort(articleID, userID, "addLike"); err != nil {
		return err
	}
	return nil
}

func (a *ArticleCacheRepository) RemoveLikes(articleID int, userID int, i int) error {
	ctx := context.Background()
	articleLikeKey := fmt.Sprintf("%s:like:%d", prefix, articleID)
	exists, err := a.client.SIsMember(ctx, articleLikeKey, userID).Result()
	if err != nil {
		return fmt.Errorf("判断用户点赞时出错，文章 ID: %d, 用户 ID: %d, 错误信息: %w", articleID, userID, err)
	}
	if !exists {
		return fmt.Errorf("无法重复取消点赞，文章 ID: %d, 用户 ID: %d", articleID, userID)
	}

	basicKey := fmt.Sprintf("%s:basic:map:%d", prefix, articleID)
	if err := a.client.HIncrBy(ctx, basicKey, "like", int64(i)).Err(); err != nil {
		return fmt.Errorf("取消基本信息哈希表点赞数时出错，文章 ID: %d, 错误信息: %w", articleID, err)
	}

	fullKey := fmt.Sprintf("%s:full:%d", prefix, articleID)
	if err := a.client.HIncrBy(ctx, fullKey, "like", int64(i)).Err(); err != nil {
		// 尝试回滚基本信息哈希表中的点赞数
		if rollbackErr := a.client.HIncrBy(ctx, basicKey, "like", int64(-i)).Err(); rollbackErr != nil {
			log.Printf("回滚基本文章点赞数时出错，文章 ID: %d, 错误信息: %v", articleID, rollbackErr)
			return fmt.Errorf("取消完整信息哈希表点赞数出错且回滚基本信息哈希表点赞数失败，文章 ID: %d, 原始错误: %w, 回滚错误: %v", articleID, err, rollbackErr)
		}
		return fmt.Errorf("取消完整文章表点赞数时出错，文章 ID: %d, 错误信息: %w", articleID, err)
	}

	if err := a.UseLikeSort(articleID, userID, "removeLike"); err != nil {
		return err
	}

	return nil
}

func (a *ArticleCacheRepository) UseLikeSort(articleID int, userID int, option string) error {
	if option != "addLike" && option != "removeLike" {
		return fmt.Errorf("option参数错误")
	}

	ctx := context.Background()
	key := fmt.Sprintf("%s:like:%d", prefix, articleID)

	if option == "addLike" {
		return a.client.SAdd(ctx, key, userID).Err()
	}

	return a.client.SRem(ctx, key, userID).Err()
}

// GetAllArticlesFromHash 获取哈希表中的所有文章信息
func (a *ArticleCacheRepository) GetAllArticlesFromHash() ([][]int, error) {
	var cursor uint64
	var results [][]int
	ctx := context.Background()
	for {
		keys, nextCursor, err := a.client.Scan(ctx, cursor, "hcl:article:basic:map:*", 0).Result()
		log.Printf("扫描结果: %v", keys)
		if err != nil {
			return nil, err
		}
		for _, key := range keys {
			// 从键中提取文章 ID
			idStr := key[len("hcl:article:basic:map:"):]
			id, err := strconv.Atoi(idStr)
			if err != nil {
				log.Printf("转换文章 ID 出错: %v", err)
				continue
			}

			likeStr, err := a.client.HGet(ctx, key, "like").Result()
			if err != nil {
				log.Printf("获取文章 %d 点赞数出错: %v", id, err)
				continue
			}
			like, err := strconv.Atoi(likeStr)
			if err != nil {
				log.Printf("转换文章 %d 点赞数出错: %v", id, err)
				continue
			}

			results = append(results, []int{id, like})
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}

	}
	return results, nil
}

func (a *ArticleCacheRepository) DeleteArticleForUpdate(id int) error {
	ctx := context.Background()
	fullKey := fmt.Sprintf("%s:full:%d", prefix, id)
	if err := a.client.Del(ctx, fullKey).Err(); err != nil {
		return fmt.Errorf("删除完整文章时出错，文章 ID: %d, 错误信息: %w", id, err)
	}

	basicKey := fmt.Sprintf("%s:basic:map:%d", prefix, id)
	if err := a.client.Del(ctx, basicKey).Err(); err != nil {
		return fmt.Errorf("删除基本文章时出错，文章 ID: %d, 错误信息: %w", id, err)
	}

	return nil
}

func (a *ArticleCacheRepository) AddBasicArticleWithoutAddList(article *article_model.Article) error {
	ctx := context.Background()
	key := fmt.Sprintf("%s:basic:map:%d", prefix, article.ID)
	basicArticleMap := map[string]interface{}{
		"id":         article.ID,
		"title":      article.Title,
		"content":    utils.Substring(article.Content, 5),
		"kind":       article.Kind,
		"like":       article.Like,
		"manager_id": article.ManagerID,
	}
	if err := a.client.HSet(ctx, key, basicArticleMap).Err(); err != nil {
		return fmt.Errorf("添加基本文章时出错，文章 ID: %d, 错误信息: %w", article.ID, err)
	}
	return nil
}

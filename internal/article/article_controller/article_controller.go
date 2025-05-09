package article_controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"huancuilou/common/error_handler"
	"huancuilou/common/utils"
	"huancuilou/internal/article/article_model"
	"huancuilou/internal/article/article_service"
	"huancuilou/response"
	"net/http"
	"strconv"
)

type ArticleController struct {
	ArticleService *article_service.ArticleService
}

func NewArticleController(articleService *article_service.ArticleService) *ArticleController {
	return &ArticleController{
		ArticleService: articleService,
	}
}

func (a *ArticleController) AddArticle(c *gin.Context) {
	var article *article_model.Article
	if err := c.BindJSON(&article); err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("ArticleController.AddArticle err: 400:将json数据绑定到结构体失败:%w", err))
		return
	}
	if !utils.ValidateArticleKind(article.Kind) {
		error_handler.HandleUserError(c, fmt.Errorf("ArticleController.AddArticle err: 400:文章类型错误"))
		return
	}
	managerID := c.MustGet("userID").(int)
	if err := a.ArticleService.AddArticle(article, managerID); err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("ArticleController.AddArticle err: 500: %w", err))
		return
	}
	c.JSON(http.StatusOK, response.SuccessWithoutData())
}

func (a *ArticleController) GetAllArticle(c *gin.Context) {
	kind := c.Query("kind")

	articles, err := a.ArticleService.GetAllArticleByKind(kind)
	if err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("ArticleController.GetAllArticle err: 500: %w", err))
		return
	}

	c.JSON(http.StatusOK, response.Success(articles))
}

func (a *ArticleController) GetArticle(c *gin.Context) {
	articleIDStr := c.Param("articleID")
	articleID, err := strconv.Atoi(articleIDStr)
	if err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("ArticleController.GetArticle err: 400: 将articleID转换为int失败:%w", err))
		return
	}

	article, err := a.ArticleService.GetArticle(articleID)

	if err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("ArticleController.GetArticle err: 500: %w", err))
		return
	}

	c.JSON(http.StatusOK, response.Success(article))
}

func (a *ArticleController) AddLikes(c *gin.Context) {
	articleIDStr := c.Param("articleID")
	articleID, err := strconv.Atoi(articleIDStr)
	if err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("ArticleController.AddLikes err: 400: 将articleID转换为int失败:%w", err))
		return
	}
	userID := c.MustGet("userID").(int)

	if err := a.ArticleService.AddLikes(articleID, userID); err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("ArticleController.AddLikes err: 500: %w", err))
		return
	}
	c.JSON(http.StatusOK, response.SuccessWithoutData())
}

func (a *ArticleController) RemoveLikes(c *gin.Context) {
	articleIDStr := c.Param("articleID")
	articleID, err := strconv.Atoi(articleIDStr)
	if err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("ArticleController.RemoveLikes err: 400: 将articleID转换为int失败:%w", err))
		return
	}
	userID := c.MustGet("userID").(int)

	if err := a.ArticleService.RemoveLikes(articleID, userID); err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("ArticleController.RemoveLikes err: 500: %w", err))
		return
	}
	c.JSON(http.StatusOK, response.SuccessWithoutData())
}

func (a *ArticleController) UpdateArticle(c *gin.Context) {
	var article *article_model.Article
	if err := c.BindJSON(&article); err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("ArticleController.UpdateArticle err: 400:将json数据绑定到结构体失败:%w", err))
		return
	}

	if err := a.ArticleService.UpdateArticle(article); err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("ArticleController.UpdateArticle err: 500: %w", err))
		return
	}
	c.JSON(http.StatusOK, response.SuccessWithoutData())
}

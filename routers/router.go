package routers

import (
	"github.com/gin-gonic/gin"
	"huancuilou/common/utils"
	"huancuilou/internal/article/article_controller"
	"huancuilou/internal/user/user_controller"
)

// SetUpRouters 设置路由
func SetUpRouters(userController *user_controller.UserController, articleController *article_controller.ArticleController) *gin.Engine {
	r := gin.Default()

	userGroup := r.Group("/user")
	{
		userGroup.GET("/send-code/:phoneNumber", userController.SendCode)
		userGroup.POST("/login", userController.Login)
		userGroup.PUT("/add-administrator/:phoneNumber", utils.SuperAdminOnlyMiddleware(), userController.AddAdministrator)
		userGroup.GET("", utils.JwtInterceptor(), userController.GetUserInfo)
		userGroup.PUT("", utils.JwtInterceptor(), userController.UpdateUserInfo)
		userGroup.POST("/add-score", utils.JwtInterceptor(), userController.AddScore)
		userGroup.POST("/add-phone-record", utils.AdminOnlyMiddleware(), userController.AddPhoneRecord)
		userGroup.GET("/get-phone-record/:phoneNumber", utils.AdminOnlyMiddleware(), userController.GetPhoneRecordByPhone)
		userGroup.GET("/add-follows/:followerID", utils.JwtInterceptor(), userController.AddFollows)
		userGroup.GET("/get-follows", utils.JwtInterceptor(), userController.GetFollows)
		userGroup.GET("/get-other-user-info/:userID", utils.JwtInterceptor(), userController.GetOtherUserInfo)
		userGroup.DELETE("/remove-follows/:followID", utils.JwtInterceptor(), userController.RemoveFollows)
		userGroup.GET("/get-common-follows/:otherUserID", utils.JwtInterceptor(), userController.GetCommonFollows)
		userGroup.GET("/add-likes/:userID", utils.JwtInterceptor(), userController.AddLikes)
		userGroup.GET("/get-likes-rank", utils.JwtInterceptor(), userController.GetLikesRank)
		userGroup.POST("/add-item", utils.AdminOnlyMiddleware(), userController.AddItem)
		userGroup.GET("/get-all-items", utils.JwtInterceptor(), userController.GetAllItems)
		userGroup.GET("/choose-item/:itemID", utils.JwtInterceptor(), userController.ChooseItem)
		userGroup.GET("/add-item-consumer", utils.AdminOnlyMiddleware(), userController.AddChooseItemConsumer)
	}

	articleGroup := r.Group("/article")
	{
		articleGroup.POST("", utils.AdminOnlyMiddleware(), articleController.AddArticle)
		articleGroup.GET("/get-all-article", utils.JwtInterceptor(), articleController.GetAllArticle)
		articleGroup.GET("/:articleID", utils.JwtInterceptor(), articleController.GetArticle)
		articleGroup.GET("/add-likes/:articleID", utils.JwtInterceptor(), articleController.AddLikes)
		articleGroup.DELETE("/remove-likes/:articleID", utils.JwtInterceptor(), articleController.RemoveLikes)
		articleGroup.PUT("", utils.AdminOnlyMiddleware(), articleController.UpdateArticle)
	}

	return r
}

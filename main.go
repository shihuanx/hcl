package main

import (
	"huancuilou/configs"
	"huancuilou/initial"
	"huancuilou/internal/article/article_controller"
	"huancuilou/internal/article/article_repository"
	"huancuilou/internal/article/article_service"
	"huancuilou/internal/user/user_controller"
	"huancuilou/internal/user/user_repository"
	"huancuilou/internal/user/user_service"
	"huancuilou/routers"
	"log"
)

func main() {
	cfg := configs.GetConfig()
	db, err := initial.InitMysql(cfg.MySQL.DSN)
	RedisClient := initial.InitRedis(cfg.Redis)
	if err != nil {
		log.Fatalf("初始化数据库失败：%v", err)
	}

	//用户相关包的依赖注入
	userRepository := user_repository.NewUserRepository(db)
	userMdbRepository := user_repository.NewUserMemoryDBRepository()
	userCacheRepository := user_repository.NewUserCacheRepository(RedisClient)
	userService := user_service.NewUserService(userRepository, &cfg, userMdbRepository, userCacheRepository)
	userController := user_controller.NewUserController(userService, cfg.Jwt)

	//文章相关包的依赖注入
	articleRepository := article_repository.NewArticleRepository(db)
	articleCacheRepository := article_repository.NewArticleCacheRepository(RedisClient)
	articleService := article_service.NewArticleService(articleRepository, articleCacheRepository)
	articleController := article_controller.NewArticleController(articleService)

	Router := routers.SetUpRouters(userController, articleController)

	go func() {
		articleService.PeriodicUpdateLikes(cfg.Article.UpdateLikesInterval)
	}()

	if err = Router.Run(":8080"); err != nil {
		log.Fatalf("初始化路由失败：%v", err)
	}

}

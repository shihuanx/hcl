package initial

import (
	"github.com/redis/go-redis/v9"
	"huancuilou/configs"
)

var RedisClient *redis.Client

func InitRedis(redisConfig configs.RedisConfig) *redis.Client {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     redisConfig.Addr,
		Password: redisConfig.Password,
		DB:       redisConfig.DB,
	})
	return RedisClient
}

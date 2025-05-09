package configs

import (
	"time"
)

// MySQLConfig 定义 MySQL 配置结构体
type MySQLConfig struct {
	DSN string
}

// JwtConfig 定义 JWT 配置结构体
type JwtConfig struct {
	SecretKey                  string        //密钥
	AccessTokenExpireDuration  time.Duration //accessToken过期时间
	RefreshTokenExpireDuration time.Duration //refreshToken过期时间
}

// ArticleConfig 定义文章配置结构体
type ArticleConfig struct {
	KindMap             map[string]bool
	UpdateLikesInterval time.Duration
}

// CodeConfig 定义验证码配置结构体
type CodeConfig struct {
	ExpireDuration time.Duration // 定时删除过期验证码的间隔时间
}

// RedisConfig 定义 Redis 配置结构体
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

type RabbitMQConfig struct {
	DSN     string
	Durable bool
}

// Config 定义配置结构体
type Config struct {
	MySQL    MySQLConfig
	Jwt      JwtConfig
	Code     CodeConfig
	Redis    RedisConfig
	Article  ArticleConfig
	RabbitMQ RabbitMQConfig
}

// GetConfig 获取配置实例
func GetConfig() Config {
	return Config{
		MySQL: MySQLConfig{
			DSN: MYSQL_USER + ":" + MYSQL_PASSWORD + "@tcp(" + MYSQL_HOST + ":" + MYSQL_PORT + ")/" + MYSQL_DB_NAME +
				"?charset=utf8mb4&parseTime=True&loc=Local",
		},
		Jwt: JwtConfig{
			SecretKey:                  "huancuilou",
			AccessTokenExpireDuration:  time.Hour * 7 * 24,
			RefreshTokenExpireDuration: time.Hour * 7 * 24,
		},
		Code: CodeConfig{
			ExpireDuration: time.Minute * 2,
		},
		Redis: RedisConfig{
			Addr:     "192.168.88.128:6379",
			Password: "1234",
			DB:       0,
		},
		Article: ArticleConfig{
			KindMap: map[string]bool{
				"生活服务": true,
				"医疗救助": true,
				"法律咨询": true,
				"心理咨询": true,
				"教育求助": true,
				"其他":   true,
			},
			UpdateLikesInterval: time.Hour,
		},
		RabbitMQ: RabbitMQConfig{
			DSN:     "amqp://" + MQ_USER + ":" + MQ_PASSWORD + "@" + MQ_HOST + ":" + MQ_PORT + "/",
			Durable: true,
		},
	}
}

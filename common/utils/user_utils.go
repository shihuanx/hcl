package utils

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"huancuilou/common/error_handler"
	"huancuilou/configs"
	"math/rand"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type JwtClaims struct {
	ID        string
	IsManager string
	jwt.RegisteredClaims
}

// ValidatePhoneNumber 验证手机号码
func ValidatePhoneNumber(phoneNumber string) bool {
	// 验证手机号码的正则表达式
	phoneRegex := regexp.MustCompile(`^1[3-9]\d{9}$`)

	// 使用正则表达式进行匹配
	return phoneRegex.MatchString(phoneNumber)
}

// MaskPhoneNumber 对手机号进行脱敏处理 方便记录日志
func MaskPhoneNumber(phoneNumber string) string {
	return phoneNumber[:3] + "****" + phoneNumber[7:]
}

// GenerateRandomUsername 生成随机用户名
func GenerateRandomUsername(length int) string {
	prefix := "user_"
	charset := configs.CHARACTER_SET
	rand.Seed(time.Now().UnixNano())
	var username strings.Builder
	for i := 0; i < (length - 5); i++ {
		username.WriteByte(charset[rand.Intn(len(charset))])
	}
	return prefix + username.String()
}

// GenAccessToken 生成 AccessToken
func GenAccessToken(id int, isManager int, accessTokenExpireDuration time.Duration, secret string) (string, error) {

	claims := JwtClaims{
		ID:        strconv.Itoa(id), // 自定义字段, userID
		IsManager: strconv.Itoa(isManager),
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(accessTokenExpireDuration)),
			Issuer:    "huancuilou", // 签发人
		},
	}
	// 使用指定的签名方法创建签名对象
	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// 使用指定的 secret 签名并获得完整的编码后的字符串 token
	return accessToken.SignedString([]byte(secret))
}

//// GenRefreshToken  生成 RefreshToken
//func GenRefreshToken(id int, isManager int, refreshTokenExpireDuration time.Duration, secret string) (string, error) {
//
//	claims := JwtClaims{
//		ID:        strconv.Itoa(id), // 自定义字段, userID
//		IsManager: strconv.Itoa(isManager),
//		RegisteredClaims: jwt.RegisteredClaims{
//			ExpiresAt: jwt.NewNumericDate(time.Now().Add(refreshTokenExpireDuration)),
//			Issuer:    "huancuilou", // 签发人
//		},
//	}
//	// 使用指定的签名方法创建签名对象
//	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
//	// 使用指定的 secret 签名并获得完整的编码后的字符串 token
//	return token.SignedString([]byte(secret))
//}
//
//// GenAccessTokenByRefreshToken 重新生成 AccessToken
//func GenAccessTokenByRefreshToken(refreshToken string) (map[string]string, error) {
//	//解析token
//	userIDStr, isManagerStr, err := ParseToken(refreshToken, configs.GetConfig().Jwt.SecretKey)
//	if err != nil {
//		return nil, err
//	}
//
//	claims := JwtClaims{
//		ID:        userIDStr, // 自定义字段, userID
//		IsManager: isManagerStr,
//		RegisteredClaims: jwt.RegisteredClaims{
//			ExpiresAt: jwt.NewNumericDate(time.Now().Add(configs.GetConfig().Jwt.AccessTokenExpireDuration)),
//			Issuer:    "huancuilou", // 签发人
//		},
//	}
//	// 使用指定的签名方法创建签名对象
//	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
//	// 使用指定的 secret 签名并获得完整的编码后的字符串 token
//	newAccessToken, err := accessToken.SignedString([]byte(configs.GetConfig().Jwt.SecretKey))
//	if err != nil {
//		return nil, err
//	}
//
//	return map[string]string{
//		"userIDStr":    userIDStr,
//		"isManagerStr": isManagerStr,
//		"accessToken":  newAccessToken,
//	}, nil
//}

// ValidateEmptyToken 验证请求头中指定类型的令牌是否为空
func ValidateEmptyToken(c *gin.Context, tokenKind string) (string, error) {
	switch tokenKind {
	case "accessToken":
		accessToken := c.Request.Header.Get("accessToken")
		if accessToken == "" {
			return "", errors.New("401:请求头中accessToken为空")
		}
		return accessToken, nil
	case "refreshToken":
		refreshToken := c.Request.Header.Get("refreshToken")
		if refreshToken == "" {
			return "", errors.New("401:请求头中refreshToken为空")
		}
		return refreshToken, nil
	default:
		return "", fmt.Errorf("500:tokenKind参数错误: %s", tokenKind)
	}
}

// ParseToken 解析 JWT
func ParseToken(tokenString string, secret string) (string, string, error) {
	// 解析 token
	token, err := jwt.ParseWithClaims(tokenString, &JwtClaims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法是否为预期的 HMAC 方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("401:ParseToken err: unexpected signing method")
		}
		return []byte(secret), nil
	})
	if err != nil {
		return "", "", fmt.Errorf("401:PraseToken err: %w", err)
	}
	// 对 token 进行校验
	if claims, ok := token.Claims.(*JwtClaims); ok && token.Valid {
		return claims.ID, claims.IsManager, nil
	}

	return "", "", errors.New("401:ParseToken err: token expired")
}

// 公共的 token 验证逻辑
func validateToken(c *gin.Context, roleCheck func(isManagerStr string) bool) {
	accessToken, err := ValidateEmptyToken(c, "accessToken")
	if err != nil {
		error_handler.HandleUserError(c, err)
		c.Abort()
		return
	}

	userIDStr, isManagerStr, err := ParseToken(accessToken, configs.GetConfig().Jwt.SecretKey)
	if err == nil {
		if !roleCheck(isManagerStr) {
			error_handler.HandleUserError(c, errors.New("403:权限不足"))
			c.Abort()
			return
		}
		userID, err := strconv.Atoi(userIDStr)
		if err != nil {
			error_handler.HandleUserError(c, errors.New("401:用户ID转换失败"))
			c.Abort()
			return
		}
		c.Set("userID", userID)
		c.Next()
		return
	}

	if !strings.Contains(err.Error(), "expired") {
		error_handler.HandleUserError(c, err)
		c.Abort()
		return
	}

	//refreshToken, err := ValidateEmptyToken(c, "refreshToken")
	//if err != nil {
	//	error_handler.HandleUserError(c, err)
	//	c.Abort()
	//	return
	//}

	//res, err := GenAccessTokenByRefreshToken(refreshToken)
	//if err != nil {
	//	error_handler.HandleUserError(c, fmt.Errorf("500: 通过refreshToken生成accessToken失败:%w", err))
	//	c.Abort()
	//	return
	//}
	//
	//userIDStr = res["userIDStr"]
	//isManagerStr = res["isManagerStr"]
	//newAccessToken := res["accessToken"]
	//c.Header("New-Access-Token", newAccessToken)

	if !roleCheck(isManagerStr) {
		error_handler.HandleUserError(c, errors.New("403:权限不足"))
		c.Abort()
		return
	}
	userID, err := strconv.Atoi(userIDStr)
	if err != nil {
		error_handler.HandleUserError(c, errors.New("401:用户ID转换失败"))
		c.Abort()
		return
	}
	c.Set("userID", userID)
	c.Next()
	return
}

// JwtInterceptor 基于JWT认证的中间件，会拦截不带token、token错误的请求
func JwtInterceptor() func(c *gin.Context) {
	return func(c *gin.Context) {
		validateToken(c, func(isManagerStr string) bool {
			return true
		})
	}
}

// AdminOnlyMiddleware 基于JWT认证的中间件，会拦截不带token、非管理员、token错误的请求
func AdminOnlyMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		validateToken(c, func(isManagerStr string) bool {
			return isManagerStr == "1" || isManagerStr == "2"
		})
	}
}

// SuperAdminOnlyMiddleware 基于JWT认证的中间件，会拦截不带token、非超级管理员、token错误的请求
func SuperAdminOnlyMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		validateToken(c, func(isManagerStr string) bool {
			return isManagerStr == "2"
		})
	}
}

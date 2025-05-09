package error_handler

import (
	"github.com/gin-gonic/gin"
	"log"
	"strings"
)

// ErrorType 定义错误类型结构
type ErrorType struct {
	Code    int
	Message string
}

// 预定义错误类型
var (
	ErrUnknown        = ErrorType{9999, "Unknown"}
	ErrBadRequest     = ErrorType{400, "Bad Request"}
	ErrInternalServer = ErrorType{500, "Internal Error"}
	ErrUnauthorized   = ErrorType{401, "Unauthorized"}
	ErrForbidden      = ErrorType{403, "Forbidden"}
)

// 错误信息与错误类型的映射
var errorMapping = map[string]ErrorType{
	"401": ErrUnauthorized,
	"500": ErrInternalServer,
	"400": ErrBadRequest,
	"403": ErrForbidden,
}

// HandleUserError 处理错误的通用函数，保留错误上下文
func HandleUserError(c *gin.Context, err error) {
	errMsg := err.Error()
	var errType ErrorType
	for key, value := range errorMapping {
		if strings.Contains(errMsg, key) {
			errType = value
			break
		}
	}
	if errType.Code == 0 {
		errType = ErrUnknown
	}
	log.Printf(err.Error())
	c.JSON(errType.Code, gin.H{
		"code": errType.Code,
		"msg":  errType.Message,
		"data": nil,
	})
}

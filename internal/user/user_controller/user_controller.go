package user_controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"huancuilou/common/error_handler"
	"huancuilou/common/utils"
	"huancuilou/configs"
	"huancuilou/internal/user/user_model"
	"huancuilou/internal/user/user_service"
	"huancuilou/response"
	"log"
	"net/http"
	"strconv"
	"time"
)

//处理用户操作的相关接口

//发送验证码接口
//登录接口
//社区管理员认证接口
//修改个人信息接口
//获取用户个人信息

// UserController 处理请求和返回响应
type UserController struct {
	userService *user_service.UserService
	jwtConfig   configs.JwtConfig
}

func NewUserController(userService *user_service.UserService, jwtConfig configs.JwtConfig) *UserController {
	return &UserController{
		userService: userService,
		jwtConfig:   jwtConfig,
	}
}

func (uc *UserController) SendCode(c *gin.Context) {
	phoneNumber := c.Param("phoneNumber")
	if !utils.ValidatePhoneNumber(phoneNumber) {
		error_handler.HandleUserError(c, fmt.Errorf("UserController.SendCode err: 400:手机号格式错误"))
	} else {
		if err := uc.userService.SendCode(phoneNumber); err != nil {
			error_handler.HandleUserError(c, fmt.Errorf("UserController.SendCode err: %w", err))
		} else {
			log.Printf("UserController.SendCode 成功发送验证码")
			c.JSON(http.StatusOK, response.SuccessWithoutData())
		}
	}
}

func (uc *UserController) Login(c *gin.Context) {
	var userCode user_model.UserCode
	if err := c.BindJSON(&userCode); err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("UserController.Login err: 400:将json数据绑定到结构体失败:%w", err))
		return
	}

	if !utils.ValidatePhoneNumber(userCode.PhoneNumber) {
		error_handler.HandleUserError(c, fmt.Errorf("UserController.Login err: 400:错误请求:手机号格式错误"))
		return
	}

	user, err := uc.userService.Login(&userCode)
	if err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("UserController.Login err: %w", err))
		return
	}

	accessToken, err := utils.GenAccessToken(user.ID, user.IsManager, uc.jwtConfig.AccessTokenExpireDuration, uc.jwtConfig.SecretKey)
	if err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("UserController.Login err: 500:生成accessToken失败:%w", err))
		return
	}
	//refreshToken, err := common.GenRefreshToken(user.ID, user.IsManager, uc.jwtConfig.RefreshTokenExpireDuration, uc.jwtConfig.SecretKey)
	//if err != nil {
	//	error_handler.HandleUserError(c, fmt.Errorf("UserController.Login err: 500:生成refreshToken失败:%w", err))
	//	return
	//}

	token := map[string]string{
		"accessToken": accessToken,
		//"refreshToken": refreshToken,
	}
	log.Printf("UserController.Login 成功登录")
	c.JSON(http.StatusOK, response.Success(token))
}

func (uc *UserController) GetUserInfo(c *gin.Context) {
	//类型断言
	userID := c.MustGet("userID").(int)
	user, err := uc.userService.GetUserByID(userID)
	if err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("UserController.GetUserInfo err: %w", err))
	} else {
		log.Printf("UserController.GetUserInfo 成功获取用户信息:%d", userID)
		c.JSON(http.StatusOK, response.Success(user))
	}
}

func (uc *UserController) AddAdministrator(c *gin.Context) {
	phoneNumber := c.Param("phoneNumber")
	if !utils.ValidatePhoneNumber(phoneNumber) {
		error_handler.HandleUserError(c, fmt.Errorf("UserController.SendCode err: 400:错误请求:手机号格式错误"))
	} else {
		if err := uc.userService.AddAdminByPhoneNumber(phoneNumber); err != nil {
			error_handler.HandleUserError(c, fmt.Errorf("UserController.AddAdministrator err: %w", err))
		} else {
			log.Printf("UserController.AddAdministrator 成功添加管理员")
			c.JSON(http.StatusOK, response.SuccessWithoutData())
		}
	}
}

func (uc *UserController) UpdateUserInfo(c *gin.Context) {
	//类型断言
	userID := c.MustGet("userID").(int)

	var user user_model.User
	if err := c.BindJSON(&user); err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("UserController.UpdateUserInfo err: 400:将json数据绑定到结构体失败:%w", err))
		return
	}
	if err := uc.userService.UpdateUserInfo(userID, &user); err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("UserController.UpdateUserInfo err: %w", err))
	} else {
		log.Printf("UserController.UpdateUserInfo 成功更新用户信息:%d", userID)
		c.JSON(http.StatusOK, response.SuccessWithoutData())
	}
}

func (uc *UserController) AddScore(c *gin.Context) {
	var scoreRecord user_model.ScoreRecord
	if err := c.BindJSON(&scoreRecord); err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("UserController.AddScore 错误请求:将json数据绑定到结构体失败:%w", err))
		return
	}
	if err := uc.userService.AddScore(&scoreRecord); err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("UserController.AddScore err:%w", err))
	} else {
		log.Printf("UserController.AddScore 成功添加评分")
		c.JSON(http.StatusOK, response.SuccessWithoutData())
	}
}

func (uc *UserController) AddPhoneRecord(c *gin.Context) {
	// 从表单数据中获取userID并转换为int
	userIDStr := c.PostForm("userID")
	managerID, err := strconv.Atoi(userIDStr)
	if err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("UserController.AddPhoneRecord err:无法将userID转换为int:%w", err))
		return
	}
	//管理员输入内容
	content := c.PostForm("content")
	if err := uc.userService.AddPhoneRecord(managerID, content); err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("UserController.AddPhoneRecord err:%w", err))
	} else {
		log.Printf("UserController.AddPhoneRecord 成功添加求助记录")
		c.JSON(http.StatusOK, response.SuccessWithoutData())
	}
}

func (uc *UserController) GetPhoneRecordByPhone(c *gin.Context) {
	phoneNumber := c.Param("phoneNumber")
	if !utils.ValidatePhoneNumber(phoneNumber) {
		error_handler.HandleUserError(c, fmt.Errorf("UserController.GetPhoneRecordByPhone err:错误请求:手机号格式错误"))
		return
	}
	phoneRecords, err := uc.userService.GetPhoneRecordByPhone(phoneNumber)
	if err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("UserController.GetPhoneRecordByPhone err:%w", err))
	} else {
		log.Printf("UserController.GetPhoneRecordByPhone 成功获取求助记录:%d", len(phoneRecords))
		c.JSON(http.StatusOK, response.Success(phoneRecords))
	}
}

func (uc *UserController) AddFollows(c *gin.Context) {
	userID := c.MustGet("userID").(int)
	followerID, err := strconv.Atoi(c.Param("followerID"))
	if err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("UserController.AddFollows err: 400: 无法将followerID转换为int:%w", err))
		return
	}
	userFollow := &user_model.UserFollow{UserID: userID, FollowID: followerID, CreateAt: time.Now()}
	err = uc.userService.AddFollows(userFollow)
	if err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("UserController.AddFollows err: 500: %w", err))
		return
	}
	c.JSON(http.StatusOK, response.SuccessWithoutData())
}

func (uc *UserController) GetFollows(c *gin.Context) {
	userID := c.MustGet("userID").(int)
	follows, err := uc.userService.GetFollows(userID)
	if err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("UserController.GetFollows err: 500: %w", err))
		return
	}
	c.JSON(http.StatusOK, response.Success(follows))
}

func (uc *UserController) GetOtherUserInfo(c *gin.Context) {
	userID := c.Param("userID")
	userIDInt, err := strconv.Atoi(userID)
	if err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("UserController.GetOtherUserInfo err: 400: 无法将userID转换为int:%w", err))
		return
	}
	user, err := uc.userService.GetUserByID(userIDInt)
	if err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("UserController.GetOtherUserInfo err: 500: %w", err))
		return
	}
	c.JSON(http.StatusOK, response.Success(user))
}

func (uc *UserController) RemoveFollows(c *gin.Context) {
	userID := c.MustGet("userID").(int)
	followerID, err := strconv.Atoi(c.Param("followID"))
	if err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("UserController.RemoveFollows err: 400: 无法将followerID转换为int:%w", err))
		return
	}
	err = uc.userService.RemoveFollows(userID, followerID)
	if err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("UserController.RemoveFollows err: 500: %w", err))
		return
	}
	c.JSON(http.StatusOK, response.SuccessWithoutData())
}

func (uc *UserController) GetCommonFollows(c *gin.Context) {
	userID := c.MustGet("userID").(int)
	otherUserID, err := strconv.Atoi(c.Param("otherUserID"))
	if err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("UserController.CommonFollows err: 400: 获取用户ID失败:%w", err))
		return
	}
	var users []*user_model.User
	users, err = uc.userService.GetCommonFollows(userID, otherUserID)
	if err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("UserController.CommonFollows err: 500: %w", err))
		return
	}
	c.JSON(http.StatusOK, response.Success(users))
}

func (uc *UserController) AddLikes(c *gin.Context) {
	userID, err := strconv.Atoi(c.Param("userID"))
	if err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("UserController.AddLikes err: 400: 无法将userID转换为int:%w", err))
		return
	}
	err = uc.userService.AddLikes(userID)
	if err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("UserController.AddLikes err: 500: %w", err))
		return
	}
	c.JSON(http.StatusOK, response.SuccessWithoutData())
}

func (uc *UserController) GetLikesRank(c *gin.Context) {
	userID := c.MustGet("userID").(int)
	likesRank, err := uc.userService.GetLikesRank(userID)
	if err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("UserController.GetLikesRank err: 500: %w", err))
		return
	}
	c.JSON(http.StatusOK, response.Success(likesRank))
}

func (uc *UserController) AddItem(c *gin.Context) {
	var communityItem user_model.CommunityItem
	if err := c.BindJSON(&communityItem); err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("UserController.AddItem err: 400: 将json数据绑定到结构体失败:%w", err))
		return
	}
	if err := uc.userService.AddItem(&communityItem); err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("UserController.AddItem err: 500: %w", err))
		return
	}
	c.JSON(http.StatusOK, response.SuccessWithoutData())
}

func (uc *UserController) GetAllItems(c *gin.Context) {
	items, err := uc.userService.GetAllItems()
	if err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("UserController.GetAllItems err: 500: %w", err))
		return
	}
	c.JSON(http.StatusOK, response.Success(items))
}

func (uc *UserController) ChooseItem(c *gin.Context) {
	userID := c.MustGet("userID").(int)
	itemID, err := strconv.Atoi(c.Param("itemID"))
	if err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("UserController.ChooseItem err: 400: 无法将itemID转换为int:%w", err))
	}
	if err := uc.userService.ChooseItemPublisher(userID, itemID); err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("UserController.ChooseItem err: 500: %w", err))
		return
	}
	c.JSON(http.StatusOK, response.SuccessWithoutData())
}

func (uc *UserController) AddChooseItemConsumer(c *gin.Context) {
	beginStr := c.Query("begin")
	endStr := c.Query("end")

	loc, err := time.LoadLocation("Local")
	if err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("UserController.AddChooseItemConsumer err: 400: %w", err))
		return
	}

	begin, err := time.ParseInLocation("2006-01-02 15:04:05", beginStr, loc)
	if err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("UserController.AddChooseItemConsumer err: 400: 无法将begin转换为time.Time:%w", err))
		return
	}

	end, err := time.ParseInLocation("2006-01-02 15:04:05", endStr, loc)
	if err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("UserController.AddChooseItemConsumer err: 400: 无法将end转换为time.Time:%w", err))
		return
	}
	if err := uc.userService.AddChooseItemConsumer(begin, end); err != nil {
		error_handler.HandleUserError(c, fmt.Errorf("UserController.AddChooseItemConsumer err: 500: %w", err))
		return
	}
	c.JSON(http.StatusOK, response.SuccessWithoutData())
}

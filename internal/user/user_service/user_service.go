package user_service

import (
	"errors"
	"fmt"
	"github.com/streadway/amqp"
	"gorm.io/gorm"
	"huancuilou/common/utils"
	"huancuilou/configs"
	"huancuilou/internal/user/user_model"
	"huancuilou/internal/user/user_repository"
	"log"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

// UserService 处理业务逻辑
type UserService struct {
	userRepository      *user_repository.UserRepository
	config              *configs.Config
	userMdbRepository   *user_repository.UserMemoryDBRepository
	userCacheRepository *user_repository.UserCacheRepository
}

func NewUserService(userRepository *user_repository.UserRepository, config *configs.Config, userMemoryDBRepository *user_repository.UserMemoryDBRepository, userCacheRepository *user_repository.UserCacheRepository) *UserService {
	return &UserService{
		userRepository:      userRepository,
		config:              config,
		userMdbRepository:   userMemoryDBRepository,
		userCacheRepository: userCacheRepository,
	}
}

// SendCode 发送验证码
func (us *UserService) SendCode(phoneNumber string) error {
	// 使用当前时间作为随机数种子
	rand.Seed(time.Now().UnixNano())
	var code string
	for i := 0; i < 6; i++ {
		// 生成 0 到 9 的随机数字
		digit := rand.Intn(10)
		code += fmt.Sprintf("%d", digit)
	}
	userCode := &user_model.UserCode{
		Code:        code,
		PhoneNumber: phoneNumber,
	}
	maskPhone := utils.MaskPhoneNumber(phoneNumber)
	// 将验证码插入内存或更新验证码
	if err := us.userMdbRepository.AddCode(userCode, us.config.Code.ExpireDuration); err != nil {
		return fmt.Errorf("UserService.SendCode err: 500: 向内存插入或更新验证码错误: 手机号: %s,err: %w", maskPhone, err)
	}

	// 启动 goroutine 异步调用外部接口发送验证码
	go func() {
		maxRetries := 3
		for i := 0; i < maxRetries; i++ {
			log.Printf(code)
			return
		}
		log.Printf("UserService.SendCode err:500: 达到最大重试次数，发送验证码失败，手机号: %s", maskPhone)
	}()

	return nil
}

// Login 一键登录注册
func (us *UserService) Login(userCode *user_model.UserCode) (*user_model.User, error) {
	//比对验证码
	err := us.userMdbRepository.ValidateCode(userCode.Code, userCode.PhoneNumber)
	if err != nil {
		return nil, fmt.Errorf("UserService.Login err: 401: 验证码验证失败: %w", err)
	}
	//检查用户是否存在，存在则登录，不存在则注册
	exists, err := us.userRepository.GetUserByPhoneNumber(userCode.PhoneNumber)
	if err != nil {
		return nil, fmt.Errorf("UserService.Login err: 500: 通过手机号查找用户错误: %w", err)
	}
	tx := us.userRepository.DB.Begin()
	if tx.Error != nil {
		return nil, fmt.Errorf("UserService.Login err: 500:%w", err)
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	if exists == nil {
		maxRetries := 5
		for i := 0; i < maxRetries; i++ {
			userDB := user_model.User{
				PhoneNumber: userCode.PhoneNumber,
				UserName:    utils.GenerateRandomUsername(20),
				Biography:   "添加个人简介，让大家更好地认识你~",
				Likes:       0,
			}
			user, err := us.userRepository.AddUser(tx, &userDB)
			if err != nil {
				if strings.Contains(err.Error(), "Error 1062 (23000): Duplicate entry") {
					log.Printf("用户名 %s 已存在，尝试重新生成...\n", userDB.UserName)
					continue
				}
				tx.Rollback()
				return nil, fmt.Errorf("UserService.Login err: 500:添加用户错误:%w", err)
			}
			err = us.userCacheRepository.AddLikes(user.ID, 0)
			if err != nil {
				tx.Rollback()
				return nil, err
			}
			return user, nil
		}
		tx.Rollback()
		return nil, errors.New("UserService.Login err: 500:达到最大重试次数，无法插入唯一用户名")
	}
	if err := tx.Commit().Error; err != nil {
		log.Printf("事务提交失败: %v", err)
		return nil, fmt.Errorf("UserService.Login err:%w", err)
	}
	return exists, nil
}

func (us *UserService) GetUserByID(userID int) (*user_model.User, error) {
	var user *user_model.User
	user, err := us.userRepository.GetUserByUserID(userID)
	if err != nil {
		return nil, fmt.Errorf("UserService.GetUserByID err: 500:通过ID查找用户错误:%w", err)
	}
	return user, nil
}

func (us *UserService) AddAdminByPhoneNumber(phoneNumber string) error {
	user, err := us.userRepository.GetUserByPhoneNumber(phoneNumber)
	if err != nil {
		return fmt.Errorf("UserService.AddAdminByPhoneNumber err: 500:通过手机号查找用户错误:%w", err)
	}
	if user.IsManager == 1 {
		return fmt.Errorf("UserService.AddAdminByPhoneNumber err: 400:该用户已经是管理员")
	}
	if err = us.userRepository.AddAdminByPhoneNumber(phoneNumber); err != nil {
		return fmt.Errorf("UserService.AddAdminByPhoneNumber err: 500:更新管理员出错:%w", err)
	}
	return nil
}

func (us *UserService) UpdateUserInfo(userID int, user *user_model.User) error {
	if err := us.userRepository.UpdateUserInfo(userID, user); err != nil {
		return fmt.Errorf("UserService.UpdateUserInfo err: 500:更新用户信息出错:%w", err)
	}
	return nil
}

func (us *UserService) AddScore(scoreRecord *user_model.ScoreRecord) error {
	if err := us.userRepository.AddScore(scoreRecord); err != nil {
		return fmt.Errorf("UserService.AddScore 数据库操作错误:添加积分记录错误:%w", err)
	}
	return nil
}

func (us *UserService) AddPhoneRecord(managerID int, content string) error {
	// 查找最新的评分记录
	var latestScoreRecord user_model.ScoreRecord
	result := us.userRepository.DB.Where("manager_id = ?", managerID).Order("created_at desc").First(&latestScoreRecord)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return fmt.Errorf("UserService.AddPhoneRecord 错误请求:未找到评分记录")
		}
		return fmt.Errorf("UserService.AddPhoneRecord 数据库操作错误:查找评分记录错误:%w", result.Error)
	}

	// 查找用户电话号码
	user, err := us.userRepository.GetUserByUserID(latestScoreRecord.UserID)
	if err != nil {
		return fmt.Errorf("UserService.AddPhoneRecord 数据库操作错误:查找用户错误:%w", err)
	}
	if user == nil {
		return fmt.Errorf("UserService.AddPhoneRecord 错误请求:用户不存在")
	}

	// 创建电话记录，Content字段让管理员自己输入
	phoneRecord := &user_model.PhoneRecord{
		UserID:       latestScoreRecord.UserID,
		ManagerID:    latestScoreRecord.ManagerID,
		Satisfaction: latestScoreRecord.Score,
		UserPhone:    user.PhoneNumber,
		CreatedAt:    time.Now(),
		Content:      content,
	}

	// 添加电话记录到数据库
	if err := us.userRepository.AddPhoneRecord(phoneRecord); err != nil {
		return fmt.Errorf("UserService.AddPhoneRecord 数据库操作错误:添加电话记录错误:%w", err)
	}
	return nil
}

func (us *UserService) GetPhoneRecordByPhone(phoneNumber string) ([]user_model.PhoneRecord, error) {
	phoneRecords, err := us.userRepository.GetPhoneRecordByPhone(phoneNumber)
	if err != nil {
		return nil, fmt.Errorf("UserService.GetPhoneRecordByPhone err:%w", err)
	}
	return phoneRecords, nil
}

func (us *UserService) AddFollows(userFollow *user_model.UserFollow) error {
	tx := us.userRepository.DB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("UserService.AddFollows err:%w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("事务已回滚")
		}
	}()
	if err := us.userRepository.AddFollows(tx, userFollow); err != nil {
		tx.Rollback()
		return fmt.Errorf("UserService.AddFollows err:%w", err)
	}
	if err := us.userCacheRepository.AddFollows(userFollow); err != nil {
		tx.Rollback()
		return fmt.Errorf("UserService.AddFollows err:%w", err)
	}
	if err := us.userCacheRepository.AddFans(userFollow); err != nil {
		tx.Rollback()
		return fmt.Errorf("UserService.AddFollows err:%w", err)
	}
	if err := tx.Commit().Error; err != nil {
		log.Printf("事务提交失败: %v", err)
		return fmt.Errorf("UserService.AddFollows err:%w", err)
	}
	return nil
}

func (us *UserService) GetFollows(userID int) ([]*user_model.User, error) {
	var follows []*user_model.User
	follows, err := us.userRepository.GetFollows(userID)
	if err != nil {
		return nil, fmt.Errorf("UserService.GetFollows err:%w", err)
	}
	return follows, nil
}

func (us *UserService) RemoveFollows(userID int, followID int) error {
	tx := us.userRepository.DB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("UserService.RemoveFollows err:%w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("事务已回滚")
		}
	}()
	if err := us.userRepository.RemoveFollows(tx, userID, followID); err != nil {
		tx.Rollback()
		return fmt.Errorf("UserService.RemoveFollows err:%w", err)
	}
	if err := us.userCacheRepository.RemoveFollows(userID, followID); err != nil {
		tx.Rollback()
		return fmt.Errorf("UserService.RemoveFollows err:%w", err)
	}
	if err := us.userCacheRepository.RemoveFans(userID, followID); err != nil {
		tx.Rollback()
		return fmt.Errorf("UserService.RemoveFollows err:%w", err)
	}
	if err := tx.Commit().Error; err != nil {
		log.Printf("事务提交失败: %v", err)
		return fmt.Errorf("UserService.RemoveFollows err:%w", err)
	}
	return nil
}

func (us *UserService) AddLikes(userID int) error {
	tx := us.userRepository.DB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("UserService.AddLikes err:%w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
			log.Printf("事务已回滚")
		}
	}()
	if err := us.userCacheRepository.AddLikes(userID, 1); err != nil {
		tx.Rollback()
		return fmt.Errorf("UserService.AddLikes err:%w", err)
	}
	if err := us.userRepository.AddLikes(tx, userID); err != nil {
		tx.Rollback()
		return fmt.Errorf("UserService.AddLikes err:%w", err)
	}
	if err := tx.Commit().Error; err != nil {
		log.Printf("事务提交失败: %v", err)
		return fmt.Errorf("UserService.AddLikes err:%w", err)
	}
	return nil
}

func (us *UserService) GetCommonFollows(userID int, otherUserID int) ([]*user_model.User, error) {
	var users []*user_model.User
	userIDs, err := us.userCacheRepository.GetCommonFollows(userID, otherUserID)
	if err != nil {
		return nil, fmt.Errorf("UserService.GetCommonFollows err:%w", err)
	}
	for _, userID = range userIDs {
		user, err := us.userRepository.GetUserByUserID(userID)
		if err != nil {
			return nil, fmt.Errorf("UserService.GetCommonFollows err:%w", err)
		}
		if user == nil {
			return nil, fmt.Errorf("UserService.GetCommonFollows err:未找到用户")
		}
		users = append(users, user)
	}
	return users, nil
}

func (us *UserService) GetLikesRank(userID int) ([]*user_model.UserLikeRank, error) {
	userLikeRanks, err := us.userCacheRepository.GetLikesRank(userID)
	if err != nil {
		return nil, fmt.Errorf("UserService.GetLikesRank err:%w", err)
	}
	return userLikeRanks, nil
}

func (us *UserService) AddItem(item *user_model.CommunityItem) error {
	item.Remain = item.Capacity
	if err := us.userRepository.AddItem(item); err != nil {
		return fmt.Errorf("UserService.AddItem err:%w", err)
	}

	newItem, err := us.userRepository.GetItemByID(item.ID)
	if err != nil {
		return fmt.Errorf("UserService.AddItem err:%w", err)
	}

	if err := us.userCacheRepository.AddItem(newItem); err != nil {
		return fmt.Errorf("UserService.AddItem err:%w", err)
	}

	return nil
}

func (us *UserService) GetAllItems() ([]*user_model.CommunityItem, error) {
	items, err := us.userCacheRepository.GetAllItems()
	if err != nil {
		return nil, fmt.Errorf("UserService.GetAllItems err:%w", err)
	}
	return items, nil
}

func (us *UserService) ChooseItem(userID int, itemID int) error {
	tx := us.userRepository.DB.Begin()
	if tx.Error != nil {
		return fmt.Errorf("UserService.ChooseItem err:%w", tx.Error)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()
	userItem := &user_model.UserItem{
		UserID:   userID,
		ItemID:   itemID,
		CreateAt: time.Now(),
	}

	if err := us.userRepository.ChooseItem(tx, userItem); err != nil {
		tx.Rollback()
		return fmt.Errorf("UserService.ChooseItem err:%w", err)
	}
	if err := us.userRepository.UpdateItemInfo(tx, itemID); err != nil {
		tx.Rollback()
		return fmt.Errorf("UserService.ChooseItem err:%w", err)
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("UserService.ChooseItem err:%w", err)
	}
	return nil
}

func (us *UserService) ChooseItemPublisher(userID int, itemID int) error {
	if err := us.userCacheRepository.ChooseItem(userID, itemID); err != nil {
		return fmt.Errorf("UserService.ChooseItemPublisher err:%w", err)
	}

	conn, err := amqp.Dial(us.config.RabbitMQ.DSN)
	if err != nil {
		return fmt.Errorf("UserService.ChooseItemPublisher err:%w", err)
	}
	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("UserService.ChooseItemPublisher err:%w", err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"hcl_user_choose_item",
		us.config.RabbitMQ.Durable,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("UserService.ChooseItemPublisher err:%w", err)
	}

	// 构造消息内容
	msgContent := fmt.Sprintf("%d,%d", userID, itemID)
	msg := amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(msgContent),
	}

	// 将消息发送到队列
	if err = ch.Publish(
		"",     // 交换器名称
		q.Name, // 路由键（队列名称）
		false,  // 是否强制
		false,  // 是否立即
		msg,
	); err != nil {
		return fmt.Errorf("UserService.ChooseItemPublisher err:%w", err)
	}

	log.Printf("UserService.ChooseItemPublisher userID:%d, itemID:%d", userID, itemID)

	return nil
}

func (us *UserService) AddChooseItemConsumer(begin time.Time, end time.Time) error {
	if time.Now().After(end) {
		return fmt.Errorf("UserService.AddChooseItemConsumer err: 400: 结束时间不能小于当前时间")
	}
	if begin == end {
		return fmt.Errorf("UserService.AddChooseItemConsumer err: 400: 开始时间不能等于结束时间")
	}

	go func() {
		if begin.After(time.Now()) {
			time.Sleep(time.Until(begin))
		}

		conn, err := amqp.Dial(us.config.RabbitMQ.DSN)
		if err != nil {
			log.Printf("UserService.ChooseItemConsumer err: %v", err)
		}
		defer func(conn *amqp.Connection) {
			err := conn.Close()
			if err != nil {
				log.Printf("UserService.ChooseItemConsumer conn.Close err: %v", err)
			}
		}(conn)
		ch, err := conn.Channel()
		if err != nil {
			log.Printf("UserService.ChooseItemConsumer err: %v", err)
		}
		defer func(ch *amqp.Channel) {
			err := ch.Close()
			if err != nil {
				log.Printf("UserService.ChooseItemConsumer ch.Close err: %v", err)
			}
		}(ch)

		q, err := ch.QueueDeclare(
			"hcl_user_choose_item",
			us.config.RabbitMQ.Durable,
			false,
			false,
			false,
			nil,
		)
		if err != nil {
			log.Printf("UserService.ChooseItemConsumer err: %v", err)
		}

		// 获取接收消息的Delivery通道
		messages, err := ch.Consume(
			q.Name, // queue
			"",     // consumer
			true,   // auto-ack
			false,  // exclusive
			false,  // no-local
			false,  // no-wait
			nil,    // args
		)
		if err != nil {
			log.Printf("Failed to register a consumer: %v", err)
		}

		stopCh := make(chan struct{})

		go func() {
			for {
				select {
				case msg, ok := <-messages:
					if !ok {
						log.Println("消息通道已关闭")
						return
					}
					log.Printf("收到消息: %s", msg.Body)
					stringParts := strings.Split(string(msg.Body), ",")
					if len(stringParts) < 2 {
						log.Printf("无效消息格式: %s", msg.Body)
						continue
					}

					userID, err := strconv.Atoi(stringParts[0])
					if err != nil {
						log.Printf("解析用户ID失败: %v", err)
						continue
					}

					itemID, err := strconv.Atoi(stringParts[1])
					if err != nil {
						log.Printf("解析物品ID失败: %v", err)
						continue
					}

					if err := us.ChooseItem(userID, itemID); err != nil {
						log.Printf("选课失败: 用户=%d, 物品=%d, 错误=%v", userID, itemID, err)
					} else {
						log.Printf("选课成功: 用户=%d, 物品=%d", userID, itemID)
					}
				case <-stopCh:
					log.Println("收到停止信号，退出消息处理")
					return
				}
			}
		}()

		log.Printf("开启消费者 监听中……")

		duration := end.Sub(time.Now())
		time.Sleep(duration)
		close(stopCh)

		log.Printf("关闭消费者")
	}()
	return nil
}

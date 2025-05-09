package user_repository

import (
	"errors"
	"huancuilou/internal/user/user_model"
	"time"
)

type UserMemoryDBRepository struct {
	codeMap  map[string]*user_model.UserCode
	timerMap map[string]*time.Timer
	//locks    sync.Map
}

func NewUserMemoryDBRepository() *UserMemoryDBRepository {
	return &UserMemoryDBRepository{
		codeMap:  make(map[string]*user_model.UserCode),
		timerMap: make(map[string]*time.Timer),
	}
}

//// 获取行锁 感觉没啥必要 浪费性能
//func (um *UserMemoryDBRepository) getLock(phoneNumber string) *sync.Mutex {
//	lock, loaded := um.locks.Load(phoneNumber)
//	if !loaded {
//		newLock := &sync.Mutex{}
//		lock, _ = um.locks.LoadOrStore(phoneNumber, newLock)
//	}
//	return lock.(*sync.Mutex)
//}

func (um *UserMemoryDBRepository) AddCode(userCode *user_model.UserCode, interval time.Duration) error {
	//lock := um.getLock(userCode.PhoneNumber)
	//lock.Lock()
	//defer lock.Unlock()

	// 检查是否已经存在该手机号的定时器，如果存在则取消它
	if timer, exists := um.timerMap[userCode.PhoneNumber]; exists {
		if timer.Stop() {
			delete(um.timerMap, userCode.PhoneNumber)
		}
	}

	// 更新 codeMap 中的验证码
	um.codeMap[userCode.PhoneNumber] = userCode

	// 创建一个新的定时器
	timer := time.AfterFunc(interval, func() {
		//lock := um.getLock(userCode.PhoneNumber)
		//lock.Lock()
		//defer lock.Unlock()
		delete(um.codeMap, userCode.PhoneNumber)
		delete(um.timerMap, userCode.PhoneNumber)
	})

	// 将新的定时器保存到 timerMap 中
	um.timerMap[userCode.PhoneNumber] = timer

	return nil
}

func (um *UserMemoryDBRepository) ValidateCode(code string, phoneNumber string) error {
	//lock := um.getLock(phoneNumber)
	//lock.Lock()
	//defer lock.Unlock()
	if userCode, exists := um.codeMap[phoneNumber]; exists {
		if userCode.Code == code {
			delete(um.codeMap, phoneNumber)
			delete(um.timerMap, phoneNumber)
			return nil
		}
		return errors.New("验证码错误")
	}
	return errors.New("验证码不存在或已过期")
}

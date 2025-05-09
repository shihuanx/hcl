package user_repository

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"huancuilou/internal/user/user_model"
)

// UserRepository 用户数据访问层
type UserRepository struct {
	DB *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		DB: db,
	}
}

func (ur *UserRepository) GetUserByUserID(userID int) (*user_model.User, error) {
	var user user_model.User
	result := ur.DB.Take(&user, "ID = ?", userID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("UserRepository.GetUserByID err:%w", result.Error)
	}
	return &user, nil
}

func (ur *UserRepository) AddUser(tx *gorm.DB, user *user_model.User) (*user_model.User, error) {
	result := tx.Create(&user)
	if result.Error != nil {
		return nil, fmt.Errorf("UserRepository.AddUser err:%w", result.Error)
	}
	return user, nil
}

func (ur *UserRepository) GetUserByPhoneNumber(phoneNumber string) (*user_model.User, error) {
	var user user_model.User
	result := ur.DB.Take(&user, "phone_number = ?", phoneNumber)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("UserRepository.GetUserByID err:%w", result.Error)
	}
	return &user, nil
}

func (ur *UserRepository) AddAdminByPhoneNumber(phoneNumber string) error {
	result := ur.DB.Model(&user_model.User{}).Where("phone_number = ?", phoneNumber).Update("is_manager", 1)
	if result.Error != nil {
		return fmt.Errorf("UserRepository.AddAdminByPhoneNumber err:%w", result.Error)
	}
	return nil
}

func (ur *UserRepository) UpdateUserInfo(id int, user *user_model.User) error {
	result := ur.DB.Model(&user_model.User{}).Where("id = ?", id).Updates(user)
	if result.Error != nil {
		return fmt.Errorf("UserRepository.UpdateUserInfo err:%w", result.Error)
	}
	return nil
}

func (ur *UserRepository) AddScore(scoreRecord *user_model.ScoreRecord) error {
	result := ur.DB.Create(&scoreRecord)
	if result.Error != nil {
		return fmt.Errorf("UserRepository.AddScore err:%w", result.Error)
	}
	return nil
}
func (ur *UserRepository) AddPhoneRecord(phoneRecord *user_model.PhoneRecord) error {
	result := ur.DB.Create(&phoneRecord)
	if result.Error != nil {
		return fmt.Errorf("UserRepository.AddPhoneRecord err:%w", result.Error)
	}
	return nil
}

func (ur *UserRepository) GetPhoneRecordByPhone(phoneNumber string) ([]user_model.PhoneRecord, error) {
	var phoneRecords []user_model.PhoneRecord
	result := ur.DB.Where("user_phone = ?", phoneNumber).Find(&phoneRecords)
	if result.Error != nil {
		return nil, fmt.Errorf("UserRepository.GetPhoneRecordByPhone err:%w", result.Error)
	}
	return phoneRecords, nil
}

func (ur *UserRepository) AddFollows(tx *gorm.DB, follow *user_model.UserFollow) error {
	result := tx.Create(&follow)
	if result.Error != nil {
		return fmt.Errorf("UserRepository.AddFollows err:%w", result.Error)
	}
	return nil
}

func (ur *UserRepository) GetFollows(userID int) ([]*user_model.User, error) {
	var follows []*user_model.User
	result := ur.DB.Raw("select * from user where id in (select follow_id from user_follow where user_id = ?)", userID).Scan(&follows)
	if result.Error != nil {
		return nil, fmt.Errorf("UserRepository.GetFollows err:%w", result.Error)
	}
	return follows, nil
}

func (ur *UserRepository) RemoveFollows(tx *gorm.DB, userID int, followID int) error {
	result := tx.Delete(&user_model.UserFollow{}, "user_id = ? and follow_id = ?", userID, followID)
	if result.Error != nil {
		return fmt.Errorf("UserRepository.RemoveFollows err:%w", result.Error)
	}
	return nil
}

func (ur *UserRepository) AddLikes(tx *gorm.DB, userID int) error {
	result := tx.Model(&user_model.User{}).Where("id = ?", userID).Update("likes", gorm.Expr("likes + ?", 1))
	if result.Error != nil {
		return fmt.Errorf("UserRepository.AddLikes err:%w", result.Error)
	}
	return nil
}

func (ur *UserRepository) AddItem(item *user_model.CommunityItem) error {
	result := ur.DB.Create(&item)
	if result.Error != nil {
		return fmt.Errorf("UserRepository.AddItem err:%w", result.Error)
	}
	return nil
}

func (ur *UserRepository) GetAllItems() ([]*user_model.CommunityItem, error) {
	var items []*user_model.CommunityItem
	result := ur.DB.Find(&items)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("UserRepository.GetAllItems err:%w", result.Error)
	}
	return items, nil
}

func (ur *UserRepository) GetItemByID(id int) (*user_model.CommunityItem, error) {
	var item *user_model.CommunityItem
	result := ur.DB.Where("id = ?", id).First(&item)
	if result.Error != nil {
		return nil, fmt.Errorf("UserRepository.GetItemsByID err:%w", result.Error)
	}

	return item, nil
}

func (ur *UserRepository) ChooseItem(tx *gorm.DB, userItem *user_model.UserItem) error {
	result := tx.Create(userItem)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (ur *UserRepository) UpdateItemInfo(tx *gorm.DB, id int) error {
	result := tx.Model(&user_model.CommunityItem{}).Where("id = ?", id).Update("remain", gorm.Expr("remain-1"))
	if result.Error != nil {
		return result.Error
	}
	return nil
}

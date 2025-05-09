package user_model

import "time"

// User 用户表
type User struct {
	ID          int       `json:"id"`
	UserName    string    `json:"user_name" gorm:"column:username" `
	PhoneNumber string    `json:"phone_number"`
	IsManager   int       `json:"is_manager"`
	CreatedAt   time.Time `json:"created_at"`
	AvatarUrl   string    `json:"avatar_url"`
	Biography   string    `json:"biography"`
	Likes       int       `json:"likes"`
}

// TableName 自定义表名
func (User) TableName() string {
	return "user"
}

package user_model

import "time"

type UserFollow struct {
	ID       int
	UserID   int
	FollowID int
	CreateAt time.Time
}

func (UserFollow) TableName() string {
	return "user_follow"
}

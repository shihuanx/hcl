package user_model

import "time"

type UserItem struct {
	ID       int       `json:"id"`
	UserID   int       `json:"user_id"`
	ItemID   int       `json:"item_id"`
	CreateAt time.Time `json:"create_at"`
}

func (UserItem) TableName() string {
	return "user_item"
}

package user_model

import "time"

type PhoneRecord struct {
	ID           int       `json:"id"`
	UserID       int       `json:"user_id"`
	ManagerID    int       `json:"manager_id"`
	Content      string    `json:"content"`
	CreatedAt    time.Time `json:"created_at"`
	Satisfaction int       `json:"satisfaction"`
	UserPhone    string    `json:"user_phone"`
}

func (PhoneRecord) TableName() string {
	return "phone_record"
}

package user_model

import "time"

// ScoreRecord 评分记录表
type ScoreRecord struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	ManagerID int       `json:"manager_id"`
	Score     int       `json:"score"`
	CreatedAt time.Time `json:"created_at"`
}

// TableName 自定义表名
func (ScoreRecord) TableName() string {
	return "score_record"
}

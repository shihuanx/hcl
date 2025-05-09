package article_model

import "time"

type Article struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Content   string    `json:"content"`
	ManagerID int       `json:"managerID"`
	CreateAt  time.Time `json:"createAt"`
	Kind      string    `json:"kind"`
	Like      int       `json:"like"`
}

func (Article) TableName() string {
	return "article"
}

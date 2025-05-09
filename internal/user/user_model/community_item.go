package user_model

import "time"

type CommunityItem struct {
	ID       int
	Name     string
	Capacity int
	Remain   int
	Price    int
	Begin    time.Time
}

func (CommunityItem) TableName() string {
	return "community_item"
}

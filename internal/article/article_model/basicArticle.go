package article_model

type BasicArticle struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Kind      string `json:"kind"`
	Like      int    `json:"like"`
	ManagerID int    `json:"manager_id"`
}

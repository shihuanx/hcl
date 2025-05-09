package user_model

// UserCode 用户验证码
type UserCode struct {
	PhoneNumber string `json:"phone_number"`
	Code        string `json:"code"`
}

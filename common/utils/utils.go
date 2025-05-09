package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

//小程序需要接入微信安全审查接口 用于审查文字和图片合法合规

type SecurityCheckRequest struct {
	AccessToken string `json:"-"`
	Content     string `json:"content"`
	Scene       int    `json:"scene"`
	Version     int    `json:"version"`
	OpenID      string `json:"openid"`
	Title       string `json:"title,omitempty"`
	Nickname    string `json:"nickname,omitempty"`
	Signature   string `json:"signature,omitempty"`
}

type SecurityCheckResponse struct {
	Errcode int                   `json:"errcode"`
	Errmsg  string                `json:"errmsg"`
	Result  *SecurityCheckResult  `json:"result,omitempty"`
	Detail  []SecurityCheckDetail `json:"detail,omitempty"`
	TraceID string                `json:"trace_id"`
}

type SecurityCheckResult struct {
	Suggest string `json:"suggest"`
	Label   int    `json:"label"`
}

type SecurityCheckDetail struct {
	Strategy string `json:"strategy"`
	Errcode  int    `json:"errcode"`
	Suggest  string `json:"suggest"`
	Label    int    `json:"label"`
	Prob     int    `json:"prob"`
	Keyword  string `json:"keyword,omitempty"`
}

// CheckContentSecurity 安全审查函数
func CheckContentSecurity(content, openid string, scene int, title string) bool {
	// 获取当前有效的 access_token
	accessToken, err := GetAccessToken()
	if err != nil {
		log.Printf("获取 access_token 失败: %v\n", err)
		return false
	}

	// 构建请求体
	requestBody := SecurityCheckRequest{
		AccessToken: accessToken,
		Content:     content,
		Scene:       scene,
		Version:     2,
		OpenID:      openid,
		Title:       title,
	}

	// 将请求体转换为 JSON
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		log.Printf("JSON 编码失败: %v\n", err)
		return false
	}

	// 构建请求 URL
	url := fmt.Sprintf("https://api.weixin.qq.com/wxa/msg_sec_check?access_token=%s", accessToken)

	// 发送 POST 请求
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Printf("发送请求失败: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("读取响应失败: %v\n", err)
		return false
	}

	// 解析响应
	var response SecurityCheckResponse
	if err := json.Unmarshal(respBody, &response); err != nil {
		log.Printf("解析响应失败: %v\n", err)
		return false
	}

	// 检查错误码
	if response.Errcode != 0 {
		log.Printf("微信接口返回错误: %s\n", response.Errmsg)
		return false
	}

	// 检查综合结果
	if response.Result != nil && response.Result.Label != 100 {
		log.Printf("内容可能存在风险，标签: %d\n", response.Result.Label)
		return false
	}

	return true
}

// 获取Accesstoken
func GetAccessToken() (string, error) {
	appID := "wx312c36de9d0cd635"
	appSecret := "521da6d7784073fbd7ae457f501e5a56"
	url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s", appID, appSecret)

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("请求微信接口失败: %v", err)
	}
	defer resp.Body.Close()

	// 读取并打印响应的内容
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("读取响应内容失败: %v", err)
	}
	fmt.Println("微信接口返回的响应:", string(body))

	// 定义一个结构体来映射响应内容
	var result struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	// 解析 JSON 响应
	if err := json.Unmarshal(body, &result); err != nil {
		return "", fmt.Errorf("解析响应内容失败: %v", err)
	}

	// 将 access_token 存储到 AccessToken 变量中
	AccessToken := result.AccessToken

	// 如果 AccessToken 不为空，返回它
	if AccessToken == "" {
		return "", fmt.Errorf("微信接口返回的access_token为空")
	}

	return AccessToken, nil
}

// timeAgo 返回类似于“X天前”或“X小时之前”的人类可读格式
func timeAgo(createdAt string, currentTime time.Time) string {
	createdTime, err := time.Parse("2006-01-02 15:04:05", createdAt)
	if err != nil {
		log.Println("Error parsing time:", err)
		return createdAt
	}

	diff := currentTime.Sub(createdTime)
	if diff.Hours() < 1 {
		return fmt.Sprintf("%d分钟之前", int(diff.Minutes()))
	} else if diff.Hours() < 24 {
		return fmt.Sprintf("%d小时之前", int(diff.Hours()))
	} else if diff.Hours() < 24*30 {
		return fmt.Sprintf("%d天之前", int(diff.Hours()/24))
	} else if diff.Hours() < 24*365 {
		return fmt.Sprintf("%d个月之前", int(diff.Hours()/24/30))
	}
	return fmt.Sprintf("%d年之前", int(diff.Hours()/24/365))
}

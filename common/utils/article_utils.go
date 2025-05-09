package utils

import "huancuilou/configs"

func ValidateArticleKind(kind string) bool {
	return configs.GetConfig().Article.KindMap[kind]
}

// Substring 截取字符串的前 n 个字符
func Substring(s string, n int) string {
	// 先将字符串转换为 rune 切片，rune 可以正确处理多字节字符
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	// 截取前 n 个 rune 并转换回 string
	return string(runes[:n])
}

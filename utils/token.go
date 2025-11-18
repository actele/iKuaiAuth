package utils

import (
	"crypto/md5"
	"fmt"
	"strings"
)

// GenerateToken 生成 iKuai 认证 token
// 参数格式: user_ip=1.2.3.4&timestamp=1542019734&mac=11:22:33:44:55:66&upload=10&download=10&key=xxx
// 注意：官方文档规定的token计算顺序为: user_ip, timestamp, mac, upload, download, key
func GenerateToken(params map[string]string, appKey string) string {
	// 严格按官方文档使用固定参数计算 token
	// 顺序: user_ip, timestamp, mac, upload, download, key
	keys := []string{"user_ip", "timestamp", "mac", "upload", "download"}
	var queryParts []string
	for _, k := range keys {
		if v, exists := params[k]; exists {
			queryParts = append(queryParts, fmt.Sprintf("%s=%s", k, v))
		} else {
			// 对于 mac，如果缺失或为空也需要以 mac= 参与计算；其余缺失的参数视为缺省空串
			if k == "mac" {
				queryParts = append(queryParts, "mac=")
			} else {
				queryParts = append(queryParts, fmt.Sprintf("%s=", k))
			}
		}
	}
	// 最后追加 key
	queryParts = append(queryParts, fmt.Sprintf("key=%s", appKey))

	queryString := strings.Join(queryParts, "&")

	// 计算 MD5
	hash := md5.Sum([]byte(queryString))
	token := fmt.Sprintf("%x", hash)

	return token
}

// ValidateRequiredParams 验证必需参数
func ValidateRequiredParams(params map[string]string) error {
	required := []string{"user_ip", "timestamp", "mac"}

	for _, param := range required {
		if _, exists := params[param]; !exists {
			return fmt.Errorf("缺少必需参数: %s", param)
		}
	}

	return nil
}

package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"ikuai-auth/config"
)

// APIAuthRequest 外部API认证请求
type APIAuthRequest struct {
	Username string
	Password string
}

// APIAuthResponse 外部API认证响应
type APIAuthResponse struct {
	Success bool
	Message string
}

// ValidateUserWithAPI 通过外部API验证用户
func ValidateUserWithAPI(username, password string) (bool, string) {
	cfg := config.Global.Auth.API

	// 验证配置
	if cfg.URL == "" {
		return false, "API认证未配置"
	}

	// 设置超时时间
	timeout := time.Duration(cfg.Timeout) * time.Second
	if cfg.Timeout == 0 {
		timeout = 5 * time.Second
	}

	client := &http.Client{
		Timeout: timeout,
	}

	var req *http.Request
	var err error

	// 构建请求
	if strings.ToUpper(cfg.Method) == "POST" {
		// 替换模板中的变量
		body := cfg.BodyTemplate
		body = strings.ReplaceAll(body, "{{username}}", username)
		body = strings.ReplaceAll(body, "{{password}}", password)

		req, err = http.NewRequest("POST", cfg.URL, bytes.NewBufferString(body))
		if err != nil {
			LogError("创建API请求失败: %v", err)
			return false, "创建请求失败"
		}

		// 设置默认Content-Type
		if cfg.Headers["Content-Type"] == "" {
			req.Header.Set("Content-Type", "application/json")
		}
	} else {
		// GET请求
		req, err = http.NewRequest("GET", cfg.URL, nil)
		if err != nil {
			LogError("创建API请求失败: %v", err)
			return false, "创建请求失败"
		}

		// 添加查询参数
		q := req.URL.Query()
		q.Add("username", username)
		q.Add("password", password)
		req.URL.RawQuery = q.Encode()
	}

	// 设置请求头
	for key, value := range cfg.Headers {
		req.Header.Set(key, value)
	}

	// 发送请求
	LogInfo("调用外部API认证: %s %s (用户: %s)", cfg.Method, cfg.URL, username)
	resp, err := client.Do(req)
	if err != nil {
		LogError("API请求失败: %v", err)
		return false, "认证服务不可用"
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		LogError("读取API响应失败: %v", err)
		return false, "读取响应失败"
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		LogError("API返回错误状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
		return false, fmt.Sprintf("认证失败(HTTP %d)", resp.StatusCode)
	}

	// 解析JSON响应
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		LogError("解析API响应失败: %v, 响应: %s", err, string(respBody))
		return false, "解析响应失败"
	}

	// 验证成功标识字段
	successField := cfg.Response.SuccessField
	if successField == "" {
		successField = "success"
	}

	successValue, exists := result[successField]
	if !exists {
		LogError("API响应中未找到成功标识字段: %s, 响应: %v", successField, result)
		return false, "响应格式错误"
	}

	// 比较成功值
	expectedValue := cfg.Response.SuccessValue
	if expectedValue == nil {
		expectedValue = true
	}

	isSuccess := false
	switch v := successValue.(type) {
	case bool:
		isSuccess = v == expectedValue
	case string:
		isSuccess = v == fmt.Sprintf("%v", expectedValue)
	case float64:
		isSuccess = v == expectedValue
	default:
		isSuccess = fmt.Sprintf("%v", v) == fmt.Sprintf("%v", expectedValue)
	}

	// 获取错误消息（如果存在）
	var message string
	if !isSuccess && cfg.Response.MessageField != "" {
		if msg, ok := result[cfg.Response.MessageField]; ok {
			message = fmt.Sprintf("%v", msg)
		}
	}

	if message == "" {
		if isSuccess {
			message = "认证成功"
		} else {
			message = "用户名或密码错误"
		}
	}

	if isSuccess {
		LogInfo("API认证成功 - 用户: %s", username)
	} else {
		LogInfo("API认证失败 - 用户: %s, 原因: %s", username, message)
	}

	return isSuccess, message
}

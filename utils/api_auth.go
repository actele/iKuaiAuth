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

// APIAuthUserInfo API认证返回的用户信息
type APIAuthUserInfo struct {
	Allowed      bool   // 是否允许认证
	Reason       string // 认证结果原因
	UserID       int    // 用户ID
	Username     string // 用户名（对应 iKuai user_id）
	RealName     string // 真实姓名（对应 iKuai name）
	Phone        string // 手机号（对应 iKuai phone）
	CustomName   string // 自定义名称（对应 iKuai custom_name）
	DeviceNumber string // 设备编号
	AssignedTo   string // 分配给谁
	VPNClient    string // VPN客户端
}

// ValidateUserWithAPI 通过外部API验证用户
// 返回: (是否成功, 错误消息, 用户信息)
func ValidateUserWithAPI(username, password string) (bool, string, *APIAuthUserInfo) {
	cfg := config.Global.Auth.API

	// 验证配置
	if cfg.URL == "" {
		return false, "API认证未配置", nil
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
			return false, "创建请求失败", nil
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
			return false, "创建请求失败", nil
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
		return false, "认证服务不可用", nil
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		LogError("读取API响应失败: %v", err)
		return false, "读取响应失败", nil
	}

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		LogError("API返回错误状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
		return false, fmt.Sprintf("认证失败(HTTP %d)", resp.StatusCode), nil
	}

	// 解析JSON响应
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		LogError("解析API响应失败: %v, 响应: %s", err, string(respBody))
		return false, "解析响应失败", nil
	}

	// 验证成功标识字段
	successField := cfg.Response.SuccessField
	if successField == "" {
		successField = "success"
	}

	successValue, exists := result[successField]
	if !exists {
		LogError("API响应中未找到成功标识字段: %s, 响应: %v", successField, result)
		return false, "响应格式错误", nil
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

	// 提取用户信息（如果认证成功）
	var userInfo *APIAuthUserInfo
	if isSuccess {
		userInfo = extractUserInfo(result)
		LogInfo("API认证成功 - 用户: %s (真实姓名: %s)", username, userInfo.RealName)
	} else {
		LogInfo("API认证失败 - 用户: %s, 原因: %s", username, message)
	}

	return isSuccess, message, userInfo
}

// ValidateByDeviceInfo 通过设备信息验证（标准API格式）
// 返回: (是否成功, 错误消息, 用户信息, API原始响应)
func ValidateByDeviceInfo(deviceNumber, userAccount, vpnNumber, macAddress, ipAddress string, systemID int) (bool, string, *APIAuthUserInfo, interface{}) {
	LogInfo("ValidateByDeviceInfo 开始 - 设备: %s, 用户账号: %s, VPN: %s, MAC: %s, IP: %s, SystemID: %d",
		deviceNumber, userAccount, vpnNumber, macAddress, ipAddress, systemID)

	// 根据认证方式处理
	authMethod := config.Global.Auth.Method

	// simple模式下直接放行（因为没有设备信息验证逻辑）
	if authMethod == "simple" {
		LogInfo("Simple模式 - 设备信息验证跳过，直接放行 - 设备: %s, 用户账号: %s, VPN: %s",
			deviceNumber, userAccount, vpnNumber)
		return true, "Simple模式直接放行", nil, nil
	}

	// API模式验证
	cfg := config.Global.Auth.API

	// 验证配置
	if cfg.URL == "" {
		LogError("API认证配置错误: URL未配置")
		return false, "API认证未配置", nil, nil
	}

	LogInfo("API配置 - URL: %s, Method: %s, Timeout: %d", cfg.URL, cfg.Method, cfg.Timeout)

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
		// 替换模板中的变量（标准API格式）
		body := cfg.BodyTemplate
		body = strings.ReplaceAll(body, "{{device_number}}", deviceNumber)
		body = strings.ReplaceAll(body, "{{user_account}}", userAccount)
		body = strings.ReplaceAll(body, "{{vpn_number}}", vpnNumber)
		body = strings.ReplaceAll(body, "{{mac_address}}", macAddress)
		body = strings.ReplaceAll(body, "{{ip_address}}", ipAddress)
		body = strings.ReplaceAll(body, "{{system_id}}", fmt.Sprintf("%d", systemID))

		LogInfo("请求体: %s", body)

		req, err = http.NewRequest("POST", cfg.URL, bytes.NewBufferString(body))
		if err != nil {
			LogError("创建API请求失败: %v", err)
			return false, "创建请求失败", nil, nil
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
			return false, "创建请求失败", nil, nil
		}

		// 添加查询参数（标准API格式）
		q := req.URL.Query()
		if deviceNumber != "" {
			q.Add("device_number", deviceNumber)
		}
		if userAccount != "" {
			q.Add("user_account", userAccount)
		}
		if vpnNumber != "" {
			q.Add("vpn_number", vpnNumber)
		}
		if macAddress != "" {
			q.Add("mac_address", macAddress)
		}
		if ipAddress != "" {
			q.Add("ip_address", ipAddress)
		}
		q.Add("system_id", fmt.Sprintf("%d", systemID))
		req.URL.RawQuery = q.Encode()
	}

	// 设置请求头
	for key, value := range cfg.Headers {
		req.Header.Set(key, value)
	}

	// 发送请求
	LogInfo("调用外部API设备认证: %s %s (设备: %s, 用户账号: %s, VPN: %s, MAC: %s, IP: %s, SystemID: %d)",
		cfg.Method, cfg.URL, deviceNumber, userAccount, vpnNumber, macAddress, ipAddress, systemID)
	resp, err := client.Do(req)
	if err != nil {
		LogError("API请求失败: %v", err)
		return false, "认证服务不可用", nil, nil
	}
	defer resp.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		LogError("读取API响应失败: %v", err)
		return false, "读取响应失败", nil, nil
	}

	LogInfo("外部API响应 - 状态码: %d, 响应体: %s", resp.StatusCode, string(respBody))

	// 检查HTTP状态码
	if resp.StatusCode != http.StatusOK {
		LogError("API返回错误状态码: %d, 响应: %s", resp.StatusCode, string(respBody))
		return false, fmt.Sprintf("认证失败(HTTP %d)", resp.StatusCode), nil, nil
	}

	// 解析JSON响应
	var result map[string]interface{}
	if err := json.Unmarshal(respBody, &result); err != nil {
		LogError("解析API响应失败: %v, 响应: %s", err, string(respBody))
		return false, "解析响应失败", nil, nil
	}

	LogInfo("解析后的响应: %+v", result)

	// 验证成功标识字段
	successField := cfg.Response.SuccessField
	if successField == "" {
		successField = "success"
	}

	LogInfo("检查成功字段: %s, 期望值: %v", successField, cfg.Response.SuccessValue)

	successValue, exists := result[successField]
	if !exists {
		LogError("API响应中未找到成功标识字段: %s, 完整响应: %v", successField, result)
		return false, "响应格式错误", nil, result
	}

	LogInfo("实际返回的 %s 值: %v (类型: %T)", successField, successValue, successValue)

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
			message = "设备认证成功"
		} else {
			message = "设备认证失败"
		}
	}

	// 提取用户信息（如果认证成功）
	var userInfo *APIAuthUserInfo
	if isSuccess {
		userInfo = extractUserInfo(result)
		LogInfo("设备API认证成功 - 设备: %s, 用户账号: %s, VPN: %s, 用户: %s",
			deviceNumber, userAccount, vpnNumber, userInfo.RealName)
	} else {
		LogInfo("设备API认证失败 - 设备: %s, 用户账号: %s, VPN: %s, 原因: %s",
			deviceNumber, userAccount, vpnNumber, message)
	}

	return isSuccess, message, userInfo, result
}

// extractUserInfo 从API响应中提取用户信息
func extractUserInfo(result map[string]interface{}) *APIAuthUserInfo {
	userInfo := &APIAuthUserInfo{}

	// 获取 data 字段
	data, ok := result["data"].(map[string]interface{})
	if !ok {
		return userInfo
	}

	// 提取字段
	if allowed, ok := data["allowed"].(bool); ok {
		userInfo.Allowed = allowed
	}
	if reason, ok := data["reason"].(string); ok {
		userInfo.Reason = reason
	}
	if userID, ok := data["user_id"].(float64); ok {
		userInfo.UserID = int(userID)
	}
	if username, ok := data["username"].(string); ok {
		userInfo.Username = username
	}
	if realName, ok := data["real_name"].(string); ok {
		userInfo.RealName = realName
	}
	if phone, ok := data["phone"].(string); ok {
		userInfo.Phone = phone
	}
	if deviceNumber, ok := data["device_number"].(string); ok {
		userInfo.DeviceNumber = deviceNumber
	}
	if assignedTo, ok := data["assigned_to"].(string); ok {
		userInfo.AssignedTo = assignedTo
	}
	if vpnClient, ok := data["vpn_client"].(string); ok {
		userInfo.VPNClient = vpnClient
	}

	// 如果没有 custom_name，使用 username
	if userInfo.Username != "" {
		userInfo.CustomName = userInfo.Username
	}

	return userInfo
}

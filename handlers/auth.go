package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"ikuai-auth/config"
	"ikuai-auth/utils"
)

// AuthRequest 网页认证请求结构体
type AuthRequest struct {
	// API标准认证字段
	DeviceNumber string `json:"device_number"` // 设备编号
	UserAccount  string `json:"user_account"`  // 使用人账号（手机号）
	VPNNumber    string `json:"vpn_number"`    // VPN线路编号
	MACAddress   string `json:"mac_address"`   // MAC地址
	IPAddress    string `json:"ip_address"`    // IP地址
	SystemID     int    `json:"system_id"`     // 路由系统ID

	// 兼容字段（旧格式）
	UserIP string `json:"user_ip"` // 用户IP地址（兼容）
	MAC    string `json:"mac"`     // MAC地址（兼容）
	Phone  string `json:"phone"`   // 手机号（兼容）

	// 传统认证字段（可选）
	Username  string `json:"username"`  // 用户名（可选）
	Password  string `json:"password"`  // 密码（可选）
	Timestamp string `json:"timestamp"` // 时间戳

	// 其他扩展字段
	Name    string `json:"name"`    // 姓名
	Comment string `json:"comment"` // 备注
	Timeout int    `json:"timeout"` // 认证超时时间(分钟)，0表示不过期
}

// AuthResponse 认证响应结构体
type AuthResponse struct {
	Success     bool              `json:"success"`
	Message     string            `json:"message"`
	Debug       bool              `json:"debug"`                  // 调试模式标志
	APIResponse interface{}       `json:"api_response,omitempty"` // 外部API原始响应（调试用）
	Data        *AuthResponseData `json:"data,omitempty"`         // 使用指针类型
}

// AuthResponseData 认证响应数据
type AuthResponseData struct {
	ReleaseURL    string      `json:"release_url,omitempty"`
	Token         string      `json:"token,omitempty"`
	ReleaseResult interface{} `json:"release_result,omitempty"`
}

// AuthHandler 处理用户认证请求
func AuthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

	var req AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		utils.LogError("解析请求失败: %v", err)
		respondWithError(w, "无效的请求数据", http.StatusBadRequest)
		return
	}

	// 兼容性处理：填充缺失的字段
	if req.MACAddress == "" && req.MAC != "" {
		req.MACAddress = req.MAC
	}
	if req.IPAddress == "" && req.UserIP != "" {
		req.IPAddress = req.UserIP
	}
	if req.UserAccount == "" && req.Phone != "" {
		req.UserAccount = req.Phone
	}
	if req.SystemID == 0 {
		req.SystemID = config.Global.Auth.API.SystemID
	}

	// 记录收到的认证请求
	utils.LogInfo("收到认证请求 - 设备: %s, 用户账号: %s, VPN: %s, IP: %s, MAC: %s, SystemID: %d",
		req.DeviceNumber, req.UserAccount, req.VPNNumber, req.IPAddress, req.MACAddress, req.SystemID)

	// 验证用户凭据
	// TODO: 临时关闭用户名密码校验，直接通过设备信息验证
	// success, userInfo := validateUser(req.Username, req.Password)
	// if !success {
	// 	utils.LogInfo("认证失败 - 用户: %s, IP: %s", req.Username, req.UserIP)
	// 	respondWithError(w, "用户名或密码错误", http.StatusUnauthorized)
	// 	return
	// }

	// 使用设备信息进行验证（设备编号、线路编号、使用人）
	utils.LogInfo("开始设备认证 - 认证方式: %s", config.Global.Auth.Method)
	success, _, userInfo, apiResponse := utils.ValidateByDeviceInfo(
		req.DeviceNumber,
		req.UserAccount,
		req.VPNNumber,
		req.MACAddress,
		req.IPAddress,
		req.SystemID,
	)
	if !success {
		utils.LogInfo("设备认证失败 - 设备: %s, 用户账号: %s, VPN: %s, IP: %s",
			req.DeviceNumber, req.UserAccount, req.VPNNumber, req.IPAddress)
		respondWithErrorWithDebug(w, "设备认证失败", http.StatusUnauthorized, apiResponse)
		return
	}

	// 生成放行 URL（传入用户信息）
	releaseURL, token, err := generateReleaseURL(req, userInfo)
	if err != nil {
		utils.LogError("生成放行URL失败 - 用户: %s, 错误: %v", req.Username, err)
		respondWithError(w, "生成放行链接失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 记录成功的认证
	utils.LogInfo("认证成功 - 用户: %s, IP: %s, MAC: %s", req.Username, req.UserIP, req.MAC)

	// 返回成功响应，让客户端在本地网络环境发起 GET 请求完成放行
	response := AuthResponse{
		Success: true,
		Message: "认证成功",
		Debug:   config.Global.Server.Debug, // 传递调试模式标志给前端
		Data: &AuthResponseData{
			ReleaseURL: releaseURL,
			Token:      token,
		},
	}

	// 使用自定义编码器，禁用HTML转义（确保&不会被转义为\u0026）
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.Encode(response)
}

// generateReleaseURL 生成 iKuai 放行链接
func generateReleaseURL(req AuthRequest, userInfo *utils.APIAuthUserInfo) (string, string, error) {
	// 从全局配置获取配置信息
	portalURL := config.Global.IKuai.PortalURL
	appKey := config.Global.IKuai.AppKey

	// 准备参数（网页认证使用默认值）
	// 如果路由器传入了 timestamp，则使用路由器的 timestamp，
	// 否则使用当前服务器时间。确保用于计算 token 的 timestamp
	// 与放行请求中的 timestamp 完全一致。
	timestamp := req.Timestamp
	if timestamp == "" {
		timestamp = strconv.FormatInt(time.Now().Unix(), 10)
	}

	// 使用 API 返回的用户信息或默认格式
	var userId, customName, name, phone, userIP, mac string

	// user_id 和 custom_name 只使用设备编号
	deviceNumber := req.DeviceNumber
	userId = deviceNumber
	customName = deviceNumber

	// 获取用户信息：优先使用API返回的数据
	if userInfo != nil {
		utils.LogInfo("API返回的用户信息 - RealName: %s, Phone: %s", userInfo.RealName, userInfo.Phone)
		// phone 使用手机号
		if userInfo.Phone != "" {
			phone = userInfo.Phone
		} else {
			phone = req.UserAccount
		}
		// name 使用姓名（real_name）
		name = userInfo.RealName
	} else {
		utils.LogInfo("未获取到API用户信息，使用请求数据")
		// 使用请求中的数据
		phone = req.UserAccount
		if phone == "" {
			phone = req.Phone
		}
		name = req.Name
	}

	utils.LogInfo("放行参数 - userId: %s, phone: %s, name: %s", userId, phone, name)

	// 优先使用新字段名 IPAddress 和 MACAddress，否则使用兼容字段
	userIP = req.IPAddress
	if userIP == "" {
		userIP = req.UserIP
	}
	mac = req.MACAddress
	if mac == "" {
		mac = req.MAC
	}

	params := map[string]string{
		"type":         "20",
		"user_id":      userId,
		"custom_name":  customName,
		"user_ip":      userIP,
		"timestamp":    timestamp,
		"mac":          mac,
		"upload":       "0", // 默认不限速
		"download":     "0", // 默认不限速
		"release_type": "1", // 1为网页认证，2为使用JSON格式（小程序/APP认证）
	}

	// 添加可选字段
	if phone != "" {
		params["phone"] = phone
	}
	if name != "" {
		params["name"] = name
	}
	// comment 使用线路编号
	if req.VPNNumber != "" {
		params["comment"] = req.VPNNumber
	} else if req.Comment != "" {
		params["comment"] = req.Comment
	}
	if req.Timeout > 0 {
		params["timeout"] = strconv.Itoa(req.Timeout)
	}

	// 生成 token
	token := utils.GenerateToken(params, appKey)

	// 构建放行 GET 请求的参数
	// 按照爱快官方文档要求的参数顺序：
	// type, user_id, custom_name, user_ip, timestamp, mac, timeout, upload, download, token, release_type
	// 可选参数: phone, name
	timeoutVal := "0"
	if t, ok := params["timeout"]; ok {
		timeoutVal = t
	}

	orderedParams := []string{
		fmt.Sprintf("type=%s", params["type"]),
		fmt.Sprintf("user_id=%s", params["user_id"]),
		fmt.Sprintf("custom_name=%s", params["custom_name"]),
		fmt.Sprintf("user_ip=%s", params["user_ip"]),
		fmt.Sprintf("timestamp=%s", params["timestamp"]),
		fmt.Sprintf("mac=%s", params["mac"]),
		fmt.Sprintf("timeout=%s", timeoutVal),
		fmt.Sprintf("upload=%s", params["upload"]),
		fmt.Sprintf("download=%s", params["download"]),
		fmt.Sprintf("token=%s", token),
		fmt.Sprintf("release_type=%s", params["release_type"]),
	}

	// 添加可选参数
	if phone, ok := params["phone"]; ok && phone != "" {
		orderedParams = append(orderedParams, fmt.Sprintf("phone=%s", phone))
	}
	if name, ok := params["name"]; ok && name != "" {
		orderedParams = append(orderedParams, fmt.Sprintf("name=%s", name))
	}
	if comment, ok := params["comment"]; ok && comment != "" {
		orderedParams = append(orderedParams, fmt.Sprintf("comment=%s", comment))
	}

	releaseURL := fmt.Sprintf("%s?%s", portalURL, strings.Join(orderedParams, "&"))

	return releaseURL, token, nil
}

// respondWithError 返回错误响应
func respondWithError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	response := AuthResponse{
		Success: false,
		Message: message,
		Debug:   config.Global.Server.Debug,
	}
	json.NewEncoder(w).Encode(response)
}

// respondWithErrorWithDebug 返回带调试信息的错误响应
func respondWithErrorWithDebug(w http.ResponseWriter, message string, statusCode int, apiResponse interface{}) {
	w.WriteHeader(statusCode)
	response := AuthResponse{
		Success: false,
		Message: message,
		Debug:   config.Global.Server.Debug,
	}
	if config.Global.Server.Debug && apiResponse != nil {
		response.APIResponse = apiResponse
	}
	json.NewEncoder(w).Encode(response)
}

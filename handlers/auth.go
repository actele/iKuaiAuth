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

// AuthRequest 网页认证请求结构体（包含iKuai所有参数）
type AuthRequest struct {
	// 用户凭据
	Username string `json:"username"`
	Password string `json:"password"`

	// 设备信息
	UserIP string `json:"user_ip"`
	MAC    string `json:"mac"`

	// iKuai路由器参数
	GWID      string `json:"gwid"`       // 网关ID
	RouterVer string `json:"router_ver"` // 路由器版本
	Firmware  string `json:"firmware"`   // 固件类型
	Timestamp string `json:"timestamp"`  // 时间戳
	Template  string `json:"template"`   // 模板类型

	// 认证类型和扩展字段
	Phone   string `json:"phone"`   // 手机号码
	Name    string `json:"name"`    // 姓名
	Comment string `json:"comment"` // 备注
	Timeout int    `json:"timeout"` // 认证超时时间(分钟)，0表示不过期
}

// AuthResponse 认证响应结构体
type AuthResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Debug   bool   `json:"debug"` // 调试模式标志
	Data    struct {
		ReleaseURL    string      `json:"release_url,omitempty"`
		Token         string      `json:"token,omitempty"`
		ReleaseResult interface{} `json:"release_result,omitempty"`
	} `json:"data,omitempty"`
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

	// 验证用户凭据
	if !validateUser(req.Username, req.Password) {
		utils.LogInfo("认证失败 - 用户: %s, IP: %s", req.Username, req.UserIP)
		respondWithError(w, "用户名或密码错误", http.StatusUnauthorized)
		return
	}

	// 生成放行 URL
	releaseURL, token, err := generateReleaseURL(req)
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
	}
	response.Data.ReleaseURL = releaseURL // 返回完整的 release URL 给前端
	response.Data.Token = token

	// 使用自定义编码器，禁用HTML转义（确保&不会被转义为\u0026）
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	encoder := json.NewEncoder(w)
	encoder.SetEscapeHTML(false)
	encoder.Encode(response)
}

// validateUser 验证用户凭据
func validateUser(username, password string) bool {
	// 根据配置的认证方式进行验证
	authMethod := config.Global.Auth.Method

	switch authMethod {
	case "api":
		// 使用外部API认证
		success, message := utils.ValidateUserWithAPI(username, password)
		if !success {
			utils.LogInfo("API认证失败 - 用户: %s, 原因: %s", username, message)
		}
		return success

	case "simple":
		fallthrough
	default:
		// 使用本地用户名密码认证
		validPassword, exists := config.Global.Auth.SimpleUsers[username]
		return exists && validPassword == password
	}
}

// generateReleaseURL 生成 iKuai 放行链接（简化版）
func generateReleaseURL(req AuthRequest) (string, string, error) {
	fmt.Println("\n🔗 ======== 生成放行URL ========")

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

	// 使用 "1020004_登录用户名" 格式的user_id和custom_name
	// 根据iKuai官方文档规定: "1020004_用户名"
	userId := fmt.Sprintf("1020004_%s", req.Username)
	customName := fmt.Sprintf("1020004_%s", req.Username)

	params := map[string]string{
		"type":         "20",
		"user_id":      userId,
		"custom_name":  customName,
		"user_ip":      req.UserIP,
		"timestamp":    timestamp,
		"mac":          req.MAC,
		"upload":       "0", // 默认不限速
		"download":     "0", // 默认不限速
		"release_type": "1", // 1为网页认证，2为使用JSON格式（小程序/APP认证）
	}

	// 添加可选字段（如果提供）
	if req.Phone != "" {
		params["phone"] = req.Phone
	}
	if req.Name != "" {
		params["name"] = req.Name
	}
	if req.Comment != "" {
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

	releaseURL := fmt.Sprintf("%s?%s", portalURL, strings.Join(orderedParams, "&"))

	return releaseURL, token, nil
}

// respondWithError 返回错误响应
func respondWithError(w http.ResponseWriter, message string, statusCode int) {
	w.WriteHeader(statusCode)
	response := AuthResponse{
		Success: false,
		Message: message,
	}
	json.NewEncoder(w).Encode(response)
}

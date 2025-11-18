package main

import (
	"embed"
	"log"
	"net/http"
	"os"

	"ikuai-auth/config"
	"ikuai-auth/utils"

	"gopkg.in/yaml.v2"
)

//go:embed web
var webFiles embed.FS

var appConfig config.Config

func main() {
	// 加载配置文件
	if err := loadConfig("config.yaml"); err != nil {
		log.Fatal("Failed to load config:", err)
	}

	// 设置全局配置
	config.SetGlobal(&appConfig)

	// 初始化日志（同时输出到控制台和文件）
	logDir := "logs"
	if err := utils.InitLogger(logDir); err != nil {
		log.Printf("Warning: 无法初始化文件日志: %v", err)
	}
	defer utils.CloseLogger()

	// 初始化路由
	router := setupRouter(webFiles)

	// 启动服务器
	addr := appConfig.Server.Host + ":" + appConfig.Server.Port
	utils.LogInfo("iKuai认证服务启动: %s (Debug: %v)", addr, appConfig.Server.Debug)

	if err := http.ListenAndServe(addr, router); err != nil {
		utils.LogError("服务器错误: %v", err)
		log.Fatal(err)
	}
}

// loadConfig 加载配置文件
func loadConfig(filename string) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, &appConfig)
}

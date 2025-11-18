package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Logger 日志记录器
type Logger struct {
	file   *os.File
	logger *log.Logger
}

var globalLogger *Logger

// InitLogger 初始化日志记录器
func InitLogger(logDir string) error {
	// 创建日志目录
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %v", err)
	}

	// 生成日志文件名（按日期）
	logFile := filepath.Join(logDir, fmt.Sprintf("service-%s.log", time.Now().Format("2006-01-02")))

	// 打开日志文件（追加模式）
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("打开日志文件失败: %v", err)
	}

	// 创建 logger
	logger := log.New(file, "", log.LstdFlags)

	globalLogger = &Logger{
		file:   file,
		logger: logger,
	}

	return nil
}

// CloseLogger 关闭日志记录器
func CloseLogger() {
	if globalLogger != nil && globalLogger.file != nil {
		globalLogger.file.Close()
	}
}

// LogInfo 记录信息日志（同时输出到控制台和文件）
func LogInfo(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)

	// 输出到控制台
	fmt.Println(message)

	// 输出到文件
	if globalLogger != nil && globalLogger.logger != nil {
		globalLogger.logger.Println("[INFO] " + message)
	}
}

// LogError 记录错误日志（同时输出到控制台和文件）
func LogError(format string, v ...interface{}) {
	message := fmt.Sprintf(format, v...)

	// 输出到控制台
	fmt.Printf("ERROR: %s\n", message)

	// 输出到文件
	if globalLogger != nil && globalLogger.logger != nil {
		globalLogger.logger.Println("[ERROR] " + message)
	}
}

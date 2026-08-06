package main

import (
	"embed"
	"io/fs"
	"net/http"

	"ikuai-auth/handlers"

	"github.com/gorilla/mux"
)

// setupRouter 设置路由
func setupRouter(webFiles embed.FS) *mux.Router {
	r := mux.NewRouter()

	// API 路由（必须在静态文件路由之前）
	api := r.PathPrefix("/api").Subrouter()

	// 认证相关路由
	api.HandleFunc("/auth", handlers.AuthHandler).Methods("POST")
	api.HandleFunc("/auth/network", handlers.AuthHandler).Methods("POST") // 网络设备认证路由（经过后端处理）

	// 配置API - 返回前端需要的配置信息
	api.HandleFunc("/config", handlers.ConfigHandler).Methods("GET")

	// 健康检查
	api.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}).Methods("GET")

	// 美化的页面路由（隐藏 .html 后缀）
	webFS, _ := fs.Sub(webFiles, "web")

	// 认证登录页面 - / 或 /login
	r.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		serveFile(w, r, webFS, "index.html")
	}).Methods("GET")

	r.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		serveFile(w, r, webFS, "index.html")
	}).Methods("GET")

	// 认证成功页面 - /success
	r.HandleFunc("/success", func(w http.ResponseWriter, r *http.Request) {
		serveFile(w, r, webFS, "success.html")
	}).Methods("GET")

	// 导航页 - /nav
	r.HandleFunc("/nav", func(w http.ResponseWriter, r *http.Request) {
		serveFile(w, r, webFS, "nav.html")
	}).Methods("GET")

	// 静态资源（JS, CSS, 图片等）
	r.PathPrefix("/").Handler(http.FileServer(http.FS(webFS)))

	return r
}

// serveFile 从嵌入的文件系统中提供文件
func serveFile(w http.ResponseWriter, r *http.Request, fsys fs.FS, name string) {
	data, err := fs.ReadFile(fsys, name)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	// 根据文件扩展名设置 Content-Type
	if name == "index.html" || name == "success.html" || name == "nav.html" {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	}

	w.Write(data)
}

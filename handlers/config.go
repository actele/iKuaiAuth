package handlers

import (
	"encoding/json"
	"net/http"

	"ikuai-auth/config"
)

// ConfigResponse 配置响应结构
type ConfigResponse struct {
	PortalURL string `json:"portal_url"`
	AppKey    string `json:"app_key"`
	Debug     bool   `json:"debug"`
}

// ConfigHandler 返回前端需要的配置信息
func ConfigHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	response := ConfigResponse{
		PortalURL: config.Global.IKuai.PortalURL,
		AppKey:    config.Global.IKuai.AppKey,
		Debug:     config.Global.Server.Debug,
	}

	json.NewEncoder(w).Encode(response)
}

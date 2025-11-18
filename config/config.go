package config

// Config 全局配置结构体
type Config struct {
	Server struct {
		Port  string `yaml:"port"`
		Host  string `yaml:"host"`
		Debug bool   `yaml:"debug"` // 调试模式开关
	} `yaml:"server"`
	IKuai struct {
		PortalURL string `yaml:"portal_url"`
		AppKey    string `yaml:"app_key"`
	} `yaml:"ikuai"`
	Auth struct {
		Method      string            `yaml:"method"`
		SimpleUsers map[string]string `yaml:"simple_users"`
		API         struct {
			URL          string            `yaml:"url"`
			Method       string            `yaml:"method"`
			Timeout      int               `yaml:"timeout"`
			Headers      map[string]string `yaml:"headers"`
			BodyTemplate string            `yaml:"body_template"`
			Response     struct {
				SuccessField string      `yaml:"success_field"`
				SuccessValue interface{} `yaml:"success_value"`
				MessageField string      `yaml:"message_field"`
			} `yaml:"response"`
		} `yaml:"api"`
	} `yaml:"auth"`
}

// Global 全局配置实例
var Global *Config

// SetGlobal 设置全局配置
func SetGlobal(c *Config) {
	Global = c
}

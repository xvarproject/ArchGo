package openai_config

import (
	"ArchGo/logger"
	"encoding/json"
	"log"
	"os"
	"sync"
)

// OpenAIConfiguration 项目配置
type OpenAIConfiguration struct {
	// gpt apikey
	ApiKey string `json:"api_key"`
	// openai提供的接口 空字符串使用默认接口
	ApiURL string `json:"api_url"`
	// 监听接口
	Listen string `json:"listen"`
	// 代理
	Proxy         string   `json:"proxy"`
	AdminEmail    []string `json:"admin_email"`
	AdminPassword string   `json:"admin_password"`
}

var config *OpenAIConfiguration
var once sync.Once

// LoadOpenAIConfig 加载配置
func LoadOpenAIConfig() *OpenAIConfiguration {
	once.Do(func() {
		// 给配置赋默认值
		config = &OpenAIConfiguration{
			ApiURL: "",
			Listen: "",
		}

		// 判断配置文件是否存在，存在直接JSON读取
		_, err := os.Stat(CLI.Config)
		if err == nil {
			f, err := os.Open(CLI.Config)
			if err != nil {
				log.Fatalf("open openai-config err: %v", err)
				return
			}
			defer f.Close()
			encoder := json.NewDecoder(f)
			err = encoder.Decode(config)
			if err != nil {
				log.Fatalf("decode openai-config err: %v", err)
				return
			}
		}
	})
	if config.ApiKey == "" {
		logger.Danger("openai-config err: api key required")
	}

	return config
}

package config

import (
	"fmt"
	"os"
	"strings"
	"sync"

	"github.com/joho/godotenv"
	"github.com/spf13/viper"
)

var (
	once      sync.Once
	globalCfg *AppConfig
)

// Load 根据环境变量 APP_ENV 自动加载对应配置文件
// APP_ENV=dev → configs/config_dev.yaml，默认 → configs/config.yaml
func Load() *AppConfig {
	once.Do(func() {
		// 加载 .env 文件
		godotenv.Load("configs/.env")

		// 获取环境
		env := os.Getenv("APP_ENV")
		cfgFile := "configs/config.yaml"
		if env == "dev" {
			cfgFile = "configs/config_dev.yaml"
		}

		viper.SetConfigFile(cfgFile)
		viper.AutomaticEnv()
		viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

		// 展开环境变量占位符
		viper.AutomaticEnv()

		if err := viper.ReadInConfig(); err != nil {
			panic(fmt.Sprintf("failed to load config: %v", err))
		}

		// 解析配置到结构体
		globalCfg = &AppConfig{}
		if err := viper.Unmarshal(globalCfg); err != nil {
			panic(fmt.Sprintf("failed to unmarshal config: %v", err))
		}

		// 处理环境变量替换
		processEnvVars(globalCfg)

		// 设置默认 embedding 维度
		if globalCfg.LLM.EmbeddingDim == 0 {
			globalCfg.LLM.EmbeddingDim = 1536 // OpenAI text-embedding-3-small 默认维度
		}
	})
	return globalCfg
}

// Get 获取已加载的配置实例
func Get() *AppConfig {
	if globalCfg == nil {
		return Load()
	}
	return globalCfg
}

// processEnvVars 处理配置中的环境变量占位符 ${ENV_VAR}
func processEnvVars(cfg *AppConfig) {
	// 处理 LLM API Key
	if strings.HasPrefix(cfg.LLM.APIKey, "${") && strings.HasSuffix(cfg.LLM.APIKey, "}") {
		envKey := cfg.LLM.APIKey[2 : len(cfg.LLM.APIKey)-1]
		if val := os.Getenv(envKey); val != "" {
			cfg.LLM.APIKey = val
		}
	}

	// 处理 Tavily API Key
	if strings.HasPrefix(cfg.Tools.Search.TavilyAPIKey, "${") && strings.HasSuffix(cfg.Tools.Search.TavilyAPIKey, "}") {
		envKey := cfg.Tools.Search.TavilyAPIKey[2 : len(cfg.Tools.Search.TavilyAPIKey)-1]
		if val := os.Getenv(envKey); val != "" {
			cfg.Tools.Search.TavilyAPIKey = val
		}
	}
}

// GetEnv 获取环境变量值，如果未设置则返回默认值
func GetEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

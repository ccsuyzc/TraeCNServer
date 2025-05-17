package config

import (
	// "fmt"
	// "os"
	// "strconv"
	"sync"
)

var Token string = "sk-340a96bc42f74a98b33b4bfe49953387"

type AIConfig struct {
	APIKey      string
	APIBaseURL  string
	Model       string
	MaxTokens   int
	Temperature float64
}

var (
	aiConfig   *AIConfig
	configOnce sync.Once
)

// func getEnvFloat(key string, fallback float64) float64 {
// 	if value := os.Getenv(key); value != "" {
// 		if f, err := strconv.ParseFloat(value, 64); err == nil {
// 			return f
// 		}
// 	}
// 	return fallback
// }

func GetAIConfig() (*AIConfig, error) {
	var err error
	configOnce.Do(func() {
		aiConfig = &AIConfig{
			// APIKey:      os.Getenv("AI_API_KEY"),
			APIKey: "sk-340a96bc42f74a98b33b4bfe49953387",
			// APIBaseURL:  os.Getenv("AI_API_BASE_URL"),
			APIBaseURL: "https://api.deepseek.com/chat/completions",
			Model:      "deepseek-chat",
			MaxTokens:  2048,
			// Temperature: getEnvFloat("AI_TEMPERATURE", 0.7),
			Temperature: 0.7,
		}
		// if aiConfig.APIKey == "" {
		// 	err = fmt.Errorf("AI_API_KEY environment variable not set")
		// }
	})
	return aiConfig, err
}

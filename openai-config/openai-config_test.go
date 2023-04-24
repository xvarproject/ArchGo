package openai_config

import (
	"testing"
)

func TestReadOpenAIConfig(t *testing.T) {
	CLI.Config = "openai-config.example.json"

	OpenaiConfig := LoadOpenAIConfig()
	for _, element := range OpenaiConfig.AdminEmail {
		println(element)
	}
}

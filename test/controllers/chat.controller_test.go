package controllers

import (
	"ArchGo/controllers"
	"github.com/magiconair/properties/assert"
	"github.com/sashabaranov/go-openai"
	"testing"
)

func TestGetModelCharge(t *testing.T) {
	usage := openai.Usage{PromptTokens: 2500, CompletionTokens: 500, TotalTokens: 3000}
	result := controllers.GetModelChargeInChineseCents(openai.GPT3Dot5Turbo0301, usage)
	assert.Equal(t, result, int64(7))
	result = controllers.GetModelChargeInChineseCents(openai.GPT4, usage)
	testString := "我们都是好朋友,你在干什么"
	count := controllers.NumTokens(testString)
	assert.Equal(t, count, 15)

	assert.Equal(t, result, int64(116))
}

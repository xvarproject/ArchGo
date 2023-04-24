package controllers

import (
	"ArchGo/constants"
	"ArchGo/logger"
	"ArchGo/models"
	"context"
	"fmt"
	"github.com/pkoukk/tiktoken-go"
	"io"
	"math"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"ArchGo/openai-config"
	"github.com/gin-gonic/gin"
	gogpt "github.com/sashabaranov/go-openai"
	"golang.org/x/net/proxy"
)

type BaseController struct {
}

func (*BaseController) ResponseJson(ctx *gin.Context, code int, errorMsg string, data interface{}) {

	ctx.JSON(code, gin.H{
		"code":     code,
		"errorMsg": errorMsg,
		"data":     data,
	})
	ctx.Abort()
}

func (*BaseController) ResponseData(ctx *gin.Context, code int, contentType string, data []byte) {
	ctx.Data(code, contentType, data)
	ctx.Abort()
}

type ChatController struct {
	BaseController
	cs *CreditSystem
}

func NewChatController(creditSystem *CreditSystem) ChatController {
	return ChatController{cs: creditSystem}
}

func NumTokens(text string) int {
	encoding := "cl100k_base"

	tke, err := tiktoken.GetEncoding(encoding)
	if err != nil {
		err = fmt.Errorf("GetEncoding: %v", err)
		return -1
	}

	// encode
	token := tke.Encode(text, nil, nil)

	// num_tokens
	numTokens := len(token)
	return numTokens
}

func (c *ChatController) CompletionWithModelInfo(ctx *gin.Context) {
	var request gogpt.ChatCompletionRequest
	err := ctx.BindJSON(&request)
	if err != nil {
		c.ResponseJson(ctx, http.StatusInternalServerError, err.Error(), nil)
		return
	}
	logger.Info("Starting a request with model %s", request.Model)
	if len(request.Messages) == 0 {
		c.ResponseJson(ctx, http.StatusBadRequest, "request messages required", nil)
		return
	}

	cnf := openai_config.LoadOpenAIConfig()
	gptConfig := gogpt.DefaultConfig(cnf.ApiKey)

	if cnf.Proxy != "" {
		transport := &http.Transport{}

		if strings.HasPrefix(cnf.Proxy, "socks5h://") {
			// 创建一个 DialContext 对象，并设置代理服务器
			dialContext, err := newDialContext(cnf.Proxy[10:])
			if err != nil {
				panic(err)
			}
			transport.DialContext = dialContext
		} else {
			// 创建一个 HTTP Transport 对象，并设置代理服务器
			proxyUrl, err := url.Parse(cnf.Proxy)
			if err != nil {
				panic(err)
			}
			transport.Proxy = http.ProxyURL(proxyUrl)
		}
		// 创建一个 HTTP 客户端，并将 Transport 对象设置为其 Transport 字段
		gptConfig.HTTPClient = &http.Client{
			Transport: transport,
		}

	}

	// 自定义gptConfig.BaseURL
	if cnf.ApiURL != "" {
		gptConfig.BaseURL = cnf.ApiURL
	}

	client := gogpt.NewClientWithConfig(gptConfig)
	if request.Model == "" {
		logger.Danger("request model is empty")
		c.ResponseJson(ctx, http.StatusBadRequest, "request model is empty", nil)
	}

	currentUser := ctx.MustGet("currentUser").(models.User)
	if !c.preCheckBalance(ctx, currentUser.Balance) {
		return
	}

	if request.Model == gogpt.GPT3Dot5Turbo0301 || request.Model == gogpt.GPT4 || request.Model == gogpt.
		GPT40314 || request.
		Model == gogpt.
		GPT3Dot5Turbo {
		if request.Stream {
			logger.Info("stream request started")
			stream, err := client.CreateChatCompletionStream(ctx, request)
			tokenPromptCount := 0
			for _, msg := range request.Messages {
				tokenPromptCount += NumTokens(msg.Content)
			}
			if err != nil {
				c.ResponseJson(ctx, http.StatusInternalServerError, err.Error(), nil)
				return
			}

			chanStream := make(chan string, 10)
			res := ""

			go func() {
				for {
					nextResp, err := stream.Recv()
					if err == io.EOF {
						stream.Close()
						tokenCompletionCount := NumTokens(res)
						totalTokenCount := tokenPromptCount + tokenCompletionCount
						usage := gogpt.Usage{PromptTokens: tokenPromptCount, CompletionTokens: tokenCompletionCount, TotalTokens: totalTokenCount}
						cost := GetModelChargeInChineseCents(request.Model, usage)
						chanStream <- fmt.Sprintf("[TotalCredit: %d]", cost)
						c.chargeUserFromBalance(currentUser.Email, cost)
						return
					} else if err != nil {
						c.ResponseJson(ctx, http.StatusInternalServerError, err.Error(), nil)
						return
					} else {
						chanStream <- nextResp.Choices[0].Delta.Content
						res += nextResp.Choices[0].Delta.Content
					}
				}

			}()

			ctx.Stream(func(w io.Writer) bool {
				if msg, ok := <-chanStream; ok {
					if !strings.HasPrefix(msg, "[TotalCredit:") {
						_, err := w.Write([]byte(msg))
						if err != nil {
							logger.Warning(err.Error())
							return false
						}
						return true
					} else {
						_, err := w.Write([]byte(msg))
						if err != nil {
							logger.Warning(err.Error())
							return false
						}
						close(chanStream)
						return true
					}
				}
				return false
			})

		} else {
			resp, err := client.CreateChatCompletion(ctx, request)
			if err != nil {
				c.ResponseJson(ctx, http.StatusInternalServerError, err.Error(), nil)
				return
			}
			cost := GetModelChargeInChineseCents(request.Model, resp.Usage)
			c.chargeUserFromBalance(currentUser.Email, cost)

			c.ResponseJson(ctx, http.StatusOK, "", gin.H{
				"reply":       resp.Choices[0].Message.Content,
				"messages":    append(request.Messages, resp.Choices[0].Message),
				"totalCredit": cost,
			})
		}
	} else {
		prompt := ""
		for _, item := range request.Messages {
			prompt += item.Content + "/n"
		}
		prompt = strings.Trim(prompt, "/n")

		logger.Info("request prompt is %s", prompt)
		req := gogpt.CompletionRequest{
			Model:            request.Model,
			MaxTokens:        request.MaxTokens,
			TopP:             request.TopP,
			FrequencyPenalty: request.FrequencyPenalty,
			PresencePenalty:  request.PresencePenalty,
			Prompt:           prompt,
		}

		resp, err := client.CreateCompletion(ctx, req)

		cost := GetModelChargeInChineseCents(request.Model, resp.Usage)
		c.chargeUserFromBalance(currentUser.Email, cost)

		if err != nil {
			c.ResponseJson(ctx, http.StatusInternalServerError, err.Error(), nil)
			return
		}

		c.ResponseJson(ctx, http.StatusOK, "", gin.H{
			"reply": resp.Choices[0].Text,
			"messages": append(request.Messages, gogpt.ChatCompletionMessage{
				Role:    "assistant",
				Content: resp.Choices[0].Text,
			}),
		})
	}

}

func (c *ChatController) preCheckBalance(ctx *gin.Context, balance int64) bool {
	if balance < 5 {
		c.ResponseJson(ctx, http.StatusBadRequest, "Insufficient balance", nil)
		return false
	}
	return true
}

func (c *ChatController) chargeUserFromBalance(email string, cost int64) {
	_, err := c.cs.UpdateBalanceByUserEmail(email, -1*cost)
	if err != nil {
		logger.Warning("charge user %s at cost %d failed", email, cost)
	} else {
		logger.Info("charge user %s at cost %d success", email, cost)
	}
}

// Pricing see https://openai.com/pricing#language-models

func GetModelChargeInChineseCents(model string, usage gogpt.Usage) int64 {
	var chargeInChineseCents float64 = 0
	switch {
	case model == gogpt.GPT3Dot5Turbo0301 || model == gogpt.GPT3Dot5Turbo:
		chargeInChineseCents = (float64(usage.PromptTokens)*(constants.GPT3PromptCharge) + float64(usage.
			CompletionTokens)*(constants.
			GPT3CompletionCharge)) * constants.DollarToChineseCentsRate
		break
	case model == gogpt.GPT4 || model == gogpt.GPT40314:
		chargeInChineseCents = (float64(usage.PromptTokens)*(constants.GPT4PromptCharge) + float64(usage.
			CompletionTokens)*(constants.
			GPT4CompletionCharge)) * constants.DollarToChineseCentsRate
		break
	}
	return int64(math.Ceil(chargeInChineseCents))
}

type dialContextFunc func(ctx context.Context, network, address string) (net.Conn, error)

func newDialContext(socks5 string) (dialContextFunc, error) {
	baseDialer := &net.Dialer{
		Timeout:   60 * time.Second,
		KeepAlive: 60 * time.Second,
	}

	if socks5 != "" {
		// split socks5 proxy string [username:password@]host:port
		var auth *proxy.Auth = nil

		if strings.Contains(socks5, "@") {
			proxyInfo := strings.SplitN(socks5, "@", 2)
			proxyUser := strings.Split(proxyInfo[0], ":")
			if len(proxyUser) == 2 {
				auth = &proxy.Auth{
					User:     proxyUser[0],
					Password: proxyUser[1],
				}
			}
			socks5 = proxyInfo[1]
		}

		dialSocksProxy, err := proxy.SOCKS5("tcp", socks5, auth, baseDialer)
		if err != nil {
			return nil, err
		}

		contextDialer, ok := dialSocksProxy.(proxy.ContextDialer)
		if !ok {
			return nil, err
		}

		return contextDialer.DialContext, nil
	} else {
		return baseDialer.DialContext, nil
	}
}

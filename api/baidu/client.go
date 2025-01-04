package baidu

import (
	"fmt"
	"strings"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

const (
	defaultUrl = "https://fanyi-api.baidu.com/api/trans/vip/translate"
	salt       = "aty123456"
)

type Client struct {
	apiKey  string
	secret  string
	httpCli *resty.Client
	logger  *zap.SugaredLogger
}

func NewClient(apiKey string, secretKey string, logger *zap.SugaredLogger, options ...func(*resty.Client)) *Client {
	httpClient := resty.New()
	for _, option := range options {
		option(httpClient)
	}
	if httpClient.BaseURL == "" {
		logger.Infof("No url provided, Using default url: %s", defaultUrl)
		httpClient.SetBaseURL(defaultUrl)
	}
	return &Client{
		httpCli: httpClient,
		logger:  logger,
		apiKey:  apiKey,
		secret:  secretKey,
	}
}

func (c Client) Translate(text string, sourceLang string, targetLang string) (string, error) {
	// text需要控制在6000bytes以内
	if len(text) > 6000 {
		c.logger.Errorw("Translation failed", zap.String("reason", "text too long"))
		return "", fmt.Errorf("Translation failed: text too long")
	}
	sourceLang = langMap[strings.ToLower(sourceLang)]
	targetLang = langMap[strings.ToLower(targetLang)]
	var result BaiduResponse
	s := sign(c.apiKey, text, salt, c.secret)
	resp, err := c.httpCli.R().
		SetFormData(map[string]string{"q": text, "from": sourceLang, "to": targetLang, "appid": c.apiKey, "salt": salt, "sign": s}).
		SetResult(&result).
		SetHeader("Content-Type", "application/x-www-form-urlencoded").
		Post("")
	if err != nil {
		c.logger.Errorw("Translation failed", zap.Error(err))
		return "", err
	}
	if resp.IsError() {
		c.logger.Errorw("Translation failed", zap.String("status", resp.Status()))
		return "", fmt.Errorf("Translation failed: %s", resp.Status())
	}
	if result.Error_code != "" && result.Error_code != "52000" {
		c.logger.Errorw("Translation failed", zap.String("error_code", result.Error_code))
		return "", fmt.Errorf("Translation failed: %s", result.Error_code)
	}
	if len(result.TransResult) == 0 {
		c.logger.Infof("Translation result is empty")
		return "", nil
	}
	var builder strings.Builder
	for _, v := range result.TransResult {
		builder.WriteString(v.Dst)
	}
	return builder.String(), nil
}

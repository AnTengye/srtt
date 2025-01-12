package baidu

import (
	"fmt"
	"net/http"
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
	httpClient.AddRetryCondition(func(r *resty.Response, err error) bool {
		switch r.StatusCode() {
		case http.StatusRequestTimeout, http.StatusTooManyRequests, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
			return true
		}
		return false
	})
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
func before(text []string) string {
	return strings.Join(text, "\n----\n")
}
func (c Client) Translate(t []string, sourceLang string, targetLang string) ([]string, error) {
	text := before(t)
	// text需要控制在6000bytes以内
	if len(text) > 6000 {
		c.logger.Errorw("Translation failed", zap.String("reason", "text too long"))
		return nil, fmt.Errorf("Translation failed: text too long")
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
		return nil, err
	}
	if resp.IsError() {
		c.logger.Errorw("Translation failed", zap.String("status", resp.Status()))
		return nil, fmt.Errorf("Translation failed: %s", resp.Status())
	}
	if result.Error_code != "" && result.Error_code != "52000" {
		c.logger.Errorw("Translation failed", zap.String("error_code", result.Error_code))
		return nil, fmt.Errorf("Translation failed: %s", result.Error_code)
	}
	if len(result.TransResult) == 0 {
		c.logger.Infof("Translation result is empty")
		return nil, nil
	}
	translatedText := make([]string, 0, len(result.TransResult))
	for _, v := range result.TransResult {
		if v.Dst != "----" {
			translatedText = append(translatedText, v.Dst)
		}
	}
	return translatedText, nil
}

func (c *Client) Close() error {
	return nil
}

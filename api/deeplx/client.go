package deeplx

import (
	"fmt"

	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

const (
	defaultUrl = "http://127.0.0.1:1188/translate"
)

type Client struct {
	httpCli *resty.Client
	logger  *zap.SugaredLogger
}

func (c Client) Translate(text string, sourceLang string, targetLang string) (string, error) {
	var result DeeplxResponse
	resp, err := c.httpCli.R().
		SetBody(map[string]interface{}{"text": text, "source_lang": sourceLang, "target_lang": targetLang}).
		SetResult(&result). // or SetResult(AuthSuccess{}).
		Post("")
	if err != nil {
		c.logger.Errorw("Translation failed", zap.Error(err))
		return "", err
	}
	if resp.IsError() {
		c.logger.Errorw("Translation failed", zap.String("status", resp.Status()))
		return "", fmt.Errorf("Translation failed: %s", resp.Status())
	}
	return result.Data, nil
}

func NewClient(logger *zap.SugaredLogger, options ...func(*resty.Client)) *Client {
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
	}
}

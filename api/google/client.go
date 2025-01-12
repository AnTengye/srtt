package google

import (
	"context"
	"fmt"

	"cloud.google.com/go/translate"
	translatev3 "cloud.google.com/go/translate/apiv3"
	"cloud.google.com/go/translate/apiv3/translatepb"
	"go.uber.org/zap"
	"golang.org/x/text/language"
	"google.golang.org/api/option"
)

type Client struct {
	ctx       context.Context
	v2Cli     *translate.Client
	v3Cli     *translatev3.TranslationClient
	logger    *zap.SugaredLogger
	isBasic   bool
	projectID string
}

func (c *Client) Translate(text []string, sourceLang string, targetLang string) ([]string, error) {
	if c.isBasic {
		return c.translateTextBasic(targetLang, text)
	}
	return c.translateTextPro(sourceLang, targetLang, text)
}

func NewClient(apikey, projectID string, logger *zap.SugaredLogger, isBasic bool) *Client {
	ctx := context.Background()
	c := &Client{
		ctx:       ctx,
		logger:    logger,
		isBasic:   isBasic,
		projectID: projectID,
	}

	if isBasic {
		client, err := translate.NewClient(ctx, option.WithAPIKey(apikey))
		if err != nil {
			logger.Errorf("translate.NewClient: %s", err)
			return nil
		}
		c.v2Cli = client
	} else {
		client, err := translatev3.NewTranslationClient(ctx, option.WithAPIKey(apikey))
		if err != nil {
			logger.Errorf("translatev3.NewTranslationClient: %s", err)
			return nil
		}
		c.v3Cli = client
	}
	return c
}

func (c *Client) translateTextBasic(targetLanguage string, text []string) ([]string, error) {
	if c.v2Cli == nil {
		return nil, fmt.Errorf("translate.TranslateClient is nil")
	}
	lang, err := language.Parse(targetLanguage)
	if err != nil {
		return nil, fmt.Errorf("language.Parse: %w", err)
	}

	resp, err := c.v2Cli.Translate(c.ctx, text, lang, nil)
	if err != nil {
		return nil, fmt.Errorf("Translate: %w", err)
	}
	if len(resp) == 0 {
		return nil, fmt.Errorf("Translate returned empty response to text: %s", text)
	}
	result := make([]string, len(resp))
	for i, r := range resp {
		result[i] = r.Text
	}

	return result, nil
}

func (c *Client) translateTextPro(sourceLang string, targetLang string, text []string) ([]string, error) {
	if c.v3Cli == nil {
		return nil, fmt.Errorf("translatev3.TranslationClient is nil")
	}
	req := &translatepb.TranslateTextRequest{
		Parent:             fmt.Sprintf("projects/%s/locations/global", c.projectID),
		SourceLanguageCode: sourceLang,
		TargetLanguageCode: targetLang,
		MimeType:           "text/plain", // Mime types: "text/plain", "text/html"
		Contents:           text,
	}

	resp, err := c.v3Cli.TranslateText(c.ctx, req)
	if err != nil {
		return nil, fmt.Errorf("TranslateText: %w", err)
	}

	result := make([]string, len(resp.GetTranslations()))
	for i, translation := range resp.GetTranslations() {
		result[i] = translation.GetTranslatedText()
	}

	return result, nil
}

func (c *Client) Close() error {
	if c.v2Cli != nil {
		return c.v2Cli.Close()
	}
	if c.v3Cli != nil {
		return c.v3Cli.Close()
	}
	return nil
}

package chatgpt

import (
	"context"
	"strings"

	"github.com/sashabaranov/go-openai"
	"go.uber.org/zap"
)

const (
	defaultUrl    = "http://127.0.0.1:8000/v1"
	defaultPrompt = `你将扮演两个角色，一个精通日语俚语和擅长中文表达的翻译家； 另一个角色是一个精通日语和中文的校对者，能够理解日语的俚语、深层次意思，也同样擅长中文表达。
每次我都会给你一段有固定格式的日语对话：
1. 请你先作为翻译家，把它翻译成中文，用尽可能地道的中文表达。在翻译之前，你应该先提取日语句子或者段落中的关键词组，先解释它们的意思，再翻译。
2. 然后你扮演校对者，审视原文和译文，检查原文和译文中意思有所出入的地方，提出修改意见
3. 最后，你再重新扮演翻译家，根据修改意见重新翻译，得到最后的译文
请注意，翻译时遇到不适当的表达也尽可能保留，不要改变语境，不要改变原文的意思。
你的思考过程应该遵循以下的格式：
译文初稿:{结合以上分析，翻译得到的译文}
校对:{重复以下列表，列出可能需要修改的地方}
- 校对意见{1...n}:
- 原文：{日语}
- 译文：{相关译文}
- 问题：{原文跟译文意见有哪些出入，或者译文的表达不够地道的地方}
- 建议：{应如何修改}
译文终稿:{结合以上意见，最终翻译得到的译文}
但是你给我的结果只需要译文终稿即可，不要任何多余的话语和原文，请保持译文的换行原样输出和----分割线`
)

type Config struct {
	token     string
	model     string
	ctxOffset int
	openaiCfg *openai.ClientConfig
}

type Client struct {
	cfg       *Config
	cli       *openai.Client
	logger    *zap.SugaredLogger
	req       openai.ChatCompletionRequest
	reqPrompt openai.ChatCompletionMessage
	queue     *Queue
}

func NewClient(token string, logger *zap.SugaredLogger, options ...func(config *Config)) *Client {
	openaiCfg := openai.DefaultConfig(token)
	cfg := &Config{
		token:     token,
		model:     openai.GPT3Dot5Turbo,
		openaiCfg: &openaiCfg,
	}
	for _, option := range options {
		option(cfg)
	}
	if cfg.openaiCfg.BaseURL == "" {
		logger.Infof("No url provided, Using default url: %s", defaultUrl)
		cfg.openaiCfg.BaseURL = defaultUrl
	}
	if cfg.ctxOffset == 0 {
		cfg.ctxOffset = 10
	}
	client := openai.NewClientWithConfig(*cfg.openaiCfg)
	req := openai.ChatCompletionRequest{
		Model:       cfg.model,
		Temperature: 0.2, //使用什么采样温度，介于 0 和 1 之间。较高的值（如 0.7）将使输出更加随机，而较低的值（如 0.2）将使其更加集中和确定性
	}
	reqPrompt := openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: defaultPrompt,
	}
	return &Client{
		cli:       client,
		cfg:       cfg,
		logger:    logger,
		req:       req,
		reqPrompt: reqPrompt,
		queue:     NewQueue(cfg.ctxOffset),
	}
}
func (c *Client) Translate(text []string, sourceLang string, targetLang string) ([]string, error) {
	c.queue.Enqueue(openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: before(text),
	})
	c.req.Messages = append([]openai.ChatCompletionMessage{c.reqPrompt}, c.queue.Get()...)
	resp, err := c.cli.CreateChatCompletion(context.Background(), c.req)
	if err != nil {
		c.logger.Errorw("ChatCompletion error", zap.Error(err))
		return nil, err
	}
	if len(resp.Choices) == 0 {
		c.logger.Infof("Translation result is empty")
		return nil, nil
	}
	c.queue.Enqueue(resp.Choices[0].Message)
	content := resp.Choices[0].Message.Content

	return handlerContent(content)
}

func handlerContent(content string) ([]string, error) {
	var r string
	if strings.HasPrefix(content, "译文终稿:") {
		r = strings.TrimPrefix(content, "译文终稿:")
	}
	if strings.HasPrefix(content, "----") {
		r = strings.TrimPrefix(content, "----")
	}

	return after(r), nil
}

func (c *Client) Close() error {
	return nil
}

func before(text []string) string {
	return strings.Join(text, "\n----\n")
}

func after(text string) []string {
	return strings.Split(text, "\n----\n")
}

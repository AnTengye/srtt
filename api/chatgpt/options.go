package chatgpt

func WithBaseUrl(baseUrl string) func(*Config) {
	return func(c *Config) {
		c.openaiCfg.BaseURL = baseUrl
	}
}

func WithModel(model string) func(*Config) {
	return func(c *Config) {
		if c.model == "" {
			return
		}
		c.model = model
	}
}

func WithCtxOffset(l int) func(*Config) {
	return func(c *Config) {
		c.ctxOffset = l
	}
}

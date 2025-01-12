package deeplx

import (
	"time"

	"github.com/go-resty/resty/v2"
)

func WithRetry(retry int) func(*resty.Client) {
	return func(c *resty.Client) {
		c.SetRetryCount(retry)
	}
}

// with retry timeout
func WithRetryWaitTime(waitTime time.Duration) func(*resty.Client) {
	return func(c *resty.Client) {
		c.SetRetryWaitTime(waitTime)
	}
}

func WithDebug(debug bool) func(*resty.Client) {
	return func(c *resty.Client) {
		c.SetDebug(debug)
	}
}

func WithBaseUrl(baseUrl string) func(*resty.Client) {
	return func(c *resty.Client) {
		c.SetBaseURL(baseUrl)
	}
}

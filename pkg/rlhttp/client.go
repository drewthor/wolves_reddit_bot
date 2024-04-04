package rlhttp

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"golang.org/x/time/rate"
)

func WithLeveledLogger(logger *slog.Logger) ClientOption {
	return func(c *Client) {
		c.client.Logger = logger
	}
}

func WithRequestTimeout(timeout time.Duration) ClientOption {
	return func(c *Client) {
		c.client.HTTPClient.Timeout = timeout
	}
}

func WithDefaultRetryWaitMax(max time.Duration) ClientOption {
	return func(c *Client) {
		c.client.RetryWaitMax = max
	}
}

func WithMaxRetries(max int) ClientOption {
	return func(c *Client) {
		c.client.RetryMax = max
	}
}

func WithRateLimiter(limiter *rate.Limiter) ClientOption {
	return func(c *Client) {
		c.limiter = limiter
	}
}

func WithTransport(transport http.RoundTripper) ClientOption {
	return func(c *Client) {
		c.client.HTTPClient.Transport = transport
	}
}

type ClientOption func(c *Client)

type Client struct {
	client  *retryablehttp.Client
	limiter *rate.Limiter
}

func NewClient(options ...ClientOption) *Client {
	retryClient := retryablehttp.NewClient()
	client := &Client{client: retryClient}

	for _, opt := range options {
		opt(client)
	}

	return client
}

func (c *Client) Do(req *retryablehttp.Request) (*http.Response, error) {
	ctx := req.Context()
	if c.limiter != nil {
		err := c.limiter.Wait(ctx)
		if err != nil {
			return nil, err
		}
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

package nba

import (
	"context"
	"net"
	"net/http"
	"time"
)

type Client interface {
	PlayByPlayForGame(ctx context.Context, gameID string, outputWriters ...OutputWriter) (PlayByPlay, error)
}

type client struct {
	client *http.Client
}

func NewClient() Client {
	c := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   time.Second,
			ResponseHeaderTimeout: time.Second,
		},
	}
	return &client{client: c}
}

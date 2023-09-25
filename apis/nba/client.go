package nba

import (
	"context"
	"net"
	"net/http"
	"time"
)

type Client interface {
	PlayByPlayForGame(ctx context.Context, gameID string, outputWriters ...OutputWriter) (PlayByPlay, error)
	GetCommonTeamInfo(ctx context.Context, leagueID string, teamID int) (TeamCommonInfo, error)
	GetBoxscoreSummary(ctx context.Context, gameID string, outputWriters ...OutputWriter) (BoxscoreSummary, error)
}

type client struct {
	client      *http.Client
	statsClient *http.Client
}

func NewClient() Client {
	dialer := &net.Dialer{Timeout: 10 * time.Second, KeepAlive: 10 * time.Second}
	c := &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			DialContext:           dialer.DialContext,
			TLSHandshakeTimeout:   10 * time.Second,
			ResponseHeaderTimeout: 10 * time.Second,
		},
	}
	statsC := &http.Client{
		Timeout: 10 * time.Second,
		Transport: nbaRoundTripper{
			r: &http.Transport{
				DialContext:           dialer.DialContext,
				TLSHandshakeTimeout:   10 * time.Second,
				ResponseHeaderTimeout: 10 * time.Second,
			},
		},
	}
	return &client{client: c, statsClient: statsC}
}

type nbaRoundTripper struct {
	r http.RoundTripper
}

func (n nbaRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Set("Host", "stats.nba.com")
	r.Header.Set("Connection", "keep-alive")
	r.Header.Set("Cache-Control", "no-cache")
	r.Header.Set("Upgrade-Insecure-Requests", "1")
	r.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/73.0.3683.86 Safari/537.36")
	r.Header.Set("Accept", "application/json, text/plain, */*")
	r.Header.Set("Accept-Encoding", "gzip, deflate, br")
	r.Header.Set("Accept-Language", "en-US,en;q=0.5")
	r.Header.Set("x-nba-stats-origin", "stats")
	r.Header.Set("x-nba-stats-token", "true")
	r.Header.Set("Referer", "https://stats.nba.com/")
	r.Header.Set("Pragma", "no-cache")

	return n.r.RoundTrip(r)
}

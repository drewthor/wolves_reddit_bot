package nba

import (
	"net/http"
	"time"

	"github.com/drewthor/wolves_reddit_bot/pkg/rlhttp"
	"golang.org/x/time/rate"
)

type Client struct {
	client      *rlhttp.Client
	statsClient *rlhttp.Client
	Cache       ObjectCacher
}

func NewClient(cache ObjectCacher, options ...rlhttp.ClientOption) Client {
	cOptions := options
	cOptions = append(cOptions, rlhttp.WithMaxRetries(2))
	cOptions = append(cOptions, rlhttp.WithDefaultRetryWaitMax(2*time.Second))
	cOptions = append(cOptions, rlhttp.WithRequestTimeout(5*time.Second))

	c := rlhttp.NewClient(cOptions...)

	limiter := rate.NewLimiter(rate.Every(time.Second), 3)
	statsOptions := options
	statsOptions = append(statsOptions, rlhttp.WithMaxRetries(2))
	statsOptions = append(statsOptions, rlhttp.WithDefaultRetryWaitMax(2*time.Second))
	statsOptions = append(statsOptions, rlhttp.WithRequestTimeout(5*time.Second))
	statsOptions = append(statsOptions, rlhttp.WithTransport(nbaRoundTripper{r: &http.Transport{
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 10 * time.Second,
	}}))
	statsOptions = append(statsOptions, rlhttp.WithRateLimiter(limiter))
	statsC := rlhttp.NewClient(statsOptions...)

	return Client{client: c, statsClient: statsC, Cache: cache}
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

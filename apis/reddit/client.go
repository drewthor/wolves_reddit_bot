package reddit

import (
	"context"
	"encoding/json"
	"golang.org/x/oauth2"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"
)

const authURI = "https://www.reddit.com/api/v1/authorize"
const authTokenURI = "https://www.reddit.com/api/v1/access_token"
const redirectURI = "https://www.reddit.com/r/timberwolves"
const userAgent = "Wolves Reddit Bot v0.1 by /u/SilverPenguino"

const configFile = "reddit_config.json"

type userConfig struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
}

type redditToken struct {
	Token                 string `json:"access_token"`
	SecondsTillExpiration int    `json:"expires_in"`
}

// custom http.RoundTripper for reddit's requirement of setting a custom
// user agent
type redditTransport struct {
	config *oauth2.Config
}

func (r *redditTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", userAgent)
	req.SetBasicAuth(r.config.ClientID, r.config.ClientSecret)
	return http.DefaultTransport.RoundTrip(req)
}

type Client struct {
	userConfig userConfig
	config     *oauth2.Config
	httpClient *http.Client
	token      *oauth2.Token
}

func (c *Client) initHTTPClient() {
	c.httpClient = &http.Client{
		Timeout: time.Second * 10,
	}
}

func (c *Client) loadConfiguration(file string) {
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		c.userConfig.Username = os.Getenv("username")
		c.userConfig.Password = os.Getenv("password")
		c.userConfig.ClientID = os.Getenv("clientID")
		c.userConfig.ClientSecret = os.Getenv("clientSecret")
	} else {
		jsonParser := json.NewDecoder(configFile)
		jsonParser.Decode(&c.userConfig)
	}
	c.config = &oauth2.Config{
		ClientID:     c.userConfig.ClientID,
		ClientSecret: c.userConfig.ClientSecret,
		Scopes:       []string{"submit"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  authURI,
			TokenURL: authTokenURI,
		},
		RedirectURL: redirectURI,
	}
}

func (c *Client) Authorize() {
	c.initHTTPClient()
	c.loadConfiguration(configFile)
	form := url.Values{
		"grant_type": {"password"},
		"username":   {c.userConfig.Username},
		"password":   {c.userConfig.Password},
	}

	req, err := http.NewRequest("POST", authTokenURI, strings.NewReader(form.Encode()))
	if err != nil {
		log.Fatal(err.Error())
	}
	req.Header.Set("User-Agent", userAgent)
	req.SetBasicAuth(c.config.ClientID, c.config.ClientSecret)
	response, err := c.httpClient.Do(req)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer response.Body.Close()
	rToken := redditToken{}
	if response.StatusCode == http.StatusOK {
		decoder := json.NewDecoder(response.Body)
		err = decoder.Decode(&rToken)
		if err != nil {
			log.Fatal(err.Error())
		}
		c.token = &oauth2.Token{}
		// Construct a custom http client that forces the user-agent and attach it
		// to the oauth2 context. https://github.com/golang/oauth2/issues/179
		client := &http.Client{
			Transport: &oauth2.Transport{
				Source: c.config.TokenSource(oauth2.NoContext, &oauth2.Token{
					AccessToken: rToken.Token,
				}),
				Base: &redditTransport{
					config: c.config,
				},
			},
		}
		ctx := context.WithValue(context.Background(), oauth2.HTTPClient, client)
		token, err := c.config.PasswordCredentialsToken(ctx, c.userConfig.Username, c.userConfig.Password)
		if err != nil {
			log.Fatal(err.Error())
		}
		c.token = token
		c.httpClient = c.config.Client(oauth2.NoContext, token)
	} else {
		log.Fatal(response.Body)
	}
}

func (c *Client) SubmitNewPost(subreddit, title, content string) {
	if c.httpClient == nil || c.token == nil {
		c.Authorize()
	}
	s := new(submit)
	s.ad = false
	s.apiType = "json"
	s.content = content
	s.kind = "self"
	s.nsfw = false
	s.sendReplies = false
	s.spoiler = false
	s.subreddit = subreddit
	s.title = title

	urlValues := url.Values{
		"ad":          {strconv.FormatBool(s.ad)},
		"api_type":    {s.apiType},
		"kind":        {s.kind},
		"nsfw":        {strconv.FormatBool(s.nsfw)},
		"sendreplies": {strconv.FormatBool(s.sendReplies)},
		"spoiler":     {strconv.FormatBool(s.spoiler)},
		"sr":          {s.subreddit},
		"text":        {s.content},
		"title":       {s.title},
	}

	request, err := http.NewRequest("POST", submitURI, strings.NewReader(urlValues.Encode()))
	if err != nil {
		log.Fatal(err)
	}
	request.PostForm = urlValues
	request.Header.Set("User-Agent", userAgent)

	response, err := c.httpClient.Do(request)
	if err != nil {
		log.Fatal(err)
	}
	if response.StatusCode != 200 {
		log.Fatal("Failed to submit post with status code: " + strconv.Itoa(response.StatusCode))
	}
}

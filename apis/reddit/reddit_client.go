package reddit

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	//"strconv"
	"strings"
	"time"
)

const RedditAuthorizationURI = "https://www.reddit.com/api/v1/access_token"

const RedditConfigFile = "reddit_config.json"

func loadConfiguration(file string) RedditConfig {
	var config RedditConfig
	configFile, err := os.Open(file)
	defer configFile.Close()
	if err != nil {
		log.Fatal(err.Error())
	}
	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&config)
	return config
}

type RedditConfig struct {
	Username     string `json:"username"`
	Password     string `json:"password"`
	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
}

/* type ExpireTime struct {
	time.Time
}

func (t *ExpireTime) UnmarshalJSON(b []byte) (err error) {
	expireStr := string(b[:])
	expireSeconds, err := strconv.Atoi(expireStr)
	*t = ExpireTime{time.Now().Add(time.Duration(expireSeconds) * time.Second)}
	return err
} */

type RedditToken struct {
	Token                 string `json:"access_token"`
	SecondsTillExpiration int    `json:"expires_in"`
	expirationTime        time.Time
}

type Client struct {
	config     RedditConfig
	httpClient *http.Client
	token      RedditToken
}

func (r *Client) initHttpClient() {
	r.httpClient = &http.Client{
		Timeout: time.Second * 10,
	}
}

func (r *Client) Authorize() {
	r.initHttpClient()
	r.config = loadConfiguration(RedditConfigFile)
	form := url.Values{
		"grant_type": {"password"},
		"username":   {r.config.Username},
		"password":   {r.config.Password},
	}

	req, err := http.NewRequest("POST", RedditAuthorizationURI, strings.NewReader(form.Encode()))
	if err != nil {
		log.Fatal(err.Error())
	}
	req.Header.Set("User-Agent", "Wolves Reddit Bot v0.1 by /u/SilverPenguino")
	req.SetBasicAuth(r.config.ClientID, r.config.ClientSecret)
	response, err := r.httpClient.Do(req)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer response.Body.Close()
	if response.StatusCode == http.StatusOK {
		r.token.expirationTime = time.Now()
		decoder := json.NewDecoder(response.Body)
		err = decoder.Decode(&r.token)
		if err != nil {
			log.Fatal(err.Error())
		}
		r.token.expirationTime.Add(time.Second * time.Duration(r.token.SecondsTillExpiration))
	} else {
		log.Fatal(response.Body)
	}
}

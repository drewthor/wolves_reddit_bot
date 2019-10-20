package reddit

const listingsURI = "https://oauth.reddit.com/by_id/"

type listingsResponse struct {
	DataNode struct {
		Listings []listingsChildren `json:"children"`
	} `json:"data"`
}

type listingsChildren struct {
	DataNode struct {
		Name string `json:"name"`
		URL  string `json:"url"`
	} `json:"data"`
}

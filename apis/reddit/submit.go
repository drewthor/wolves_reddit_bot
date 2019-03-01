package reddit

import ()

const submitURI = "https://oauth.reddit.com/api/submit"

type submit struct {
	ad          bool
	apiType     string
	content     string
	kind        string
	nsfw        bool
	sendReplies bool
	spoiler     bool
	subreddit   string
	title       string
}

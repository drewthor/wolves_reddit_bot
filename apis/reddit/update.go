package reddit

const updateUserTextURI = "https://oauth.reddit.com/api/editusertext"

type updateUserText struct {
	apiType       string
	content       string
	thingFullname string
}

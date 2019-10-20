package main

import (
	"fmt"

	"github.com/drewthor/wolves_reddit_bot/apis/reddit"
)

func main() {
	//var wg sync.WaitGroup
	//wg.Add(1)
	//go gt.CreateGameThread(nba.MinnesotaTimberwolves, &wg)
	//wg.Wait()
	redditClient := reddit.Client{}
	redditClient.Authorize()
	mappings := redditClient.GetThingURLs([]string{"t3_djea34", "t3_dif8n6"}, "timberwolves")
	fmt.Println(mappings)
}

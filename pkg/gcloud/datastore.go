package gcloud

import (
	"cloud.google.com/go/datastore"
	"context"
	"fmt"
	"log"
)

const (
	projectKey string = "TeamGameEvent"
	projectID  string = "wolvesredditbot"
)

type Datastore struct {
	ctx      context.Context
	dsClient *datastore.Client
}

func (d *Datastore) initClient() {
	d.ctx = context.Background()
	if d.dsClient == nil {
		client, err := datastore.NewClient(d.ctx, projectID)
		if err != nil {
			log.Println(err)
			log.Fatal("failed to create datastore client")
		}
		d.dsClient = client
	}
}

type TeamGameEvent struct {
	GameID         string
	TeamID         string
	PreGameThread  bool
	GameThread     bool
	PostGameThread bool
}

func MakeKeyName(gameID, teamID string) string {
	return gameID + teamID
}

func (d *Datastore) GetTeamGameEvent(gameID, teamID string) (TeamGameEvent, bool) {
	if d.dsClient == nil {
		d.initClient()
	}
	gameEvent := TeamGameEvent{}
	keyName := MakeKeyName(gameID, teamID)
	key := datastore.NameKey(projectKey, keyName, nil)
	if err := d.dsClient.Get(d.ctx, key, &gameEvent); err != nil {
		log.Println(fmt.Sprintf("failed to get GameEvent with ID: %s", gameID))
		log.Println(err)
		return gameEvent, false
	}
	return gameEvent, true
}

func (d *Datastore) SaveTeamGameEvent(gameEvent TeamGameEvent) {
	if d.dsClient == nil {
		d.initClient()
	}
	keyName := MakeKeyName(gameEvent.GameID, gameEvent.TeamID)
	key := datastore.NameKey(projectKey, keyName, nil)
	if _, err := d.dsClient.Put(d.ctx, key, &gameEvent); err != nil {
		log.Println("failed to save TeamGameEvent")
		log.Println(err)
	}
	log.Println("saved TeamGameEvent")
}

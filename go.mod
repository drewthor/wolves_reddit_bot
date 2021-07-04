module github.com/drewthor/wolves_reddit_bot

go 1.16

require (
	cloud.google.com/go v0.46.2 // indirect
	cloud.google.com/go/datastore v1.0.0
	github.com/go-chi/chi v1.5.1
	github.com/google/go-cmp v0.3.1 // indirect
	github.com/jackc/pgconn v1.8.1
	github.com/jackc/pgtype v1.7.0
	github.com/jackc/pgx/v4 v4.11.0
	go.opencensus.io v0.22.2 // indirect
	golang.org/x/exp v0.0.0-20190912063710-ac5d2bfcbfe0 // indirect
	golang.org/x/lint v0.0.0-20190930215403-16217165b5de // indirect
	golang.org/x/net v0.0.0-20210226172049-e18ecbb05110 // indirect
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	golang.org/x/sync v0.0.0-20190911185100-cd5d95a43a6e // indirect
	golang.org/x/tools v0.0.0-20200103221440-774c71fcf114 // indirect
	google.golang.org/api v0.10.0 // indirect
	google.golang.org/appengine v1.6.2 // indirect
	google.golang.org/grpc v1.26.0 // indirect
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.12.0

replace github.com/golang/lint v0.0.0-20190909230951-414d861bb4ac => golang.org/x/lint v0.0.0-20190909230951-414d861bb4ac

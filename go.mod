module github.com/drewthor/wolves_reddit_bot

go 1.16

require (
	cloud.google.com/go v0.88.0 // indirect
	cloud.google.com/go/datastore v1.1.0
	github.com/getsentry/sentry-go v0.12.0
	github.com/go-chi/chi v1.5.1
	github.com/go-co-op/gocron v1.9.0
	github.com/gofrs/uuid v4.0.0+incompatible // indirect
	github.com/google/uuid v1.3.0 // indirect
	github.com/jackc/pgproto3/v2 v2.0.7 // indirect
	github.com/jackc/pgx/v4 v4.11.0
	github.com/lib/pq v1.10.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/shopspring/decimal v1.2.0 // indirect
	github.com/sirupsen/logrus v1.8.1
	golang.org/x/net v0.0.0-20211013171255-e13a2654a71e // indirect
	golang.org/x/oauth2 v0.0.0-20210628180205-a41e5a781914
	golang.org/x/sys v0.0.0-20211013075003-97ac67df715c // indirect
	google.golang.org/api v0.51.0 // indirect
	google.golang.org/genproto v0.0.0-20211013025323-ce878158c4d4 // indirect
	google.golang.org/grpc v1.41.0 // indirect
)

replace git.apache.org/thrift.git => github.com/apache/thrift v0.12.0

replace github.com/golang/lint v0.0.0-20190909230951-414d861bb4ac => golang.org/x/lint v0.0.0-20190909230951-414d861bb4ac

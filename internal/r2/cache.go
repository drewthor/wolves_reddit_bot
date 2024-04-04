package r2

import (
	"context"
	"errors"
	"io"
	"net/http"

	awshttp "github.com/aws/aws-sdk-go-v2/aws/transport/http"
	"github.com/drewthor/wolves_reddit_bot/apis/nba"

	"github.com/drewthor/wolves_reddit_bot/apis/cloudflare"
	"github.com/drewthor/wolves_reddit_bot/util"
)

type R2ObjectCacher struct {
	R2Client cloudflare.Client
	Bucket   string
}

func (r R2ObjectCacher) GetObject(ctx context.Context, key string) ([]byte, error) {
	obj, err := r.R2Client.GetObject(ctx, r.Bucket, key)
	if err != nil {
		var respErr *awshttp.ResponseError
		if errors.As(err, &respErr) {
			if respErr.ResponseError.Response.StatusCode == http.StatusNotFound {
				return nil, nba.ErrNotFound
			}
		}
	}
	return obj, nil
}

func (r R2ObjectCacher) PutObject(ctx context.Context, key string, obj io.Reader) error {
	return r.R2Client.PutObject(ctx, r.Bucket, key, util.ContentTypeJSON, obj)
}

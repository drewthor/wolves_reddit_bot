package cloudflare

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// CreateBucket creates a bucket named bucket in R2
func (c Client) CreateBucket(ctx context.Context, bucket string) error {
	input := &s3.CreateBucketInput{Bucket: &bucket}
	_, err := c.client.CreateBucket(ctx, input)
	if err != nil {
		var bae *types.BucketAlreadyExists
		var baoby *types.BucketAlreadyOwnedByYou
		if errors.As(err, &bae) || errors.As(err, &baoby) {
			return nil // bucket already exists
		}
		return fmt.Errorf("failed to create request to create r2 bucket: %w", err)
	}

	return nil
}

// CreateObject creates an object with key objectKey in the R2 bucket
func (c Client) CreateObject(ctx context.Context, bucket, objectKey string, contentType string, objReader io.Reader) error {
	_, err := c.client.PutObject(ctx, &s3.PutObjectInput{Bucket: &bucket, Key: &objectKey, ContentType: &contentType, Body: objReader})
	if err != nil {
		return fmt.Errorf("failed to create request to create r2 object: %w", err)
	}

	return nil
}

// CreateObject gets an object with key objectKey from the R2 bucket and returns it
func (c Client) GetObject(ctx context.Context, bucket, objectKey string) ([]byte, error) {
	output, err := c.client.GetObject(ctx, &s3.GetObjectInput{Bucket: &bucket, Key: &objectKey})
	if err != nil {
		return nil, fmt.Errorf("failed to create request to get r2 object: %w", err)
	}
	defer output.Body.Close()

	b, err := io.ReadAll(output.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body when getting r2 object: %w", err)
	}

	return b, nil
}

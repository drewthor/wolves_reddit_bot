package cloudflare

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const baseURI = "https://%s.r2.cloudflarestorage.com"

func getBaseURI(accountID string) string {
	return fmt.Sprintf(baseURI, accountID)
}

type Client struct {
	client    *s3.Client
	accountID string
}

func NewClient(cloudflareAccountID string, cloudflareAccessKeyID string, cloudflareAccessKeySecret string) Client {
	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(sv, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: getBaseURI(cloudflareAccountID),
		}, nil
	})

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cloudflareAccessKeyID, cloudflareAccessKeySecret, "")),
	)
	if err != nil {
		//return err
	}

	s3Client := s3.NewFromConfig(cfg)

	return Client{
		accountID: cloudflareAccountID,
		client:    s3Client,
	}
}

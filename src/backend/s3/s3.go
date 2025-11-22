package s3

import (
	"context"
	"fmt"
	"infralog/config"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Backend struct {
	bucket string
	key    string
	region string
}

func New(cfg config.S3Config) *S3Backend {
	return &S3Backend{
		bucket: cfg.Bucket,
		key:    cfg.Key,
		region: cfg.Region,
	}
}

func (b *S3Backend) GetState() ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(b.region))
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	client := s3.NewFromConfig(cfg)

	result, err := client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(b.bucket),
		Key:    aws.String(b.key),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to download file from S3: %w", err)
	}
	defer result.Body.Close()

	body, err := io.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read file contents: %w", err)
	}

	return body, nil
}

func (b *S3Backend) Name() string {
	return "s3"
}

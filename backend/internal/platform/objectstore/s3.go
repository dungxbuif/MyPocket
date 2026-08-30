package objectstore

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"mypocket/internal/platform/config"
)

const smokeObjectKey = "platform/smoke.txt"

type S3Store struct {
	client *s3.Client
	bucket string
}

func NewS3(cfg config.Config) (*S3Store, error) {
	if strings.TrimSpace(cfg.S3Endpoint) == "" {
		return nil, fmt.Errorf("missing required config: S3_ENDPOINT")
	}
	if strings.TrimSpace(cfg.S3Bucket) == "" {
		return nil, fmt.Errorf("missing required config: S3_BUCKET")
	}
	if cfg.S3AccessKey == "" {
		return nil, fmt.Errorf("missing required config: S3_ACCESS_KEY")
	}
	if cfg.S3SecretKey == "" {
		return nil, fmt.Errorf("missing required config: S3_SECRET_KEY")
	}

	awsConfig, err := awscfg.LoadDefaultConfig(context.Background(),
		awscfg.WithRegion("us-east-1"),
		awscfg.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.S3AccessKey, cfg.S3SecretKey, "")),
	)
	if err != nil {
		return nil, fmt.Errorf("configure object store: %w", err)
	}

	client := s3.NewFromConfig(awsConfig, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(cfg.S3Endpoint)
		options.UsePathStyle = true
	})
	return &S3Store{client: client, bucket: cfg.S3Bucket}, nil
}

func (s *S3Store) PutSmokeObject(ctx context.Context) error {
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(smokeObjectKey),
		Body:   bytes.NewReader([]byte("mypocket object store smoke\n")),
	})
	if err != nil {
		return fmt.Errorf("object store smoke put failed")
	}
	return nil
}

func (s *S3Store) DeleteSmokeObject(ctx context.Context) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(smokeObjectKey),
	})
	if err != nil {
		return fmt.Errorf("object store smoke delete failed")
	}
	return nil
}

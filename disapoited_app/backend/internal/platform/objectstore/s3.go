package objectstore

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"mypocket/internal/platform/config"
)

const smokeObjectKey = "platform/smoke.txt"

type S3Store struct {
	client  *s3.Client
	presign *s3.PresignClient
	bucket  string
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
	return &S3Store{client: client, presign: s3.NewPresignClient(client), bucket: cfg.S3Bucket}, nil
}

func (s *S3Store) PresignPut(ctx context.Context, key string, contentType string, expires time.Duration) (string, error) {
	if expires <= 0 {
		expires = 10 * time.Minute
	}
	result, err := s.presign.PresignPutObject(ctx, &s3.PutObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key), ContentType: aws.String(contentType)}, s3.WithPresignExpires(expires))
	if err != nil {
		return "", fmt.Errorf("presign upload: %w", err)
	}
	return result.URL, nil
}

func (s *S3Store) PresignGet(ctx context.Context, key string, expires time.Duration) (string, error) {
	if expires <= 0 {
		expires = 10 * time.Minute
	}
	result, err := s.presign.PresignGetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)}, s3.WithPresignExpires(expires))
	if err != nil {
		return "", fmt.Errorf("presign download: %w", err)
	}
	return result.URL, nil
}

func (s *S3Store) DeleteObject(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)})
	if err != nil {
		return fmt.Errorf("delete object: %w", err)
	}
	return nil
}

func (s *S3Store) PutObject(ctx context.Context, key, contentType string, body io.Reader, size int64) error {
	if size < 0 || size > 20<<20 {
		return fmt.Errorf("object size is outside the allowed range")
	}
	payload, err := io.ReadAll(io.LimitReader(body, size+1))
	if err != nil || int64(len(payload)) != size {
		return fmt.Errorf("object body does not match declared size")
	}
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key), ContentType: aws.String(contentType), Body: bytes.NewReader(payload), ContentLength: aws.Int64(size)})
	if err != nil {
		return fmt.Errorf("put object failed")
	}
	return nil
}

func (s *S3Store) GetObject(ctx context.Context, key string, maxSize int64) (io.ReadCloser, int64, error) {
	result, err := s.client.GetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(s.bucket), Key: aws.String(key)})
	if err != nil {
		return nil, 0, fmt.Errorf("get object failed")
	}
	size := aws.ToInt64(result.ContentLength)
	if size < 0 || size > maxSize {
		result.Body.Close()
		return nil, 0, fmt.Errorf("object size is outside the allowed range")
	}
	return result.Body, size, nil
}

func (s *S3Store) DeletePrefix(ctx context.Context, prefix string) error {
	if !strings.HasPrefix(prefix, "users/") || !strings.HasSuffix(prefix, "/") {
		return fmt.Errorf("invalid private object prefix")
	}
	var token *string
	for {
		listed, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{Bucket: aws.String(s.bucket), Prefix: aws.String(prefix), ContinuationToken: token})
		if err != nil {
			return fmt.Errorf("list private objects failed")
		}
		for _, object := range listed.Contents {
			if _, err = s.client.DeleteObject(ctx, &s3.DeleteObjectInput{Bucket: aws.String(s.bucket), Key: object.Key}); err != nil {
				return fmt.Errorf("delete private object failed")
			}
		}
		if !aws.ToBool(listed.IsTruncated) {
			return nil
		}
		token = listed.NextContinuationToken
	}
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

package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Parallels/pd-ai-agents-registry/internal/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/aws/smithy-go"
)

type S3Service struct {
	client *s3.Client
	bucket string
}

func NewS3Service(cfg config.S3Config) (*S3Service, error) {
	var options []func(*awsconfig.LoadOptions) error

	// Add region
	options = append(options, awsconfig.WithRegion(cfg.Region))

	// Add credentials
	options = append(options, awsconfig.WithCredentialsProvider(
		credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		),
	))

	// Load the config with all options
	awsCfg, err := awsconfig.LoadDefaultConfig(context.TODO(), options...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create S3 client with custom endpoint and path style addressing
	clientOptions := []func(*s3.Options){
		func(o *s3.Options) {
			if cfg.Endpoint != "" {
				o.BaseEndpoint = aws.String(cfg.Endpoint)
				// Force path style addressing for custom endpoints (like MinIO)
				o.UsePathStyle = true
				// Disable SSL verification for local development
				if strings.Contains(cfg.Endpoint, "localhost") || strings.Contains(cfg.Endpoint, "127.0.0.1") {
					o.EndpointOptions.DisableHTTPS = true
				}
			}
		},
	}

	client := s3.NewFromConfig(awsCfg, clientOptions...)

	// Verify bucket exists and is accessible
	_, err = client.HeadBucket(context.TODO(), &s3.HeadBucketInput{
		Bucket: aws.String(cfg.Bucket),
	})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			return nil, fmt.Errorf("bucket access error: %s: %s", apiErr.ErrorCode(), apiErr.ErrorMessage())
		}
		return nil, fmt.Errorf("failed to access bucket: %w", err)
	}

	return &S3Service{
		client: client,
		bucket: cfg.Bucket,
	}, nil
}

func (s *S3Service) Upload(ctx context.Context, key string, reader io.Reader) error {
	input := &s3.PutObjectInput{
		Bucket:       aws.String(s.bucket),
		Key:          aws.String(key),
		Body:         reader,
		ContentType:  aws.String("application/octet-stream"),
		StorageClass: types.StorageClassStandard,
	}

	_, err := s.client.PutObject(ctx, input)
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			return fmt.Errorf("s3 upload failed: %s: %s", apiErr.ErrorCode(), apiErr.ErrorMessage())
		}
		return fmt.Errorf("s3 upload failed: %w", err)
	}
	return nil
}

func (s *S3Service) Download(ctx context.Context, key string) (io.ReadCloser, error) {
	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	result, err := s.client.GetObject(ctx, input)
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			if apiErr.ErrorCode() == "NoSuchKey" {
				return nil, fmt.Errorf("object not found: %s", key)
			}
			return nil, fmt.Errorf("s3 download failed: %s: %s", apiErr.ErrorCode(), apiErr.ErrorMessage())
		}
		return nil, fmt.Errorf("s3 download failed: %w", err)
	}

	return result.Body, nil
}

func (s *S3Service) Size(ctx context.Context, key string) (int64, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	result, err := s.client.HeadObject(ctx, input)
	if err != nil {
		return 0, fmt.Errorf("s3 size failed: %w", err)
	}

	return *result.ContentLength, nil
}

func (s *S3Service) Delete(ctx context.Context, key string) error {
	input := &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	_, err := s.client.DeleteObject(ctx, input)
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			return fmt.Errorf("s3 delete failed: %s: %s", apiErr.ErrorCode(), apiErr.ErrorMessage())
		}
		return fmt.Errorf("s3 delete failed: %w", err)
	}

	// Wait until the object is deleted
	waiter := s3.NewObjectNotExistsWaiter(s.client)
	err = waiter.Wait(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, 30*time.Second)
	if err != nil {
		return fmt.Errorf("error waiting for object deletion: %w", err)
	}

	return nil
}

func (s *S3Service) GetSignedURL(ctx context.Context, key string, duration time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	input := &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	request, err := presignClient.PresignGetObject(ctx, input, func(options *s3.PresignOptions) {
		options.Expires = duration
	})
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			return "", fmt.Errorf("failed to generate presigned URL: %s: %s", apiErr.ErrorCode(), apiErr.ErrorMessage())
		}
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return request.URL, nil
}

func (s *S3Service) Exists(ctx context.Context, key string) (bool, error) {
	input := &s3.HeadObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}

	_, err := s.client.HeadObject(ctx, input)
	if err != nil {
		var apiErr smithy.APIError
		if errors.As(err, &apiErr) {
			if apiErr.ErrorCode() == "NotFound" {
				return false, nil
			}
			return false, fmt.Errorf("s3 head object failed: %s: %s", apiErr.ErrorCode(), apiErr.ErrorMessage())
		}
		return false, fmt.Errorf("s3 head object failed: %w", err)
	}

	return true, nil
}

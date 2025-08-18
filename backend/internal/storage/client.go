package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
)

type B2Client struct {
	bucketName string
	client     *s3.Client
	logger     *zap.Logger
}

func NewB2Client(bucketName, keyID, applicationKey, endpoint, region string, logger *zap.Logger) (*B2Client, error) {
	cfg, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(region),
		config.WithCredentialsProvider(
			credentials.NewStaticCredentialsProvider(
				keyID,
				applicationKey,
				"",
			),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config, %v", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})

	return &B2Client{
		client:     client,
		bucketName: bucketName,
		logger:     logger,
	}, nil
}

func (b *B2Client) UploadFile(ctx context.Context, file io.Reader, objectKey string, metadata map[string]string) error {
	_, err := b.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:   aws.String(b.bucketName),
		Key:      aws.String(objectKey),
		Body:     file,
		Metadata: metadata,
	})
	if err != nil {
		b.logger.Error("failed to upload file", zap.String("objectKey", objectKey), zap.Error(err))
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

func (b *B2Client) DeleteFile(ctx context.Context, objectKey string) error {
	_, err := b.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(b.bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		b.logger.Error("failed to delete file", zap.String("objectKey", objectKey), zap.Error(err))
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

func (b *B2Client) GetFileURL(ctx context.Context, objectKey string) (string, error) {
	// Create a presign client
	presignClient := s3.NewPresignClient(b.client)

	// Create a presigned URL for GetObject with 15 minutes expiration
	presignResult, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(b.bucketName),
		Key:    aws.String(objectKey),
	}, func(opts *s3.PresignOptions) {
		opts.Expires = 15 * time.Minute
	})
	if err != nil {
		b.logger.Error("failed to generate presigned URL", zap.String("objectKey", objectKey), zap.Error(err))
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignResult.URL, nil
}

func (b *B2Client) GetFileMetadata(ctx context.Context, objectKey string) (map[string]string, error) {
	result, err := b.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(b.bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		b.logger.Error("failed to get file metadata", zap.String("objectKey", objectKey), zap.Error(err))
		return nil, fmt.Errorf("failed to get file metadata: %w", err)
	}

	return result.Metadata, nil
}

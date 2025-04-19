package storage

import (
	"context"
	"fmt"
	"mime/multipart"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"go.uber.org/zap"
)

type StorageClient interface {
	UploadFile(ctx context.Context, file multipart.File, objectKey string, metadata map[string]string) error
	// DownloadFile(ctx context.Context, objectKey string, destinationPath string) error
	// DeleteFile(ctx context.Context, objectKey string) error
	// ListFiles(ctx context.Context, prefix string) ([]string, error)
	// GetFileURL(ctx context.Context, objectKey string) (string, error)
	// GetFileMetadata(ctx context.Context, objectKey string) (map[string]string, error)
}

type B2Client struct {
	bucketName string
	client     *s3.Client
	logger     *zap.Logger
}

func NewB2Client(bucketName, keyID, applicationKey, endpoint, region string, logger *zap.Logger) (StorageClient, error) {
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
func (b *B2Client) UploadFile(ctx context.Context, file multipart.File, objectKey string, metadata map[string]string) error {
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

func (b *B2Client) DownloadFile(ctx context.Context, objectKey string, destinationPath string) error {
	// Implement the download logic here
	return nil
}

func (b *B2Client) DeleteFile(ctx context.Context, objectKey string) error {
	// Implement the delete logic here
	return nil
}

func (b *B2Client) ListFiles(ctx context.Context, prefix string) ([]string, error) {
	// Implement the list files logic here
	return nil, nil
}

func (b *B2Client) GetFileURL(ctx context.Context, objectKey string) (string, error) {
	// Implement the get file URL logic here
	return "", nil
}

func (b *B2Client) GetFileMetadata(ctx context.Context, objectKey string) (map[string]string, error) {
	// Implement the get file metadata logic here
	return nil, nil
}

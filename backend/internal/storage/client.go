package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"mime"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/eaguilar88/deu/internal/entities"
	"go.uber.org/zap"
)

type B2Client struct {
	bucketName string
	baseURL    string
	client     *s3.Client
	logger     *zap.Logger
}

func NewB2Client(bucketName, keyID, applicationKey, endpoint, region, baseURL string, logger *zap.Logger) (*B2Client, error) {
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
		baseURL:    baseURL,
		logger:     logger,
	}, nil
}

func (b *B2Client) UploadFile(ctx context.Context, files []*entities.File) error {
	if len(files) == 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	b.logger.Debug("starting file upload and save",
		zap.Int("file_count", len(files)),
		zap.String("action", "upload_and_save"),
	)

	for _, file := range files {
		if err := validateFile(file); err != nil {
			b.logger.Error("invalid file",
				zap.Error(err),
				zap.String("file_key", file.Key),
				zap.String("action", "validate_file"),
			)
			return fmt.Errorf("invalid file %s: %w", file.Key, err)
		}

		contentType := mime.TypeByExtension(filepath.Ext(file.Name))
		if contentType == "" {
			contentType = "application/octet-stream"
		}

		_, err := b.client.PutObject(ctx, &s3.PutObjectInput{
			Bucket:             aws.String(b.bucketName),
			Key:                aws.String(file.Key),
			Body:               file.Body,
			Metadata:           file.MetaData,
			ContentType:        aws.String(contentType),
			ContentDisposition: aws.String("inline"),
		})
		if err != nil {
			b.logger.Error("failed to upload file", zap.String("objectKey", file.Key), zap.Error(err))
			return fmt.Errorf("failed to upload file: %w", err)
		}
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

func (b *B2Client) GetFileURL(_ context.Context, objectKey string) (string, error) {
	if b.baseURL != "" {
		return fmt.Sprintf("%s/files/%s", b.baseURL, objectKey), nil
	}

	// Fallback: generate a pre-signed URL when no base URL is configured
	presignClient := s3.NewPresignClient(b.client)
	presignResult, err := presignClient.PresignGetObject(context.Background(), &s3.GetObjectInput{
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

func (b *B2Client) GetObject(ctx context.Context, objectKey string) (io.ReadCloser, string, error) {
	result, err := b.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(b.bucketName),
		Key:    aws.String(objectKey),
	})
	if err != nil {
		b.logger.Error("failed to get object", zap.String("objectKey", objectKey), zap.Error(err))
		return nil, "", fmt.Errorf("failed to get object: %w", err)
	}

	contentType := aws.ToString(result.ContentType)
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	return result.Body, contentType, nil
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

func validateFile(file *entities.File) error {
	if file == nil {
		return errors.New("file is nil")
	}
	if file.Body == nil {
		return errors.New("file body is nil")
	}
	if file.Key == "" {
		return errors.New("file key is empty")
	}
	if file.Name == "" {
		return errors.New("file name is empty")
	}
	if file.Purpose == "" {
		return errors.New("file purpose is empty")
	}
	return nil
}

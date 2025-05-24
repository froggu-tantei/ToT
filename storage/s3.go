package storage

import (
	"context"
	"mime/multipart"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// S3Storage implements FileStorage for AWS S3
type S3Storage struct {
	Client     *s3.Client
	BucketName string
	Region     string
	BaseURL    string
}

// NewS3Storage creates a new S3Storage instance
func NewS3Storage(bucketName, region, baseURL string) (*S3Storage, error) {
	// Load AWS configuration
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(region),
	)
	if err != nil {
		return nil, err
	}

	// Create S3 client
	client := s3.NewFromConfig(cfg)

	return &S3Storage{
		Client:     client,
		BucketName: bucketName,
		Region:     region,
		BaseURL:    baseURL,
	}, nil
}

// Store uploads a file to S3 and returns its public URL
func (s *S3Storage) Store(file multipart.File, filename string) (string, error) {
	ctx := context.Background()

	// Upload the file to S3
	_, err := s.Client.PutObject(ctx, &s3.PutObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(filename),
		Body:   file,
	})
	if err != nil {
		return "", err
	}

	// Return the path to the file
	return "/" + filename, nil
}

// Delete removes a file from S3
func (s *S3Storage) Delete(path string) error {
	ctx := context.Background()

	// Remove leading slash if present
	if path != "" && path[0] == '/' {
		path = path[1:]
	}

	// Delete the file from S3
	_, err := s.Client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.BucketName),
		Key:    aws.String(path),
	})
	return err
}

// GetPublicURL returns the public URL for a stored file
func (s *S3Storage) GetPublicURL(path string) string {
	// If a custom base URL is provided (like CloudFront), use it
	if s.BaseURL != "" {
		if path != "" && path[0] != '/' {
			path = "/" + path
		}
		return s.BaseURL + path
	}

	// Remove leading slash if present
	if path != "" && path[0] == '/' {
		path = path[1:]
	}

	// Return the standard S3 URL
	return "https://" + s.BucketName + ".s3." + s.Region + ".amazonaws.com/" + path
}

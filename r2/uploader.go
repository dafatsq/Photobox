package r2

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// Config holds the Cloudflare R2 credentials entered by the admin.
type Config struct {
	AccountID       string
	AccessKeyID     string
	SecretAccessKey string
	BucketName      string
	PublicBaseURL   string // e.g. "https://pub-xxx.r2.dev" or custom domain
}

// IsConfigured returns true if all required fields are set.
func (c Config) IsConfigured() bool {
	return c.AccountID != "" && c.AccessKeyID != "" &&
		c.SecretAccessKey != "" && c.BucketName != "" && c.PublicBaseURL != ""
}

// UploadPhoto uploads the image at filePath to Cloudflare R2 and returns
// the full public URL of the uploaded object.
func UploadPhoto(cfg Config, filePath string) (string, error) {
	if !cfg.IsConfigured() {
		return "", fmt.Errorf("R2 is not configured — please enter credentials in the admin panel")
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read image file: %w", err)
	}

	// Generate a unique object key based on timestamp + original filename
	key := fmt.Sprintf("photos/%s_%s", time.Now().Format("20060102_150405"), filepath.Base(filePath))

	// Build the S3-compatible client pointing at Cloudflare R2
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID)

	client := s3.NewFromConfig(aws.Config{
		Region: "auto",
		Credentials: credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.SecretAccessKey,
			"",
		),
		EndpointResolverWithOptions: aws.EndpointResolverWithOptionsFunc(
			func(service, region string, opts ...interface{}) (aws.Endpoint, error) {
				return aws.Endpoint{URL: endpoint}, nil
			},
		),
	})

	contentType := "image/png"
	if filepath.Ext(filePath) == ".jpg" || filepath.Ext(filePath) == ".jpeg" {
		contentType = "image/jpeg"
	}

	_, err = client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket:      aws.String(cfg.BucketName),
		Key:         aws.String(key),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("R2 upload failed: %w", err)
	}

	// Construct the public URL
	baseURL := cfg.PublicBaseURL
	// Trim trailing slash
	for len(baseURL) > 0 && baseURL[len(baseURL)-1] == '/' {
		baseURL = baseURL[:len(baseURL)-1]
	}
	publicURL := fmt.Sprintf("%s/%s", baseURL, key)
	return publicURL, nil
}

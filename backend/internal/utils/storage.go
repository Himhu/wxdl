package utils

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

type ObjectStorageConfig struct {
	Enabled      bool
	Provider     string
	Endpoint     string
	Bucket       string
	AccessKeyID  string
	SecretKey    string
	Region       string
	CustomDomain string
	PathPrefix   string
}

type UploadResult struct {
	Key string
	URL string
}

func UploadToS3Compatible(ctx context.Context, cfg ObjectStorageConfig, fileData io.Reader, contentType string, scene string) (*UploadResult, error) {
	if !cfg.Enabled {
		return nil, fmt.Errorf("对象存储未开启")
	}

	region := cfg.Region
	if region == "" {
		region = "us-east-1"
	}

	client := s3.New(s3.Options{
		Region: region,
		Credentials: credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID, cfg.SecretKey, "",
		),
		BaseEndpoint: aws.String(cfg.Endpoint),
		UsePathStyle: true,
	})

	ext := ".jpg"
	if strings.Contains(contentType, "png") {
		ext = ".png"
	} else if strings.Contains(contentType, "gif") {
		ext = ".gif"
	} else if strings.Contains(contentType, "webp") {
		ext = ".webp"
	}

	now := time.Now()
	objectKey := path.Join(
		cfg.PathPrefix,
		scene,
		now.Format("2006/01/02"),
		uuid.New().String()+ext,
	)
	objectKey = strings.TrimPrefix(objectKey, "/")

	data, err := io.ReadAll(fileData)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	_, err = client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(cfg.Bucket),
		Key:         aws.String(objectKey),
		Body:        bytes.NewReader(data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return nil, fmt.Errorf("上传失败: %w", err)
	}

	var fileURL string
	if cfg.CustomDomain != "" {
		domain := strings.TrimRight(cfg.CustomDomain, "/")
		fileURL = domain + "/" + objectKey
	} else {
		endpoint := strings.TrimRight(cfg.Endpoint, "/")
		fileURL = endpoint + "/" + cfg.Bucket + "/" + objectKey
	}

	return &UploadResult{Key: objectKey, URL: fileURL}, nil
}

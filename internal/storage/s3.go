package storage

import (
	"context"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// S3Storage S3 兼容存储实现
type S3Storage struct {
	client *minio.Client
	bucket string
}

func NewS3Storage(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*S3Storage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	// 确保 bucket 存在
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, err
	}
	if !exists {
		err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, err
		}
	}

	return &S3Storage{client: client, bucket: bucket}, nil
}

func (s *S3Storage) Upload(path string, reader io.Reader, size int64) error {
	ctx := context.Background()
	_, err := s.client.PutObject(ctx, s.bucket, path, reader, size, minio.PutObjectOptions{})
	return err
}

func (s *S3Storage) Download(path string) (io.ReadCloser, error) {
	ctx := context.Background()
	return s.client.GetObject(ctx, s.bucket, path, minio.GetObjectOptions{})
}

func (s *S3Storage) Delete(path string) error {
	ctx := context.Background()
	return s.client.RemoveObject(ctx, s.bucket, path, minio.RemoveObjectOptions{})
}

func (s *S3Storage) Exists(path string) (bool, error) {
	ctx := context.Background()
	_, err := s.client.StatObject(ctx, s.bucket, path, minio.StatObjectOptions{})
	if err != nil {
		// Check if error is "NoSuchKey"
		if minio.ToErrorResponse(err).Code == "NoSuchKey" {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (s *S3Storage) GetSize(path string) (int64, error) {
	ctx := context.Background()
	info, err := s.client.StatObject(ctx, s.bucket, path, minio.StatObjectOptions{})
	if err != nil {
		return 0, err
	}
	return info.Size, nil
}

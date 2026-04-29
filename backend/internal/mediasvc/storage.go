package mediasvc

import (
	"context"
	"fmt"
	"io"
	"path"
	"strings"
	"time"

	"github.com/bitik/backend/internal/config"
	"github.com/minio/minio-go/v7"
)

type Storage interface {
	Put(ctx context.Context, objectKey, contentType string, size int64, body io.Reader) (string, error)
	PresignPut(ctx context.Context, objectKey, contentType string, expiry time.Duration) (string, string, error)
	Get(ctx context.Context, objectKey string, maxBytes int64) ([]byte, string, int64, error)
	Delete(ctx context.Context, objectKey string) error
	Bucket() string
	PublicURL(objectKey string) string
}

type minioStorage struct {
	client *minio.Client
	cfg    config.StorageConfig
}

func NewMinioStorage(client *minio.Client, cfg config.StorageConfig) Storage {
	if client == nil {
		return nil
	}
	return &minioStorage{client: client, cfg: cfg}
}

func (s *minioStorage) Bucket() string {
	return s.cfg.Bucket
}

func (s *minioStorage) Put(ctx context.Context, objectKey, contentType string, size int64, body io.Reader) (string, error) {
	_, err := s.client.PutObject(ctx, s.cfg.Bucket, objectKey, body, size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", err
	}
	return s.PublicURL(objectKey), nil
}

func (s *minioStorage) PresignPut(ctx context.Context, objectKey, contentType string, expiry time.Duration) (string, string, error) {
	u, err := s.client.PresignedPutObject(ctx, s.cfg.Bucket, objectKey, expiry)
	if err != nil {
		return "", "", err
	}
	return u.String(), s.PublicURL(objectKey), nil
}

func (s *minioStorage) Get(ctx context.Context, objectKey string, maxBytes int64) ([]byte, string, int64, error) {
	obj, err := s.client.GetObject(ctx, s.cfg.Bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, "", 0, err
	}
	defer obj.Close()

	info, err := obj.Stat()
	if err != nil {
		return nil, "", 0, err
	}
	if info.Size > maxBytes {
		return nil, info.ContentType, info.Size, errTooLarge
	}
	body, err := io.ReadAll(io.LimitReader(obj, maxBytes+1))
	if err != nil {
		return nil, info.ContentType, info.Size, err
	}
	if int64(len(body)) > maxBytes {
		return nil, info.ContentType, info.Size, errTooLarge
	}
	return body, info.ContentType, info.Size, nil
}

func (s *minioStorage) Delete(ctx context.Context, objectKey string) error {
	return s.client.RemoveObject(ctx, s.cfg.Bucket, objectKey, minio.RemoveObjectOptions{})
}

func (s *minioStorage) PublicURL(objectKey string) string {
	scheme := "http"
	if s.cfg.UseSSL {
		scheme = "https"
	}
	base := fmt.Sprintf("%s://%s", scheme, strings.TrimRight(s.cfg.Endpoint, "/"))
	return base + "/" + path.Join(s.cfg.Bucket, objectKey)
}

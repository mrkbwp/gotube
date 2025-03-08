package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioClient struct {
	client         *minio.Client
	internalClient *minio.Client
	publicEndpoint string
	publicURL      string
}

func NewMinioClient(internalEndpoint, publicEndpoint, accessKey, secretKey string, useSSL, storageUseSSL bool) (*MinioClient, error) {
	// Клиент для внутренних операций (загрузка, удаление и т.д.)
	internalClient, err := minio.New(internalEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create internal MinIO client: %w", err)
	}

	// Клиент для генерации публичных URL
	publicClient, err := minio.New(publicEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: storageUseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create public MinIO client: %w", err)
	}

	publicURL := fmt.Sprintf("https://%s", publicEndpoint)
	if !storageUseSSL {
		publicURL = fmt.Sprintf("http://%s", publicEndpoint)
	}

	return &MinioClient{
		client:         publicClient,
		internalClient: internalClient,
		publicEndpoint: publicEndpoint,
		publicURL:      publicURL,
	}, nil
}

// EnsureBucketExists проверяет и создает бакет используя внутренний клиент
func (m *MinioClient) EnsureBucketExists(ctx context.Context, bucketName string) error {
	exists, err := m.internalClient.BucketExists(ctx, bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = m.internalClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return nil
}

// UploadFile использует внутренний клиент для загрузки
func (m *MinioClient) UploadFile(ctx context.Context, bucketName, objectName string, reader io.Reader) error {
	if err := m.EnsureBucketExists(ctx, bucketName); err != nil {
		return err
	}

	_, err := m.internalClient.PutObject(ctx, bucketName, objectName, reader, -1, minio.PutObjectOptions{
		ContentType: "application/octet-stream",
	})
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}

	return nil
}

// GetFileURL использует публичный клиент для генерации URL
func (m *MinioClient) GetFileURL(ctx context.Context, bucketName, objectName string, expires int) (string, error) {
	url, err := m.client.PresignedGetObject(ctx, bucketName, objectName, time.Duration(expires), nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return url.String(), nil
}

// GetPublicURL возвращает публичный URL для доступа к файлу
func (m *MinioClient) GetPublicURL(bucketName, objectName string) string {
	return fmt.Sprintf("%s/%s/%s", m.publicURL, bucketName, objectName)
}

// DeleteFile использует внутренний клиент для удаления
func (m *MinioClient) DeleteFile(ctx context.Context, bucketName, objectName string) error {
	err := m.internalClient.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// DownloadFile использует внутренний клиент для скачивания
func (m *MinioClient) DownloadFile(ctx context.Context, bucketName, objectName string, filePath string) error {
	err := m.internalClient.FGetObject(ctx, bucketName, objectName, filePath, minio.GetObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to download file: %w", err)
	}
	return nil
}

// GetPermanentURL использует публичный клиент для генерации постоянного URL
func (m *MinioClient) GetPermanentURL(ctx context.Context, bucketName, objectName string) (string, error) {
	url, err := m.client.PresignedGetObject(ctx, bucketName, objectName, time.Hour*24*365*10, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate permanent URL: %w", err)
	}
	return url.String(), nil
}

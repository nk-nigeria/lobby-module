package objectstorage

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioWrapper struct {
	minioClient *minio.Client
	init        bool
}

func NewMinioWrapper(endpoint, accessKeyID, secretAccessKey string, useSSL bool) (*MinioWrapper, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}
	w := &MinioWrapper{}
	w.minioClient = minioClient
	w.init = true
	return w, nil
}

func (w *MinioWrapper) MakeBucket(bucketName string) error {
	if !w.init {
		return fmt.Errorf("Minio wrapper not init")
	}
	ctx := context.Background()
	location := "us-east-1"
	err := w.minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{Region: location})
	return err
}
func (w *MinioWrapper) PresignGetObject(bucketName string, objectName string, expiry time.Duration, params map[string]interface{}) (string, error) {
	// reqParams := make(url.Values)
	// reqParams.Set("response-content-disposition", fmt.Sprintf("attachment; filename=\"%s\"", objectName))
	presignedURL, err := w.minioClient.PresignedGetObject(
		context.Background(), bucketName,
		objectName,
		expiry,
		nil)
	if err != nil {
		return "", err
	}
	return url.PathUnescape(presignedURL.String())
}

func (w *MinioWrapper) PresigPutObject(bucketName string, objectName string, expiry time.Duration, params map[string]interface{}) (string, error) {
	presignedURL, err := w.minioClient.PresignedPutObject(
		context.Background(),
		bucketName,
		objectName,
		expiry)
	if err != nil {
		return "", err
	}
	return url.PathUnescape(presignedURL.String())
}

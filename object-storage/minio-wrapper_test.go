package objectstorage

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const (
	MinioHost      = "127.0.0.1:9000"
	MinioKey       = "minio"
	MinioAccessKey = "12345678"
)

func TestMakeBucket(t *testing.T) {
	client, err := NewMinioWrapper(MinioHost, MinioKey, MinioAccessKey, false)
	assert.Nil(t, err)
	assert.NotNil(t, client)
	err = client.MakeBucket(fmt.Sprintf("bucket-test-%d", time.Now().UnixMilli()))
	assert.Nil(t, err)
}

func TestPreSignPut(t *testing.T) {
	client, err := NewMinioWrapper(MinioHost, MinioKey, MinioAccessKey, false)
	assert.Nil(t, err)
	assert.NotNil(t, client)
	bucketName := "bucket-test"
	fileUpload := "/home/sondq/Downloads/index.m3u8"
	putUrl, err := client.PresigPutObject(bucketName, "index.m3u8", 1*time.Hour, nil)
	assert.Nil(t, err)
	assert.NotEmpty(t, putUrl)
	err = UploadFile(fileUpload, putUrl)
	assert.Nil(t, err)
}

func TestPreSignGet(t *testing.T) {
	client, err := NewMinioWrapper(MinioHost, MinioKey, MinioAccessKey, false)
	assert.Nil(t, err)
	assert.NotNil(t, client)
	bucketName := "bucket-test"
	fileDownload := "index.m3u8"
	putUrl, err := client.PresignGetObject(bucketName, fileDownload, 1*time.Hour, nil)
	assert.Nil(t, err)
	assert.NotEmpty(t, putUrl)
	err = DownloadFile(fileDownload, putUrl)
	if err != nil {
		t.Error(err)
	}
	assert.Nil(t, err)
}

func UploadFile(fileUpload string, urlUpload string) error {
	f, err := os.OpenFile(fileUpload, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	defer f.Close()
	req, err := http.NewRequest("PUT", urlUpload, f)
	if err != nil {
		return err
	}
	fileStat, err := os.Stat(fileUpload)
	if err != nil {
		return err
	}
	// contentLength := strconv.FormatInt(fileStat.Size(), 10)
	// req.Header.Set("Content-Length", contentLength)
	// req.Header.Set("Content-Type", "multipart/form-data")
	req.ContentLength = fileStat.Size()

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	if res == nil {
		return fmt.Errorf("Response is nil")
	}
	if res.StatusCode > 300 {
		bodyData, _ := io.ReadAll(res.Body)
		return fmt.Errorf("Status not ok: %d, body: %s", res.StatusCode, string(bodyData))
	}
	defer res.Body.Close()

	return nil
}

func DownloadFile(fileDownload string, urlDownload string) error {
	req, err := http.NewRequest("GET", urlDownload, nil)
	if err != nil {
		return err
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return err
	}
	if res == nil {
		return fmt.Errorf("Response is nil")
	}
	bodyData, _ := io.ReadAll(res.Body)

	if res.StatusCode > 300 {
		return fmt.Errorf("Status not ok: %d, body: %s", res.StatusCode, string(bodyData))
	}
	fmt.Printf("Body: %s", string(bodyData))
	defer res.Body.Close()

	return nil
}

func TestUpload(t *testing.T) {
	url := "http://172.17.0.1:9000/avatar/30432ce8-8f28-43a2-8002-0d87b5804722_image?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Credential=minio/20220415/us-east-1/s3/aws4_request&X-Amz-Date=20220415T140255Z&X-Amz-Expires=3600&X-Amz-SignedHeaders=host&X-Amz-Signature=1b67e5b678c9bcfb8657cc52b43e3531864fd4b8bcac53b9b8fc09f3747e3df7"
	fileUpload := "/home/sondq/Downloads/2022-04-07_17-11.png"
	UploadFile(fileUpload, url)
}

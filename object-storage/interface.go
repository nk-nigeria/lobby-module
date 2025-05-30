package objectstorage

import "time"

type ObjStorage interface {
	MakeBucket(bucketName string) error
	PresignGetObject(bucketName string, objectName string, expiry time.Duration, params map[string]interface{}) (string, error)
	PresigPutObject(bucketName string, objectName string, expiry time.Duration, params map[string]interface{}) (string, error)
}

type EmptyStorage struct{}

func (e *EmptyStorage) MakeBucket(bucketName string) error {
	return nil
}
func (e *EmptyStorage) PresignGetObject(bucketName string, objectName string, expiry time.Duration, params map[string]interface{}) (string, error) {
	return "", nil
}

func (e *EmptyStorage) PresigPutObject(bucketName string, objectName string, expiry time.Duration, params map[string]interface{}) (string, error) {
	return "", nil
}

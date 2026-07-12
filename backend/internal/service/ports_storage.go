package service

import "context"

// StorageClient is the object-write port. Real implementations wrap object
// storage (see internal/adapter and internal/worker).
type StorageClient interface {
	Put(ctx context.Context, key string, body []byte, contentType string) (url string, err error)
}

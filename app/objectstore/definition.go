package objectstore

import (
	"context"
)

type PutObject struct {
	Data     []byte
	MetaData map[string]string
}

type Object struct {
	PutObject
	Key string
}

type Store interface {
	PutObject(ctx context.Context, key string, obj PutObject) (Object, error)
	GetDownloadURL(ctx context.Context, key string) (string, error)
	GetObject(ctx context.Context, key string) (Object, error)
	DeleteObject(ctx context.Context, key string) error
}

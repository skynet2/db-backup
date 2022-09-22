package storage

import (
	"context"
	"io"
	"time"
)

type Provider interface {
	Validate(ctx context.Context) error
	List(ctx context.Context, prefix string) ([]File, error)
	Remove(ctx context.Context, absolutePath string) error
	Upload(ctx context.Context, finalFilePath string, reader io.Reader) error
	GetType() string
}

type File struct {
	AbsolutePath string
	CreatedAt    time.Time
}

package storage

import (
	"context"
	"os"
	"time"
)

type Provider interface {
	Validate(ctx context.Context) error
	List(ctx context.Context, prefix string) ([]File, error)
	Remove(ctx context.Context, absolutePath string) error
	Upload(ctx context.Context, finalFilePath string, reader *os.File) error
	GetType() string
}

type File struct {
	AbsolutePath string
	CreatedAt    time.Time
}

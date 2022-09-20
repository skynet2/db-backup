package storage

import (
	"context"
	"time"
)

type Provider interface {
	Validate(ctx context.Context) error
	ListFiles(ctx context.Context, prefix string) ([]File, error)
}

type File struct {
	AbsolutePath string
	CreatedAt    time.Time
}

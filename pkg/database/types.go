package database

import "context"

type Provider interface {
	Validate(ctx context.Context) error
	ListDatabase(ctx context.Context) ([]string, error)
	BackupDatabase(ctx context.Context, databaseName string) (string, string, error)
	GetType() string
	GetParametersList() []Parameter
}
type Parameter struct {
	Name        string
	Description string
}

type Service interface {
}

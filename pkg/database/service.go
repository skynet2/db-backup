package database

import (
	"context"
	"github.com/skynet2/db-backup/pkg/configuration"
)

type DefaultService struct {
	cfg configuration.Configuration
}

func (d DefaultService) ProcessBackup(ctx context.Context) {
	
}

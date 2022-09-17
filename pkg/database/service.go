package database

import (
	"context"
	"github.com/pkg/errors"
	"github.com/skynet2/db-backup/pkg/configuration"
	"golang.org/x/exp/slices"
)

type DefaultService struct {
	cfg      configuration.Configuration
	provider Provider
}

func NewDefaultService(
	cfg configuration.Configuration,
	provider Provider,
) Service {
	return &DefaultService{
		cfg:      cfg,
		provider: provider,
	}
}

func (d DefaultService) Validate(ctx context.Context) error {
	return d.provider.Validate(ctx)
}

func (d DefaultService) ProcessBackup(ctx context.Context) error {
	dbList, err := d.provider.ListDatabase(ctx)

	if err != nil {
		return errors.WithStack(err)
	}

	toBackup := d.getBackupTargets(dbList)

	if len(toBackup) == 0 {
		return errors.New("no target databases found")
	}
}

func (d DefaultService) getBackupTargets(existingDbs []string) []string {
	var toBackup []string

	if len(d.cfg.IncludeDbs) > 0 {
		for _, dbName := range d.cfg.IncludeDbs {
			if !slices.Contains(existingDbs, dbName) {
				// todo log

				continue
			}

			toBackup = append(toBackup, dbName)
		}

		return toBackup
	}

	for _, existing := range existingDbs {
		if slices.Contains(d.cfg.ExcludeDbs, existing) {
			continue
		}

		toBackup = append(toBackup, existing)
	}

	return toBackup
}

package main

import (
	"context"
	"github.com/pkg/errors"
	"github.com/skynet2/db-backup/pkg/configuration"
	"github.com/skynet2/db-backup/pkg/database"
	"github.com/skynet2/db-backup/pkg/storage"
	"golang.org/x/exp/slices"
)

type Service struct {
	dbProvider      database.Provider
	storageProvider storage.Provider
	cfg             configuration.Configuration
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Process(ctx context.Context) error {
	if err := s.validate(ctx); err != nil {
		return err
	}

	dbs, err := s.dbProvider.ListDatabase(ctx)

	if err != nil {
		return err
	}

	dbs = s.getDbsToBackup(dbs)

	if len(dbs) == 0 {
		return errors.New("no databases to backup")
	}

	for _, db := range dbs {
		func() {
			innerCtx, cancel := context.WithCancel(ctx)

			defer func() {
				cancel()
			}()

			s.dbProvider.BackupDatabase(innerCtx, db)
		}()
	}
}

func (s *Service) getDbsToBackup(existingDbs []string) []string {
	var toBackup []string

	if len(s.cfg.IncludeDbs) > 0 {
		for _, dbName := range s.cfg.IncludeDbs {
			if !slices.Contains(existingDbs, dbName) {
				continue
			}

			toBackup = append(toBackup, dbName)
		}

		return toBackup
	}

	for _, existing := range existingDbs {
		if slices.Contains(s.cfg.ExcludeDbs, existing) {
			continue
		}

		toBackup = append(toBackup, existing)
	}

	return toBackup
}

func (s *Service) validate(ctx context.Context) error {
	if err := s.dbProvider.Validate(ctx); err != nil {
		return err
	}

	if err := s.storageProvider.Validate(ctx); err != nil {
		return err
	}

	return nil
}

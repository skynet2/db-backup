package main

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/skynet2/db-backup/pkg/common"
	"github.com/skynet2/db-backup/pkg/configuration"
	"github.com/skynet2/db-backup/pkg/database"
	"github.com/skynet2/db-backup/pkg/storage"
	"golang.org/x/exp/slices"
	"os"
	"path/filepath"
	"time"
)

type Service struct {
	dbProvider      database.Provider
	storageProvider storage.Provider
	cfg             configuration.Configuration
}

func NewService(
	dbProvider database.Provider,
	storageProvider storage.Provider,
	cfg configuration.Configuration,
) *Service {
	return &Service{
		dbProvider:      dbProvider,
		storageProvider: storageProvider,
		cfg:             cfg,
	}
}

func (s *Service) Process(ctx context.Context) ([]common.Job, error) {
	if err := s.validate(ctx); err != nil {
		return nil, err
	}

	dbs, err := s.dbProvider.ListDatabase(ctx)

	if err != nil {
		return nil, err
	}

	dbs = s.getDbsToBackup(dbs)

	if len(dbs) == 0 {
		return nil, errors.New("no databases to backup")
	}

	var finalErrors error
	var jobs []common.Job

	for _, db := range dbs {
		func() {
			job := common.Job{
				DatabaseName: db,
				StartedAt:    time.Now().UTC(),
				Error:        nil,
				FileLocation: "",
			}

			innerLogger := zerolog.Ctx(ctx).With().Str("db_name", db).Logger()
			innerCtx, cancel := context.WithCancel(ctx)
			innerCtx = innerLogger.WithContext(innerCtx)

			defer func() {
				cancel()

				job.EndAt = time.Now().UTC()

				if job.Error != nil {
					zerolog.Ctx(innerCtx).Err(err).Send()
				}

				jobs = append(jobs, job)
			}()

			filePrefixName, fileName, absolutePath := s.getFinalFilename(db)

			zerolog.Ctx(innerCtx).Debug().Msgf("prefix: %v\nfileName: %v\nabsolutePath: %v",
				filePrefixName, fileName, absolutePath)

			job.FileLocation = absolutePath

			zerolog.Ctx(innerCtx).Info().Msgf("backup for database [%v] => [%v]",
				db, job.FileLocation)

			job.DatabaseBackupStartedAt = time.Now().UTC()

			if output, err := s.dbProvider.BackupDatabase(innerCtx, db, job.FileLocation); err != nil {
				finalErrors = multierror.Append(finalErrors, err)
				job.Error = err
				job.Output = output

				return // stop job
			} else {
				job.Output = output
				job.DatabaseBackupEndedAt = time.Now().UTC()
			}

			zerolog.Ctx(innerCtx).Info().Msgf("backup for database [%v] finished in %v", db,
				job.DatabaseBackupEndedAt.Sub(job.DatabaseBackupStartedAt))

			file, err := os.Open(job.FileLocation)

			if err != nil {
				job.Error = errors.WithStack(err)
				return
			}

			defer func() {
				if closeErr := file.Close(); closeErr != nil {
					job.Error = multierror.Append(job.Error, errors.WithStack(closeErr))
				}

				zerolog.Ctx(innerCtx).Info().Msgf("removing local file copy at %v", job.FileLocation)

				if delErr := os.Remove(job.FileLocation); delErr != nil {
					wrapped := errors.Wrap(delErr, "can not remove local file")
					job.Error = multierror.Append(job.Error, errors.WithStack(wrapped))
				}
			}()

			n := time.Now().UTC()
			job.StorageProviderType = s.storageProvider.GetType()
			job.StorageProviderStartedAt = &n
			job.StorageFileLocation = fmt.Sprintf("%v/%v", s.cfg.Storage.S3, fileName)

			if err = s.storageProvider.Upload(ctx, job.StorageFileLocation, file); err != nil {
				job.Error = errors.WithStack(err)

				return
			}

			n = time.Now().UTC()
			job.UploadEndedAt = &n

			files, err := s.storageProvider.List(ctx, filePrefixName)

			if err != nil {
				job.Error = errors.WithStack(err)

				return
			}

			filesForRemoving := s.getFilesForRemoving(innerCtx, files)

			for _, toRemove := range filesForRemoving {
				job.RemovedFiles = append(job.RemovedFiles, toRemove.AbsolutePath)
				zerolog.Ctx(innerCtx).Info().Msgf("removing deprecated file from storage %v", toRemove.AbsolutePath)

				if err := s.storageProvider.Remove(innerCtx, toRemove.AbsolutePath); err != nil {
					job.Error = multierror.Append(job.Error, errors.WithStack(err))
				}
			}
		}()
	}

	return jobs, nil
}

func (s *Service) getFilesForRemoving(ctx context.Context, files []storage.File) []storage.File {
	filesToStore := s.cfg.Storage.MaxFiles

	if filesToStore == 0 {
		filesToStore = 5
	}

	if filesToStore >= len(files) {
		return nil
	}

	return files[:filesToStore]
}

func (s *Service) getFinalFilename(dbName string) (string, string, string) {
	prefix := fmt.Sprintf("db-%v-", dbName)
	fileName := fmt.Sprintf("%v%v.sql.gzip", prefix,
		time.Now().UTC().Format("2006_01_02-15_04_05"))

	fullPath := filepath.Join(s.cfg.Db.DumpDir, fileName)

	return prefix, fileName, fullPath
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

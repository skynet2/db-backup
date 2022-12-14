package main

import (
	"context"
	"fmt"
	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigyaml"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/skynet2/db-backup/pkg/configuration"
	"github.com/skynet2/db-backup/pkg/database"
	"github.com/skynet2/db-backup/pkg/notifier"
	"github.com/skynet2/db-backup/pkg/storage"
	"strings"
)

func main() {
	setupZeroLog()
	cfg := configuration.Configuration{}

	if err := aconfig.LoaderFor(&cfg, aconfig.Config{
		Files:              []string{"./config.yaml", "./config.local.yaml"},
		MergeFiles:         true,
		AllowUnknownFields: true,
		FileDecoders: map[string]aconfig.FileDecoder{
			".yaml": aconfigyaml.New(),
			".yml":  aconfigyaml.New(),
		},
	}).Load(); err != nil {
		log.Fatal().Err(err).Send()
	}

	notifyService, err := notifier.NewDefaultService(cfg.Notifications)

	if err != nil {
		log.Fatal().Err(err).Send()
	}

	storageProvider, err := getStorageProvider(cfg.Storage)

	if err != nil {
		log.Fatal().Err(err).Send()
	}

	dbProvider, err := getDbProvider(cfg.Db)

	if err != nil {
		log.Fatal().Err(err).Send()
	}

	service := NewService(dbProvider, storageProvider, cfg)

	ctx := log.Logger.WithContext(context.Background())

	jobs, err := service.Process(ctx)

	if err != nil {
		if innerErr := notifyService.SendError(ctx, err); innerErr != nil {
			log.Err(err).Send()
		}

		log.Fatal().Err(err).Send()
	}

	if err = notifyService.SendResults(ctx, jobs); err != nil {
		if innerErr := notifyService.SendError(ctx, err); innerErr != nil {
			log.Err(err).Send()
		}

		log.Err(err).Send()
	}
}

func setupZeroLog() {
	zerolog.CallerMarshalFunc = func(pc uintptr, file string, line int) string {
		sp := strings.Split(file, "/")

		segments := 4

		if len(sp) == 0 { // just in case
			segments = 0
		}

		if segments > 0 && segments > len(sp) {
			segments = len(sp) - 1
		}

		return fmt.Sprintf("%s:%v", strings.Join(sp[segments:], "/"), line)
	}

	log.Logger = log.Logger.With().Caller().Logger()
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.DefaultContextLogger = &log.Logger
}

func getDbProvider(cfg configuration.DbConfiguration) (database.Provider, error) {
	provider := strings.TrimSpace(strings.ToLower(cfg.Provider))

	switch provider {
	case "postgres":
		return database.NewPostgresProvider(cfg.Postgres), nil
	default:
		return nil, errors.New(fmt.Sprintf("no implementation for database provider %v", provider))
	}
}

func getStorageProvider(cfg configuration.StorageConfiguration) (storage.Provider, error) {
	provider := strings.TrimSpace(strings.ToLower(cfg.Provider))

	switch provider {
	case "s3":
		return storage.NewS3Provider(cfg.S3), nil
	default:
		return nil, errors.New(fmt.Sprintf("no implementation for storage provider %v", provider))
	}
}

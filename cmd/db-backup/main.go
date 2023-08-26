package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigdotenv"
	"github.com/cristalhq/aconfig/aconfigyaml"
	"github.com/davecgh/go-spew/spew"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"

	"github.com/skynet2/db-backup/pkg/configuration"
	"github.com/skynet2/db-backup/pkg/database"
	"github.com/skynet2/db-backup/pkg/notifier"
	"github.com/skynet2/db-backup/pkg/storage"
)

func main() {
	setupZeroLog()
	registerMetrics()

	cfg := configuration.Configuration{}

	configFiles := []string{
		"./config.yaml",
		"./config.local.yaml",
	}

	if v := os.Getenv("ADDITIONAL_CONFIGS"); v != "" {
		configFiles = append(configFiles, strings.Split(v, ",")...)
	}

	log.Info().Msgf("using additional configuration files: %v", spew.Sdump(configFiles))

	if err := aconfig.LoaderFor(&cfg, aconfig.Config{
		Files:              configFiles,
		MergeFiles:         true,
		AllowUnknownFields: true,
		FileDecoders: map[string]aconfig.FileDecoder{
			".yaml": aconfigyaml.New(),
			".yml":  aconfigyaml.New(),
			".env":  aconfigdotenv.New(),
		},
	}).Load(); err != nil {
		log.Fatal().Err(err).Send()
	}

	defer func() {
		if pushErr := pushMetrics(cfg.Metrics.PrometheusPushGatewayUrl, cfg.Metrics.PrometheusJobName); pushErr != nil {
			log.Err(pushErr).Send()
		}
	}()

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

package main

import (
	"context"
	"fmt"
	"github.com/cristalhq/aconfig"
	"github.com/cristalhq/aconfig/aconfigyaml"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/skynet2/db-backup/pkg/configuration"
	"github.com/skynet2/db-backup/pkg/database"
	"github.com/skynet2/db-backup/pkg/storage"
	"strings"
)

func main() {
	cfg := configuration.Configuration{}

	if err := aconfig.LoaderFor(&cfg, aconfig.Config{
		Files:      []string{"./config.yaml", "./config.local.yaml"},
		MergeFiles: true,
		FileDecoders: map[string]aconfig.FileDecoder{
			".yaml": aconfigyaml.New(),
			".yml":  aconfigyaml.New(),
		},
	}).Load(); err != nil {
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
		log.Fatal().Err(err).Send()
	}

	fmt.Println(jobs)
	//ss := NewService(nil, nil, nil)
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

package database

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"os/exec"
	"path/filepath"
	"time"
)

type PostgresProvider struct {
}

func NewPostgresProvider() Provider {
	return &PostgresProvider{}
}

func (p PostgresProvider) GetParametersList() []Parameter {
	return []Parameter{
		{
			Name:        "DB_HOST",
			Description: "Database host. ex: localhost",
		},
		{
			Name:        "DB_PORT",
			Description: "Database port. ex: 5432. Default 5432",
		},
		{
			Name:        "DB_DB_DEFAULT_NAME",
			Description: "Database default name. ex: postgres. Default postgres",
		},
		{
			Name:        "DB_USER",
			Description: "Database user",
		},
		{
			Name:        "DB_PASSWORD",
			Description: "Database password",
		},
		{
			Name:        "DB_DUMP_DIR",
			Description: "Temporary database dump dir",
		},
		{
			Name:        "DB_TLS_ENABLED",
			Description: "Database enable TLS support",
		},
		{
			Name:        "DB_PGDUMP_CUSTOM_ARGS",
			Description: "Database pgdump custom args",
		},
		{
			Name:        "DB_COMPRESSION_LEVEL",
			Description: "Database compression level. Default 5",
		},
	}
}

func (p PostgresProvider) Validate(ctx context.Context) error {
	// todo validate pgdump location
	con, err := p.getConnection(ctx)

	if err != nil {
		return errors.WithStack(err)
	}

	_ = con.Close(ctx)
	return nil
}

func (p PostgresProvider) ListDatabase(ctx context.Context) ([]string, error) {
	con, err := p.getConnection(ctx)

	if err != nil {
		return nil, err
	}

	rows, err := con.Query(ctx, "select datname from pg_database where datistemplate = false")

	if err != nil {
		return nil, errors.WithStack(err)
	}

	var dbs []string
	var dbName string

	for rows.Next() {
		if err = rows.Scan(&dbName); err != nil {
			return nil, errors.WithStack(err)
		}

		dbs = append(dbs, dbName)
	}

	return dbs, nil
}

func (p PostgresProvider) getCompressionLevel() int {
	level := viper.GetInt("DB_COMPRESSION_LEVEL")

	if level == 0 {
		return 5
	}

	return level
}

func (p PostgresProvider) BackupDatabase(ctx context.Context, databaseName string) (string, string, error) {
	targetFileName := p.getTargetFileName(databaseName)

	cmd := exec.Command("pg_dump",
		"-F p",
		fmt.Sprintf("-U %v", viper.GetString("DB_USER")),
		fmt.Sprintf("-h %v", viper.GetString("DB_HOST")),
		fmt.Sprintf("-f %v", targetFileName),
		fmt.Sprintf("-Z %v", p.getCompressionLevel()),
		fmt.Sprintf("-d %v", databaseName),
	)

	dbPassword := viper.GetString("DB_PASSWORD")

	if len(dbPassword) > 0 {
		cmd.Env = append(cmd.Env, fmt.Sprintf("PGPASSWORD=%v", dbPassword))
	}

	output, err := cmd.CombinedOutput()

	if err != nil {
		return "", string(output), errors.WithStack(err)
	}

	return targetFileName, string(output), nil
}

func (p PostgresProvider) getTargetFileName(dbName string) string {
	return filepath.Join(viper.GetString("DB_DUMP_DIR"), fmt.Sprintf("%v.sql.gzip", dbName))
}

func (p PostgresProvider) GetType() string {
	return "postgres"
}

func (p PostgresProvider) getConnection(ctx context.Context) (*pgx.Conn, error) {
	dbPort := uint16(viper.GetInt32("DB_PORT"))

	if dbPort == 0 {
		dbPort = 5432
	}

	defaultDbName := viper.GetString("DB_DB_DEFAULT_NAME")

	if len(defaultDbName) == 0 {
		defaultDbName = "postgres"
	}

	var tlsConfig *tls.Config

	if viper.GetBool("DB_TLS_ENABLED") {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	con, err := pgx.ConnectConfig(ctx, &pgx.ConnConfig{
		Config: pgconn.Config{
			Host:           viper.GetString("DB_HOST"),
			Port:           dbPort,
			Database:       defaultDbName,
			User:           viper.GetString("DB_USER"),
			Password:       viper.GetString("DB_PASSWORD"),
			TLSConfig:      tlsConfig,
			ConnectTimeout: 15 * time.Second,
		},
	})

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return con, nil
}

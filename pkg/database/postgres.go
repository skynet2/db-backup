package database

import (
	"context"
	"crypto/tls"
	"fmt"
	"os/exec"

	"github.com/cockroachdb/errors"
	"github.com/jackc/pgx/v4"

	"github.com/skynet2/db-backup/pkg/configuration"
)

type PostgresProvider struct {
	cfg configuration.PostgresConfiguration
}

func NewPostgresProvider(cfg configuration.PostgresConfiguration) Provider {
	return &PostgresProvider{
		cfg: cfg,
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
	level := p.cfg.CompressionLevel

	if level == 0 {
		return 5
	}

	return level
}

func (p PostgresProvider) BackupDatabase(
	ctx context.Context,
	databaseName string,
	finalFileName string,
) (string, error) {
	cmd := exec.Command("pg_dump",
		fmt.Sprintf("--username=%v", p.cfg.User),
		fmt.Sprintf("--host=%v", p.cfg.Host),
		fmt.Sprintf("--file=%v", finalFileName),
		fmt.Sprintf("--compress=%v", p.getCompressionLevel()),
		fmt.Sprintf("--dbname=%v", databaseName),
	)

	dbPassword := p.cfg.Password

	if len(dbPassword) > 0 {
		cmd.Env = append(cmd.Env, fmt.Sprintf("PGPASSWORD=%v", dbPassword))
	}

	output, err := cmd.CombinedOutput()

	if err != nil {
		return string(output), errors.Wrap(err, string(output))
	}

	return string(output), nil
}

func (p PostgresProvider) GetType() string {
	return "postgres"
}

func (p PostgresProvider) getConnection(ctx context.Context) (*pgx.Conn, error) {
	defaultDbName := p.cfg.DbDefaultName

	if len(defaultDbName) == 0 {
		defaultDbName = "postgres"
	}

	var tlsConfig *tls.Config

	if p.cfg.TlsEnabled {
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	conStr, err := pgx.ParseConfig(fmt.Sprintf("postgresql://%v:%v@%v:%v/%v?connect_timeout=10&application_name=backup",
		p.cfg.User, p.cfg.Password, p.cfg.Host, p.cfg.Port, defaultDbName))

	if err != nil {
		return nil, errors.WithStack(err)
	}

	conStr.TLSConfig = tlsConfig

	con, err := pgx.ConnectConfig(ctx, conStr)

	if err != nil {
		return nil, errors.WithStack(err)
	}

	return con, nil
}

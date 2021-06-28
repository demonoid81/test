package app

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/sphera-erp/sphera/pkg/migrate"
	"github.com/sphera-erp/sphera/pkg/pglx"
)

// InitCockroachDB - Инициализация Базы данных
func (app *App) InitCockroachDB(ctx context.Context) (*pglx.DB, error) {
	var dbURL string
	// получим конфиги
	cfg := app.Cfg.DB
	dbURL = cockroachURL(cfg)
	// распарсим
	connConfig, err := pgxpool.ParseConfig(dbURL)
	if err != nil {
		app.Logger.Fatal().Str("module", "cockroach").Str("func", "initCockroachDB").Err(err).Msg("Parse config connection error")
		return nil, err
	}
	// подключим логгер
	connConfig.ConnConfig.Logger = app
	connConfig.MaxConns = 20
	connConfig.LazyConnect = true
	// присоединимся к БД
	db, err := pglx.Connect(ctx, connConfig)
	if err != nil {
		app.Logger.Fatal().Str("module", "cockroach").Str("func", "initCockroachDB").Err(err).Msg("CockroachDB connection error")
		return nil, err
	}

	app.Logger.Info().Str("module", "cockroach").Str("func", "initCockroachDB").Err(err).Msg("CockroachDB is connected")
	return db, nil
}

func cockroachURL(cfg CockroachCfg) string {
	fmt.Println(cfg)
	url := "postgres://"
	if cfg.Username != "" {
		url += cfg.Username
		if cfg.Password != "" {
			url += ":" + cfg.Password
		}
		url += "@"
	}
	if cfg.DirectoryPath != "" {
		url += "?host=" + cfg.DirectoryPath
		if cfg.Port != 0 {
			url += "&port=" + fmt.Sprint(cfg.Port)
		}
		if cfg.Database != "" {
			url += "&database=" + cfg.Database
		}
		if cfg.ApplicationName != "" || cfg.SslMode != "" {
			url += "&"
		}
	} else {
		url += cfg.Host
		if cfg.Port != 0 {
			url += ":" + fmt.Sprint(cfg.Port)
		}
		if cfg.Database != "" {
			url += "/" + cfg.Database
		}
		if cfg.ApplicationName != "" || cfg.SslMode != "" {
			url += "?"
		}
	}
	if cfg.ApplicationName != "" {
		url += "application_name=" + cfg.ApplicationName
		if cfg.SslMode != "" {
			url += "&"
		}
	}
	if cfg.SslMode != "" {
		url += "sslmode=" + cfg.SslMode
		if cfg.SslMode != "disable" {
			url += "&sslrootcert" + cfg.SslRootCert + "&sslcert" + cfg.SslCert + "&sslkey" + cfg.SslKey
		}
	}
	return url
}

func (app *App) MigrateDatabase(ctx context.Context, conn *pglx.DB) error {

	migrator, err := migrate.NewMigrator(ctx, conn, "schema_version", nil)
	if err != nil {
		app.Logger.Fatal().Str("module", "cockroach").Str("func", "migrateDatabase").Err(err).Msg("Unable to create a migrator")
		return err
	}

	err = migrator.LoadMigrations(app.Cfg.MigrationsPath)
	if err != nil {
		app.Logger.Info().Msg(fmt.Sprintf("Unable to migrate: %v", err))
		//return err
	}

	err = migrator.Migrate(ctx)
	if err != nil {
		app.Logger.Fatal().Err(err).Msg("Unable to migrate")
		return err
	}

	ver, err := migrator.GetCurrentVersion(ctx)
	if err != nil {
		app.Logger.Error().Err(err).Msg("Unable to get current schema version")
		return err
	}

	app.Logger.Info().Msgf("Migration done. Current schema version: %v\n", ver)
	return nil
}

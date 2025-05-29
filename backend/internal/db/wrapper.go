package db

import (
	"context"
	"database/sql"
	"log/slog"
	"pluralkit/status/internal/util"

	_ "github.com/mattn/go-sqlite3"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/extra/bundebug"
)

type DB struct {
	logger   *slog.Logger
	database *bun.DB
}

func NewDB(config util.Config, logger *slog.Logger) *DB {

	moduleLogger := logger.With(slog.String("module", "db"))

	sqldb, err := sql.Open("sqlite3", config.DBLoc)
	if err != nil {
		moduleLogger.Error("error while opening database", slog.Any("error", err))
		return nil
	}
	defer sqldb.Close()

	bunDB := bun.NewDB(sqldb, sqlitedialect.New())
	bunDB.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true))) //tmp - dev/debugging

	db := &DB{
		logger:   moduleLogger,
		database: bunDB,
	}

	err = db.initDB()
	if err != nil {
		return nil
	}

	return db
}

func (d *DB) initDB() error {
	ctx := context.Background()

	_, err := d.database.NewCreateTable().Model((*util.Incident)(nil)).IfNotExists().Exec(ctx)
	if err != nil {
		d.logger.Error("error while creating incidents table", slog.Any("error", err))
		return err
	}

	_, err = d.database.NewCreateTable().Model((*util.IncidentUpdate)(nil)).IfNotExists().Exec(ctx)
	if err != nil {
		d.logger.Error("error while creating incidents updates table", slog.Any("error", err))
		return err
	}

	return nil
}

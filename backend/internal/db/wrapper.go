package db

import (
	"context"
	"database/sql"
	"log/slog"
	"pluralkit/status/internal/util"
	"time"

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
func (d *DB) CloseDB() {
	d.database.Close()
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

	_, err = d.database.NewCreateTable().Model((*util.StatusWrapper)(nil)).IfNotExists().Exec(ctx)
	if err != nil {
		d.logger.Error("error while creating status table", slog.Any("error", err))
		return err
	}

	return nil
}

func (d *DB) LoadStatus(ctx context.Context) (util.Status, error) {
	statusWrapper := &util.StatusWrapper{
		ID: 1,
		Status: util.Status{
			Status:          util.StatusOperational,
			Impact:          util.ImpactNone,
			ActiveIncidents: make([]string, 0),
		},
	}
	err := d.database.NewSelect().Model(statusWrapper).Limit(1).Scan(ctx)
	return statusWrapper.Status, err
}

func (d *DB) SaveStatus(ctx context.Context, status util.Status) error {
	statusWrapper := &util.StatusWrapper{
		ID:     1,
		Status: status,
	}
	_, err := d.database.NewInsert().On("CONFLICT (id) DO UPDATE").Model(statusWrapper).Column("status").Exec(ctx)
	return err
}

func (d *DB) GetIncidents(ctx context.Context, ids []string) (util.IncidentList, error) {
	list := util.IncidentList{
		Timestamp: time.Now(),
		Incidents: make(map[string]util.Incident),
	}

	if len(ids) == 0 {
		return list, nil
	}

	incidents := make([]util.Incident, len(ids))
	err := d.database.NewSelect().Model((*util.Incident)(nil)).Where("id IN (?)", bun.In(ids)).Scan(ctx, &incidents)
	if err != nil {
		return list, err
	}

	for _, incident := range incidents {
		list.Incidents[incident.ID] = incident
	}
	return list, nil
}

func (d *DB) CreateIncident(ctx context.Context, incident util.Incident) error {
	return nil
}

func (d *DB) EditIncident(ctx context.Context, incident util.Incident) error {
	return nil
}

func (d *DB) DeleteIncident(ctx context.Context, id string) error {
	return nil
}

func (d *DB) CreateUpdate(ctx context.Context, update util.IncidentUpdate) error {
	return nil
}

func (d *DB) EditUpdate(ctx context.Context, update util.IncidentUpdate) error {
	return nil
}

func (d *DB) DeleteUpdate(ctx context.Context, id string) error {
	return nil
}

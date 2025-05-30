package db

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"pluralkit/status/internal/util"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/sqids/sqids-go"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/sqlitedialect"
	"github.com/uptrace/bun/extra/bundebug"
)

type DB struct {
	logger   *slog.Logger
	database *bun.DB
	sq       *sqids.Sqids
}

func NewDB(config util.Config, logger *slog.Logger) *DB {
	moduleLogger := logger.With(slog.String("module", "db"))

	sqldb, err := sql.Open("sqlite3", config.DBLoc)
	if err != nil {
		moduleLogger.Error("error while opening database", slog.Any("error", err))
		return nil
	}

	bunDB := bun.NewDB(sqldb, sqlitedialect.New(), bun.WithDiscardUnknownColumns())
	if config.LogLevel == util.SlogLevel(slog.LevelDebug) {
		bunDB.AddQueryHook(bundebug.NewQueryHook(bundebug.WithVerbose(true)))
	}

	sq, err := sqids.New(sqids.Options{
		MinLength: 8,
	})
	if err != nil {
		moduleLogger.Error("error while intializing sqids", slog.Any("error", err))
		return nil
	}

	db := &DB{
		logger:   moduleLogger,
		database: bunDB,
		sq:       sq,
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
			OverallStatus:   util.StatusOperational,
			ActiveIncidents: make([]string, 0),
		},
	}
	err := d.database.NewSelect().
		Model(statusWrapper).
		Limit(1).
		Scan(ctx)
	return statusWrapper.Status, err
}

func (d *DB) SaveStatus(ctx context.Context, status util.Status) error {
	statusWrapper := &util.StatusWrapper{
		ID:     1,
		Status: status,
	}
	_, err := d.database.NewInsert().
		On("CONFLICT (id) DO UPDATE").
		Model(statusWrapper).
		Exec(ctx)
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
	err := d.database.NewSelect().
		Model(&incidents).
		Relation("Updates").
		Where("id IN (?)", bun.In(ids)).
		Scan(ctx)
	if err != nil {
		return list, err
	}

	for _, incident := range incidents {
		list.Incidents[incident.ID] = incident
	}
	return list, nil
}

func (d *DB) GetActiveIncidents(ctx context.Context) (util.IncidentList, error) {
	list := util.IncidentList{
		Timestamp: time.Now(),
		Incidents: make(map[string]util.Incident),
	}

	incidents := make([]util.Incident, 0)
	//TODO: do this in a faster way lmao
	err := d.database.NewSelect().
		Model(&incidents).
		Relation("Updates").
		Where("status != ?", util.StatusResolved).
		Scan(ctx)
	if err != nil {
		return list, err
	}

	for _, incident := range incidents {
		list.Incidents[incident.ID] = incident
	}
	return list, nil
}

func (d *DB) CreateIncident(ctx context.Context, incident util.Incident) (string, error) {
	count, err := d.database.NewSelect().Model((*util.Incident)(nil)).Count(ctx)
	if err != nil {
		return "", err
	}

	id, err := d.sq.Encode([]uint64{uint64(count)})
	if err != nil {
		return "", err
	}

	incident.ID = id

	_, err = d.database.NewInsert().
		Model(&incident).
		Exec(ctx)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (d *DB) EditIncident(ctx context.Context, id string, patch util.IncidentPatch) error {
	// we use a map to patch because... that seems to be the easiest way?
	// TODO: validation of status and impact fields
	patchMap := make(map[string]interface{})

	if patch.Name != nil {
		patchMap["name"] = *patch.Name
	}
	if patch.Description != nil {
		patchMap["description"] = *patch.Description
	}
	if patch.Status != nil {
		patchMap["status"] = *patch.Status
	}
	if patch.Impact != nil {
		patchMap["impact"] = *patch.Impact
	}

	if len(patchMap) == 0 {
		return nil // prevent update if there isn't anything to update
	}
	patchMap["last_update"] = time.Now()

	_, err := d.database.NewUpdate().
		Model(&patchMap).
		Table("incidents").
		Where("id = ?", id).
		Exec(ctx)
	return err
}

func (d *DB) DeleteIncident(ctx context.Context, incident util.Incident) error {
	_, err := d.database.NewDelete().
		Model(&incident).
		WherePK().
		Exec(ctx)
	return err
}

func (d *DB) CreateUpdate(ctx context.Context, update util.IncidentUpdate) (string, error) {
	if len(update.IncidentID) == 0 {
		return "", errors.New("incidentID not provided")
	}

	count, err := d.database.NewSelect().Model((*util.IncidentUpdate)(nil)).Count(ctx)
	if err != nil {
		return "", err
	}

	iid := d.sq.Decode(update.IncidentID)
	id, err := d.sq.Encode(append(iid, uint64(count)))
	if err != nil {
		return "", err
	}

	update.ID = id

	_, err = d.database.NewInsert().
		Model(&update).
		Exec(ctx)
	if err != nil {
		return "", err
	}

	_, err = d.database.NewUpdate().
		Model(&util.Incident{}).
		Set("last_update = ?", time.Now()).
		Where("id = ?", update.IncidentID).
		Exec(ctx)
	if err != nil {
		return "", err
	}

	return id, err
}

func (d *DB) GetUpdate(ctx context.Context, id string) (util.IncidentUpdate, error) {
	update := util.IncidentUpdate{ID: id}

	if len(id) == 0 {
		return update, nil
	}

	err := d.database.NewSelect().
		Model(&update).
		WherePK().
		Limit(1).
		Scan(ctx)
	if err != nil {
		return update, err
	}

	return update, nil
}

func (d *DB) EditUpdate(ctx context.Context, id string, update util.UpdatePatch) error {
	patchMap := make(map[string]interface{})

	if update.Text != nil {
		patchMap["text"] = *update.Text
	}

	if len(patchMap) == 0 {
		return nil // prevent update if there isn't anything to update
	}

	_, err := d.database.NewUpdate().
		Model(&patchMap).
		Table("incident_updates").
		Where("id = ?", id).
		Exec(ctx)
	return err
}

func (d *DB) DeleteUpdate(ctx context.Context, update util.IncidentUpdate) error {
	_, err := d.database.NewDelete().
		Model(&update).
		WherePK().
		Exec(ctx)
	return err
}

package db

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"pluralkit/status/util"
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
	events   chan util.Event
	sq       *sqids.Sqids
}

func NewDB(config util.Config, logger *slog.Logger, eventChannel chan util.Event) *DB {
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
		events:   eventChannel,
		sq:       sq,
	}

	err = db.initDB()
	if err != nil {
		return nil
	}

	return db
}
func (d *DB) CloseDB() error {
	return d.database.Close()
}

func (d *DB) initDB() error {
	ctx := context.Background()

	_, err := d.database.NewCreateTable().
		Model((*util.Incident)(nil)).
		IfNotExists().
		Exec(ctx)
	if err != nil {
		d.logger.Error("error while creating incidents table", slog.Any("error", err))
		return err
	}
	_, err = d.database.NewCreateIndex().
		Model((*util.Incident)(nil)).
		IfNotExists().
		Index("idx_status").
		Column("status").
		Exec(ctx)
	if err != nil {
		d.logger.Error("error while creating incidents status index", slog.Any("error", err))
		return err
	}

	_, err = d.database.NewCreateTable().
		Model((*util.IncidentUpdate)(nil)).
		IfNotExists().
		ForeignKey(`("incident_id") REFERENCES "incidents" ("id") ON DELETE CASCADE`).
		Exec(ctx)
	if err != nil {
		d.logger.Error("error while creating incidents updates table", slog.Any("error", err))
		return err
	}

	_, err = d.database.NewCreateTable().
		Model((*util.StatusWrapper)(nil)).
		IfNotExists().
		Exec(ctx)
	if err != nil {
		d.logger.Error("error while creating status table", slog.Any("error", err))
		return err
	}

	_, err = d.database.NewCreateTable().
		Model((*util.WebhookMessage)(nil)).
		IfNotExists().
		Exec(ctx)
	if err != nil {
		d.logger.Error("error while creating webhook messages table", slog.Any("error", err))
		return err
	}

	return nil
}

func (d *DB) GetStatus(ctx context.Context) (util.Status, error) {
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

func (d *DB) GetMessageID(ctx context.Context, id string, msgType string) (int64, error) {
	msg := util.WebhookMessage{}
	err := d.database.NewSelect().
		Model(&msg).
		Where("id = ?", id).
		Where("type = ?", msgType).
		Scan(ctx)
	if err != nil {
		return 0, err
	}
	return msg.MessageID, nil
}
func (d *DB) SaveMessageID(ctx context.Context, msgInfo util.WebhookMessage) error {
	_, err := d.database.NewInsert().
		Model(&msgInfo).
		Exec(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (d *DB) GetIncidents(ctx context.Context, ids []string) (util.IncidentList, error) {
	list := util.IncidentList{
		Timestamp: time.Now(),
		Incidents: make(map[string]util.Incident, len(ids)),
	}

	if len(ids) == 0 {
		return list, nil
	}

	for _, id := range ids {
		err := util.Validate.Var(id, "required,sqid")
		if err != nil {
			return list, util.ErrInvalid
		}
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

func (d *DB) GetIncident(ctx context.Context, id string) (util.Incident, error) {
	incident := util.Incident{}

	err := util.Validate.Var(id, "required,sqid")
	if err != nil {
		return incident, util.ErrInvalid
	}

	err = d.database.NewSelect().
		Model(&incident).
		Relation("Updates").
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return incident, util.ErrNotFound
		}
		return incident, err
	}
	return incident, nil
}

func (d *DB) GetIncidentsBefore(ctx context.Context, before time.Time) (util.IncidentList, error) {
	list := util.IncidentList{
		Timestamp: time.Now(),
		Incidents: make(map[string]util.Incident, 0),
	}

	incidents := make([]util.Incident, 0)
	err := d.database.NewSelect().
		Model(&incidents).
		Relation("Updates").
		Where("timestamp < ?", before).
		Limit(25).
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
	if !incident.Impact.IsValid() {
		return "", errors.New("invalid impact field")
	}
	if !incident.Status.IsValid() {
		return "", errors.New("invalid status field")
	}

	var maxRow sql.NullInt64
	var id int64
	err := d.database.NewRaw("SELECT MAX(rowid) FROM incidents").Scan(ctx, &maxRow)
	if err != nil {
		return "", err
	}
	if !maxRow.Valid {
		id = 0
	} else {
		id = maxRow.Int64
	}

	sqid, err := d.sq.Encode([]uint64{uint64(id)})
	if err != nil {
		return "", err
	}

	incident.ID = sqid

	err = util.Validate.Struct(incident)
	if err != nil {
		return "", util.ErrInvalid
	}

	_, err = d.database.NewInsert().
		Model(&incident).
		Returning("*").
		Exec(ctx)
	if err != nil {
		return "", err
	}

	d.events <- util.Event{
		Type:     util.EventCreateIncident,
		Modified: incident,
	}

	return incident.ID, nil
}

func (d *DB) EditIncident(ctx context.Context, id string, patch util.IncidentPatch) error {
	err := util.Validate.Struct(patch)
	if err != nil {
		return util.ErrInvalid
	}

	// we use a map to patch because... that seems to be the easiest way?
	patchMap := make(map[string]interface{})

	if patch.Name != nil {
		patchMap["name"] = *patch.Name
	}
	if patch.Description != nil {
		patchMap["description"] = *patch.Description
	}
	if patch.Status != nil {
		if !patch.Status.IsValid() {
			return errors.New("invalid status field")
		}
		patchMap["status"] = *patch.Status
		if *patch.Status == util.StatusResolved {
			patchMap["resolution_timestamp"] = time.Now()
		} else {
			patchMap["resolution_timestamp"] = time.Time{}
		}
	}
	if patch.Impact != nil {
		if !patch.Impact.IsValid() {
			return errors.New("invalid impact field")
		}
		patchMap["impact"] = *patch.Impact
	}

	if len(patchMap) == 0 {
		return nil // prevent update if there isn't anything to update
	}
	patchMap["last_update"] = time.Now()

	incident := util.Incident{}
	res, err := d.database.NewUpdate().
		Model(&patchMap).
		Table("incidents").
		Returning("*").
		Where("id = ?", id).
		Exec(ctx, &incident)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return util.ErrNotFound
		}
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	} else if rows == 0 {
		return util.ErrNotFound
	}

	d.events <- util.Event{
		Type:     util.EventEditIncident,
		Modified: incident,
	}

	return nil
}

func (d *DB) DeleteIncident(ctx context.Context, incident util.Incident) error {
	err := util.Validate.Var(incident.ID, "required,sqid")
	if err != nil {
		return util.ErrInvalid
	}

	res, err := d.database.NewDelete().
		Model(&incident).
		WherePK().
		Exec(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return util.ErrNotFound
		}
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	} else if rows == 0 {
		return util.ErrNotFound
	}

	return nil
}

func (d *DB) CreateUpdate(ctx context.Context, update util.IncidentUpdate) (string, error) {
	if len(update.IncidentID) == 0 {
		return "", errors.New("incidentID not provided")
	}

	exists, err := d.database.NewSelect().Model((*util.Incident)(nil)).Where("id = ?", update.IncidentID).Exists(ctx)
	if err != nil || !exists {
		return "", util.ErrNotFound
	}

	var maxRow sql.NullInt64
	var id int64
	err = d.database.NewRaw("SELECT MAX(rowid) FROM incident_updates").Scan(ctx, &maxRow)
	if err != nil {
		return "", err
	}
	if !maxRow.Valid {
		id = 0
	} else {
		id = maxRow.Int64
	}

	isqid := d.sq.Decode(update.IncidentID)
	sqid, err := d.sq.Encode([]uint64{isqid[0], uint64(id)})
	if err != nil {
		return "", err
	}
	update.ID = sqid

	err = util.Validate.Struct(update)
	if err != nil {
		return "", util.ErrInvalid
	}

	_, err = d.database.NewInsert().
		Model(&update).
		Returning("*").
		Exec(ctx)
	if err != nil {
		return "", err
	}

	if update.Status != nil && update.Status.IsValid() {
		resTime := time.Time{}
		if *update.Status == util.StatusResolved {
			resTime = time.Now()
		}
		_, err = d.database.NewUpdate().
			Model(&util.Incident{}).
			Set("last_update = ?", time.Now()).
			Set("status = ?", update.Status).
			Set("resolution_timestamp = ?", resTime).
			Where("id = ?", update.IncidentID).
			Exec(ctx)
	} else {
		_, err = d.database.NewUpdate().
			Model(&util.Incident{}).
			Set("last_update = ?", time.Now()).
			Where("id = ?", update.IncidentID).
			Exec(ctx)
	}

	if err != nil {
		return "", err
	}

	d.events <- util.Event{
		Type:     util.EventCreateUpdate,
		Modified: update,
	}

	return sqid, err
}

func (d *DB) GetUpdate(ctx context.Context, id string) (util.IncidentUpdate, error) {
	update := util.IncidentUpdate{ID: id}
	err := util.Validate.Var(id, "required,sqid")
	if err != nil {
		return update, util.ErrInvalid
	}

	err = d.database.NewSelect().
		Model(&update).
		WherePK().
		Limit(1).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return update, util.ErrNotFound
		}
		return update, err
	}

	return update, nil
}

func (d *DB) EditUpdate(ctx context.Context, id string, update util.UpdatePatch) error {
	err := util.Validate.Struct(update)
	if err != nil {
		return util.ErrInvalid
	}

	patchMap := make(map[string]interface{})

	if update.Text != nil {
		patchMap["text"] = *update.Text
	}
	if update.Status != nil {
		patchMap["status"] = *update.Status
	}

	if len(patchMap) == 0 {
		return nil // prevent update if there isn't anything to update
	}

	updated := util.IncidentUpdate{}
	res, err := d.database.NewUpdate().
		Model(&patchMap).
		Table("incident_updates").
		Returning("*").
		Where("id = ?", id).
		Exec(ctx, &updated)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return util.ErrNotFound
		}
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	} else if rows == 0 {
		return util.ErrNotFound
	}

	d.events <- util.Event{
		Type:     util.EventEditUpdate,
		Modified: updated,
	}
	return nil
}

func (d *DB) DeleteUpdate(ctx context.Context, update util.IncidentUpdate) error {
	err := util.Validate.Var(update.ID, "required,sqid")
	if err != nil {
		return util.ErrInvalid
	}

	res, err := d.database.NewDelete().
		Model(&update).
		WherePK().
		Exec(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return util.ErrNotFound
		}
		return err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return err
	} else if rows == 0 {
		return util.ErrNotFound
	}
	return nil
}

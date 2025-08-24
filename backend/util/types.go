package util

import (
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/uptrace/bun"
)

/* Validation Helpers =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=- */
const sqidAlphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var Validate = validator.New()

func init() {
	err := Validate.RegisterValidation("impact", validateImpact)
	if err != nil {
		slog.Error("error in init", slog.Any("error", err))
		os.Exit(1)
	}
	err = Validate.RegisterValidation("incidentstatus", validateIncidentStatus)
	if err != nil {
		slog.Error("error in init", slog.Any("error", err))
		os.Exit(1)
	}
	err = Validate.RegisterValidation("sqid", validateSqid)
	if err != nil {
		slog.Error("error in init", slog.Any("error", err))
		os.Exit(1)
	}
}

func validateSqid(fl validator.FieldLevel) bool {
	id := fl.Field().String()
	if len(id) < 8 {
		return false
	}
	for _, char := range id {
		if !strings.ContainsRune(sqidAlphabet, char) {
			return false
		}
	}
	return true
}

func validateImpact(fl validator.FieldLevel) bool {
	if impact, ok := fl.Field().Interface().(Impact); ok {
		return impact.IsValid()
	}
	if impact, ok := fl.Field().Interface().(*Impact); ok {
		if impact == nil {
			return true
		}
		return impact.IsValid()
	}
	return false
}

func validateIncidentStatus(fl validator.FieldLevel) bool {
	if status, ok := fl.Field().Interface().(IncidentStatus); ok {
		return status.IsValid()
	}
	if status, ok := fl.Field().Interface().(*IncidentStatus); ok {
		if status == nil {
			return true
		}
		return status.IsValid()
	}
	return false
}

/* Incidents + Status =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=- */

// a type representing the impact of an incident or event
type Impact string

const (
	ImpactNone  Impact = "none"
	ImpactMinor Impact = "minor"
	ImpactMajor Impact = "major"
)

var impactSeverity = map[Impact]int{
	ImpactNone:  0,
	ImpactMinor: 1,
	ImpactMajor: 2,
}

// returns true if x is higher level than y
func (x Impact) IsGreater(y Impact) bool {
	return impactSeverity[x] > impactSeverity[y]
}

// helper function for validating Impact
func (i Impact) IsValid() bool {
	switch i {
	case ImpactNone, ImpactMinor, ImpactMajor:
		return true
	default:
		return false
	}
}

// a type representing the status of an incident
type IncidentStatus string

const (
	StatusMaintenance   IncidentStatus = "maintenance"
	StatusInvestigating IncidentStatus = "investigating"
	StatusIdentified    IncidentStatus = "identified"
	StatusMonitoring    IncidentStatus = "monitoring"
	StatusResolved      IncidentStatus = "resolved"
)

// helper function for validating IncidentStatus
func (i IncidentStatus) IsValid() bool {
	switch i {
	case StatusMaintenance, StatusInvestigating, StatusIdentified, StatusMonitoring, StatusResolved:
		return true
	default:
		return false
	}
}

// a type representing the overall system status
type OverallStatus string

const (
	StatusOperational OverallStatus = "operational"
	StatusDegraded    OverallStatus = "degraded"
	StatusMajorOutage OverallStatus = "major_outage"
)

// struct representing a single update for an incident
type IncidentUpdate struct {
	bun.BaseModel `bun:"table:incident_updates,alias:upd"`

	ID        string          `json:"id" bun:"id,pk" validate:"required,sqid"`
	Text      string          `json:"text" bun:"text,notnull" validate:"required,max=1800"`
	Status    *IncidentStatus `json:"status,omitempty" bun:"status" validate:"omitempty,incidentstatus"`
	Timestamp time.Time       `json:"timestamp" bun:"timestamp,notnull,default:current_timestamp"`

	IncidentID string `json:"-" bun:"incident_id,notnull"`
}

// helper struct for update patching
type UpdatePatch struct {
	Text *string `json:"text" validate:"omitempty,required,max=1800"`
}

// render helper function for IncidentUpdate
func (i *IncidentUpdate) Render(w http.ResponseWriter, r *http.Request) error { return nil }

// struct representing a single incident, rougly based upon the atlassian statuspage format
type Incident struct {
	bun.BaseModel `bun:"table:incidents,alias:inc"`

	ID                  string         `json:"id" bun:"id,pk" validate:"required,sqid"`
	Timestamp           time.Time      `json:"timestamp" bun:"timestamp,nullzero,notnull,default:current_timestamp"`
	LastUpdate          time.Time      `json:"last_update" bun:"last_update,nullzero,notnull,default:current_timestamp"`
	ResolutionTimestamp time.Time      `json:"resolution_timestamp" bun:"resolution_timestamp,nullzero"`
	Status              IncidentStatus `json:"status" bun:"status" validate:"required,incidentstatus"`
	Impact              Impact         `json:"impact" bun:"impact" validate:"required,impact"`
	Name                string         `json:"name" bun:"name,notnull" validate:"required,max=100"`
	Description         string         `json:"description" bun:"description" validate:"max=1800"`

	Updates []*IncidentUpdate `json:"updates" bun:"rel:has-many,join:id=incident_id"  validate:"dive"`
}

// helper struct for incident patching
type IncidentPatch struct {
	Name        *string         `json:"name" validate:"max=100"`
	Description *string         `json:"description" validate:"max=1800"`
	Status      *IncidentStatus `json:"status" validate:"incidentstatus"`
	Impact      *Impact         `json:"impact" validate:"impact"`
}

// render helper function for Incident
func (i *Incident) Render(w http.ResponseWriter, r *http.Request) error { return nil }

// wrapper for easier use with API
type IncidentList struct {
	Timestamp time.Time           `json:"timestamp"` //timestamp that this list was generated/retrieved at
	Incidents map[string]Incident `json:"incidents"`
}

// render helper function for IncidentList
func (i *IncidentList) Render(w http.ResponseWriter, r *http.Request) error { return nil }

// struct representing system status, rougly based upon the atlassian statuspage format
type Status struct {
	OverallStatus   OverallStatus `json:"status"`
	ActiveIncidents []string      `json:"active_incidents"` //list of active incident IDs formatted as a slice of strings
}

// render helper function for Status
func (s *Status) Render(w http.ResponseWriter, r *http.Request) error { return nil }

type StatusWrapper struct {
	bun.BaseModel `bun:"table:status"`
	ID            int    `bun:",pk"`
	Status        Status `json:"status"`
}

/* Misc =-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=- */

// a type representing possible internal events
type EventType string

const (
	EventCreateIncident EventType = "create_incident"
	EventCreateUpdate   EventType = "create_update"
	EventEditIncident   EventType = "edit_incident"
	EventEditUpdate     EventType = "edit_update"
	EventDeleteIncident EventType = "delete_incident"
	EventDeleteUpdate   EventType = "delete_update"
)

// helper struct for internal events
type Event struct {
	Type     EventType
	Modified any
}
type WebhookMessage struct {
	bun.BaseModel `bun:"table:webhook_messages,alias:msg"`

	ID        string `bun:"id,pk"`
	Type      string `bun:"type"`
	MessageID int64  `bun:"message_id"`
}

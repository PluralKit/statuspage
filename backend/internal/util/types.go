package util

import (
	"net/http"
	"time"

	"github.com/uptrace/bun"
)

// a type representing the impact of an incident or event
type Impact string

const (
	ImpactNone  Impact = "none"
	ImpactMinor Impact = "minor"
	ImpactMajor Impact = "major"
)

// a type representing the status of an incident
type IncidentStatus string

const (
	StatusInvestigating IncidentStatus = "investigating"
	StatusIdentified    IncidentStatus = "identified"
	StatusMonitoring    IncidentStatus = "monitoring"
	StatusResolved      IncidentStatus = "resolved"
)

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

	ID        string    `json:"id" bun:"id,pk"`
	Text      string    `json:"text" bun:"text,notnull"`
	Timestamp time.Time `json:"timestamp" bun:"timestamp,notnull,default:current_timestamp"`

	IncidentID string `json:"-" bun:"incident_id,notnull"`
}

// struct representing a single incident, rougly based upon the atlassian statuspage format
type Incident struct {
	bun.BaseModel `bun:"table:incidents,alias:inc"`

	ID                  string         `json:"id" bun:"id,pk"`
	Timestamp           time.Time      `json:"timestamp" bun:"timestamp,nullzero,notnull,default:current_timestamp"`
	LastUpdate          time.Time      `json:"last_update" bun:"last_update,nullzero,notnull,default:current_timestamp"`
	ResolutionTimestamp time.Time      `json:"resolution_timestamp" bun:"resolution_timestamp,nullzero"`
	Status              IncidentStatus `json:"status" bun:"status"`
	Impact              Impact         `json:"impact" bun:"impact"`
	Name                string         `json:"name" bun:"name,notnull"`
	Description         string         `json:"description" bun:"description"`

	Updates []*IncidentUpdate `json:"updates" bun:"rel:has-many,join:id=incident_id"`
}

// wrapper for easier use with API
type IncidentList struct {
	Timestamp time.Time           `json:"timestamp"` //timestamp that this list was generated/retrieved at
	Incidents map[string]Incident `json:"incidents"`
}

// render helper function for IncidentList
func (i *IncidentList) Render(w http.ResponseWriter, r *http.Request) error { return nil }

// struct representing system status, rougly based upon the atlassian statuspage format
type Status struct {
	Status          OverallStatus `json:"status"`
	Impact          Impact        `json:"impact"`
	ActiveIncidents []string      `json:"active_incidents"` //list of active incident IDs formatted as a slice of strings
	Timestamp       time.Time     `json:"timestamp"`        //timestamp that this status report was generated/retrieved at
}

// render helper function for Status
func (s *Status) Render(w http.ResponseWriter, r *http.Request) error { return nil }

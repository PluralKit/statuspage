package api

import (
	"net/http"
	"time"
)

// a type representing the impact of an incident or event
type Impact string

const (
	ImpactNone  Impact = "none"
	ImpactMinor Impact = "minor"
	ImpactMajor Impact = "major"
)

type IncidentStatus string

const (
	StatusInvestigating IncidentStatus = "investigating"
	StatusIdentified    IncidentStatus = "identified"
	StatusMonitoring    IncidentStatus = "monitoring"
	StatusResolved      IncidentStatus = "resolved"
)

type IncidentUpdate struct {
	ID        string    `json:"id"`
	Text      string    `json:"text"`
	Timestamp time.Time `json:"timestamp"`
}

type OverallStatus string

const (
	StatusOperational OverallStatus = "operational"
	StatusDegraded    OverallStatus = "degraded"
	StatusMajorOutage OverallStatus = "major_outage"
)

// structure roughly based upon the atlassian status page format
type Incident struct {
	ID                  string           `json:"id"`
	Timestamp           time.Time        `json:"timestamp"`
	Status              IncidentStatus   `json:"status"`
	Impact              Impact           `json:"impact"`
	Updates             []IncidentUpdate `json:"updates"`
	Name                string           `json:"name"`
	Description         string           `json:"description"`
	LastUpdate          time.Time        `json:"last_update"`
	ResolutionTimestamp time.Time        `json:"resolution_timestamp"`
}

type IncidentList struct {
	Timestamp time.Time           `json:"timestamp"` //timestamp that this list was generated/retrieved at
	Incidents map[string]Incident `json:"incidents"`
}

func (i *IncidentList) Render(w http.ResponseWriter, r *http.Request) error { return nil }

// structure roughly based upon the atlassian status page format?
type Status struct {
	Status          OverallStatus `json:"status"`
	Impact          Impact        `json:"impact"`
	ActiveIncidents []string      `json:"active_incidents"` //list of active incident IDs formatted as a slice of strings
	Timestamp       time.Time     `json:"timestamp"`        //timestamp that this status report was generated/retrieved at
}

func (s *Status) Render(w http.ResponseWriter, r *http.Request) error { return nil }

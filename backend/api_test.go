package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"pluralkit/status/api"
	"pluralkit/status/db"
	"pluralkit/status/util"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAuthToken = "test_secret_token"

func setupTestAPI(t *testing.T) (*chi.Mux, *db.DB, func()) {
	cfg := util.Config{
		DBLoc:     "file::memory:?cache=shared",
		LogLevel:  util.SlogLevel(slog.LevelError),
		AuthToken: testAuthToken,
	}
	logger := slog.Default()
	eventChannel := make(chan util.Event, 10)

	database := db.NewDB(cfg, logger, eventChannel)
	require.NotNil(t, database)

	apiInstance := api.NewAPI(cfg, logger, database)
	router := chi.NewRouter()
	router.Use(render.SetContentType(render.ContentTypeJSON))
	apiInstance.SetupRoutes(router)

	teardown := func() {
		err := database.CloseDB()
		assert.NoError(t, err)
		close(eventChannel)
	}
	return router, database, teardown
}

func TestGetIncidents(t *testing.T) {
	router, dbInstance, teardown := setupTestAPI(t)
	defer teardown()

	ctx := context.Background()
	_, err := dbInstance.CreateIncident(ctx, util.Incident{Name: "1", Status: util.StatusInvestigating, Impact: util.ImpactMinor, Timestamp: time.Now().Add(-1 * time.Hour)})
	require.NoError(t, err)
	_, err = dbInstance.CreateIncident(ctx, util.Incident{Name: "2", Status: util.StatusResolved, Impact: util.ImpactNone, Timestamp: time.Now().Add(-2 * time.Hour)})
	require.NoError(t, err)

	t.Run("no 'before' param", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/incidents", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var list util.IncidentList
		err := json.NewDecoder(rr.Body).Decode(&list)
		require.NoError(t, err)
		assert.Len(t, list.Incidents, 2)
	})

	t.Run("with valid 'before' param", func(t *testing.T) {
		beforeTime := time.Now().Add(-61 * time.Minute).Format(time.RFC3339)
		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/incidents?before=%s", beforeTime), nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var list util.IncidentList
		err := json.NewDecoder(rr.Body).Decode(&list)
		require.NoError(t, err)
		assert.Len(t, list.Incidents, 1)
		found := false
		for _, inc := range list.Incidents {
			if inc.Name == "2" {
				found = true
				break
			}
		}
		assert.True(t, found)
	})

	t.Run("with invalid 'before' param", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/incidents?before=invalid-time", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestGetActiveIncidents(t *testing.T) {
	router, dbInstance, teardown := setupTestAPI(t)
	defer teardown()

	ctx := context.Background()
	_, err := dbInstance.CreateIncident(ctx, util.Incident{Name: "Active", Status: util.StatusInvestigating, Impact: util.ImpactMinor})
	require.NoError(t, err)
	_, err = dbInstance.CreateIncident(ctx, util.Incident{Name: "Resolved", Status: util.StatusResolved, Impact: util.ImpactNone})
	require.NoError(t, err)

	req, _ := http.NewRequest("GET", "/api/v1/incidents/active", nil)
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	var list util.IncidentList
	err = json.NewDecoder(rr.Body).Decode(&list)
	require.NoError(t, err)
	assert.Len(t, list.Incidents, 1)
	for _, inc := range list.Incidents {
		assert.Equal(t, "Active", inc.Name)
	}
}

func TestGetIncident(t *testing.T) {
	router, dbInstance, teardown := setupTestAPI(t)
	defer teardown()

	ctx := context.Background()
	id, err := dbInstance.CreateIncident(ctx, util.Incident{Name: "incident", Status: util.StatusInvestigating, Impact: util.ImpactMinor})
	require.NoError(t, err)

	t.Run("found", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/incidents/%s", id), nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var inc util.Incident
		err := json.NewDecoder(rr.Body).Decode(&inc)
		require.NoError(t, err)
		assert.Equal(t, "incident", inc.Name)
		assert.Equal(t, util.StatusInvestigating, inc.Status)
		assert.Equal(t, util.ImpactMinor, inc.Impact)
	})

	t.Run("not found", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/incidents/asdfasdf", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("invalid id", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/incidents/asdf", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})
}

func TestCreateIncident(t *testing.T) {
	router, _, teardown := setupTestAPI(t)
	defer teardown()

	incidentData := util.Incident{
		Name:        "incident",
		Description: "incident",
		Status:      util.StatusIdentified,
		Impact:      util.ImpactMajor,
	}
	body, _ := json.Marshal(incidentData)

	req, _ := http.NewRequest("POST", "/api/v1/admin/incidents/create", bytes.NewBuffer(body))
	req.Header.Set("Authorization", "Bearer "+testAuthToken)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)
	respBody, _ := io.ReadAll(rr.Body)
	assert.NotEmpty(t, string(respBody))
}

func TestEditIncident(t *testing.T) {
	router, dbInstance, teardown := setupTestAPI(t)
	defer teardown()

	ctx := context.Background()
	id, err := dbInstance.CreateIncident(ctx, util.Incident{Name: "not edited", Description: "not edited", Status: util.StatusInvestigating, Impact: util.ImpactMinor})
	require.NoError(t, err)

	name := "edited"
	description := "edited"
	status := util.StatusMonitoring
	impact := util.ImpactNone
	patchData := util.IncidentPatch{
		Name:        &name,
		Description: &description,
		Status:      &status,
		Impact:      &impact,
	}
	body, _ := json.Marshal(patchData)

	t.Run("not found", func(t *testing.T) {
		req, _ := http.NewRequest("PATCH", "/api/v1/admin/incidents/asdfasdf", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+testAuthToken)
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("found", func(t *testing.T) {
		req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/v1/admin/incidents/%s", id), bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+testAuthToken)
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		updatedInc, err := dbInstance.GetIncident(ctx, id)
		require.NoError(t, err)
		assert.Equal(t, name, updatedInc.Name)
		assert.Equal(t, description, updatedInc.Description)
		assert.Equal(t, status, updatedInc.Status)
		assert.Equal(t, impact, updatedInc.Impact)
	})
}

func TestDeleteIncident(t *testing.T) {
	router, dbInstance, teardown := setupTestAPI(t)
	defer teardown()

	ctx := context.Background()
	id, err := dbInstance.CreateIncident(ctx, util.Incident{Name: "not deleted yet", Status: util.StatusInvestigating, Impact: util.ImpactMinor})
	require.NoError(t, err)

	t.Run("not found", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/api/v1/admin/incidents/asdfasdf", nil)
		req.Header.Set("Authorization", "Bearer "+testAuthToken)
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("found", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/admin/incidents/%s", id), nil)
		req.Header.Set("Authorization", "Bearer "+testAuthToken)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		_, err = dbInstance.GetIncident(ctx, id)
		assert.Error(t, err)
	})
}

func TestAddUpdate(t *testing.T) {
	router, dbInstance, teardown := setupTestAPI(t)
	defer teardown()

	ctx := context.Background()
	incidentID, err := dbInstance.CreateIncident(ctx, util.Incident{Name: "incident", Status: util.StatusInvestigating, Impact: util.ImpactMinor})
	require.NoError(t, err)

	status := util.StatusInvestigating
	updateData := util.IncidentUpdate{
		Text:   "update",
		Status: &status,
	}
	body, _ := json.Marshal(updateData)

	t.Run("not found", func(t *testing.T) {
		req, _ := http.NewRequest("POST", "/api/v1/admin/incidents/asdfasdf/update", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+testAuthToken)
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("found", func(t *testing.T) {
		req, _ := http.NewRequest("POST", fmt.Sprintf("/api/v1/admin/incidents/%s/update", incidentID), bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+testAuthToken)
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		respBody, _ := io.ReadAll(rr.Body)
		assert.NotEmpty(t, string(respBody))

		updatedInc, err := dbInstance.GetIncident(ctx, incidentID)
		require.NoError(t, err)
		require.Len(t, updatedInc.Updates, 1)
		assert.Equal(t, "update", updatedInc.Updates[0].Text)
		assert.Equal(t, util.StatusInvestigating, *updatedInc.Updates[0].Status)
	})
}

func TestEditUpdate(t *testing.T) {
	router, dbInstance, teardown := setupTestAPI(t)
	defer teardown()

	ctx := context.Background()
	incidentID, err := dbInstance.CreateIncident(ctx, util.Incident{Name: "incident", Status: util.StatusInvestigating, Impact: util.ImpactMinor})
	require.NoError(t, err)
	status := util.StatusInvestigating
	updateID, err := dbInstance.CreateUpdate(ctx, util.IncidentUpdate{IncidentID: incidentID, Text: "not updated update", Status: &status})
	require.NoError(t, err)

	text := "updated update"
	newStatus := util.StatusIdentified
	patch := util.UpdatePatch{
		Text:   &text,
		Status: &newStatus,
	}
	body, _ := json.Marshal(patch)

	t.Run("not found", func(t *testing.T) {
		req, _ := http.NewRequest("PATCH", "/api/v1/admin/updates/asdfasdf", bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+testAuthToken)
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("found", func(t *testing.T) {
		req, _ := http.NewRequest("PATCH", fmt.Sprintf("/api/v1/admin/updates/%s", updateID), bytes.NewBuffer(body))
		req.Header.Set("Authorization", "Bearer "+testAuthToken)
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		editedUpdate, err := dbInstance.GetUpdate(ctx, updateID)
		require.NoError(t, err)
		assert.Equal(t, "updated update", editedUpdate.Text)
		assert.Equal(t, util.StatusIdentified, *editedUpdate.Status)
	})
}

func TestGetUpdate(t *testing.T) {
	router, dbInstance, teardown := setupTestAPI(t)
	defer teardown()

	ctx := context.Background()
	incidentID, err := dbInstance.CreateIncident(ctx, util.Incident{Name: "incident", Status: util.StatusInvestigating, Impact: util.ImpactMinor})
	require.NoError(t, err)
	updateID, err := dbInstance.CreateUpdate(ctx, util.IncidentUpdate{IncidentID: incidentID, Text: "update"})
	require.NoError(t, err)

	t.Run("found", func(t *testing.T) {
		req, _ := http.NewRequest("GET", fmt.Sprintf("/api/v1/updates/%s", updateID), nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		var upd util.IncidentUpdate
		err := json.NewDecoder(rr.Body).Decode(&upd)
		require.NoError(t, err)
		assert.Equal(t, "update", upd.Text)
	})

	t.Run("not found", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/api/v1/updates/nonexistentupdateid", nil)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})
}

func TestDeleteUpdate(t *testing.T) {
	router, dbInstance, teardown := setupTestAPI(t)
	defer teardown()

	ctx := context.Background()
	incidentID, err := dbInstance.CreateIncident(ctx, util.Incident{Name: "incident", Status: util.StatusInvestigating, Impact: util.ImpactMinor})
	require.NoError(t, err)
	updateID, err := dbInstance.CreateUpdate(ctx, util.IncidentUpdate{IncidentID: incidentID, Text: "update"})
	require.NoError(t, err)

	t.Run("not found", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", "/api/v1/admin/updates/asdfasdf", nil)
		req.Header.Set("Authorization", "Bearer "+testAuthToken)
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)
		assert.Equal(t, http.StatusNotFound, rr.Code)
	})

	t.Run("found", func(t *testing.T) {
		req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/admin/updates/%s", updateID), nil)
		req.Header.Set("Authorization", "Bearer "+testAuthToken)
		rr := httptest.NewRecorder()
		router.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)

		_, err = dbInstance.GetUpdate(ctx, updateID)
		assert.Error(t, err)
	})
}

func TestAdminRoutes_Unauthorized(t *testing.T) {
	router, _, teardown := setupTestAPI(t)
	defer teardown()

	endpoints := []struct {
		method string
		path   string
	}{
		{"POST", "/api/v1/admin/incidents/create"},
		{"PATCH", "/api/v1/admin/incidents/someid"},
		{"DELETE", "/api/v1/admin/incidents/someid"},
		{"POST", "/api/v1/admin/incidents/someid/update"},
		{"PATCH", "/api/v1/admin/updates/someupdateid"},
		{"DELETE", "/api/v1/admin/updates/someupdateid"},
	}

	for _, ep := range endpoints {
		t.Run(fmt.Sprintf("%s %s unauthorized", ep.method, ep.path), func(t *testing.T) {
			var reqBody io.Reader
			if ep.method == "POST" || ep.method == "PATCH" {
				reqBody = strings.NewReader("{}")
			}
			req, _ := http.NewRequest(ep.method, ep.path, reqBody)
			if reqBody != nil {
				req.Header.Set("Content-Type", "application/json")
			}

			rr := httptest.NewRecorder()
			router.ServeHTTP(rr, req)
			assert.Equal(t, http.StatusUnauthorized, rr.Code)

			reqWithWrongToken, _ := http.NewRequest(ep.method, ep.path, reqBody)
			if reqBody != nil {
				reqWithWrongToken.Header.Set("Content-Type", "application/json")
			}
			reqWithWrongToken.Header.Set("Authorization", "Bearer wrongtoken")
			rrWithWrongToken := httptest.NewRecorder()
			router.ServeHTTP(rrWithWrongToken, reqWithWrongToken)
			assert.Equal(t, http.StatusUnauthorized, rrWithWrongToken.Code)
		})
	}
}

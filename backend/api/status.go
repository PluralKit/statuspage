package api

import (
	"net/http"
	"pluralkit/status/util"
	"time"

	"github.com/go-chi/render"
)

type wrapper struct {
	util.Status
	Timestamp time.Time `json:"timestamp"`
}

func (a *API) GetStatus(w http.ResponseWriter, r *http.Request) {
	data := wrapper{
		util.Status{},
		time.Now(),
	}

	//TODO: do this in a better/more efficent way :3c
	incidents, err := a.Database.GetActiveIncidents(r.Context())
	if err != nil {
		http.Error(w, "error while checking status", 500)
	}
	highestImpact := util.ImpactNone
	for key, val := range incidents.Incidents {
		data.ActiveIncidents = append(data.ActiveIncidents, key)

		//janky, ik lol
		if val.Impact == util.ImpactMinor && highestImpact != util.ImpactMajor {
			highestImpact = util.ImpactMinor
		} else if val.Impact == util.ImpactMajor {
			highestImpact = util.ImpactMajor
		}
	}
	if highestImpact == util.ImpactMinor {
		data.OverallStatus = util.StatusDegraded
	} else if highestImpact == util.ImpactMajor {
		data.OverallStatus = util.StatusMajorOutage
	}

	if err := render.Render(w, r, &data); err != nil {
		http.Error(w, "error while rendering json", 500)
		return
	}
}

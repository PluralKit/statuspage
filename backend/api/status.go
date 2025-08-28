package api

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"pluralkit/status/util"
	"sort"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
)

type wrapper struct {
	util.Status
	Timestamp time.Time `json:"timestamp"`
}

type ShardsWrapper struct {
	Shards []Shard `json:"shards"`
}

func (a *API) GetStatus(w http.ResponseWriter, r *http.Request) {
	status, err := a.Database.GetStatus(r.Context())
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		a.Logger.Error("error while getting status", slog.Any("error", err))
		return
	}

	data := wrapper{
		status,
		time.Now(),
	}
	render.JSON(w, r, data)
}

const cacheTTL = 10 * time.Second

func (a *API) getClustersCached() (*ClustersInfo, error) {
	a.cacheMutex.RLock()
	validCache := time.Since(a.cacheTimestamp) < cacheTTL
	if validCache {
		a.cacheMutex.RUnlock()
		return &a.clustersCache, nil
	}
	a.cacheMutex.RUnlock()

	a.cacheMutex.Lock()
	defer a.cacheMutex.Unlock()
	validCache = time.Since(a.cacheTimestamp) < cacheTTL
	if validCache {
		return &a.clustersCache, nil
	}

	req, err := http.NewRequest(http.MethodGet, a.Config.ShardsEndpoint, nil)
	if err != nil {
		return nil, err
	}
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	} else if resp.StatusCode != 200 {
		return nil, err
	}
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	a.clustersCache.ShardsUp = 0
	a.clustersCache.AvgLatency = 0
	shards := ShardsWrapper{}
	err = json.Unmarshal(bodyBytes, &shards)
	if err != nil {
		return nil, err
	}

	sort.Slice(shards.Shards, func(i, j int) bool {
		return shards.Shards[i].ShardID < shards.Shards[j].ShardID
	})

	a.clustersCache.NumShards = len(shards.Shards)
	if len(a.clustersCache.Clusters) == 0 {
		a.clustersCache.Clusters = make([]*Cluster, a.clustersCache.NumShards/a.clustersCache.MaxConcurrency)
	}
	for key, shard := range shards.Shards {
		cluster := a.clustersCache.Clusters[shard.ClusterID]
		if key%a.Config.MaxConcurrency == 0 {
			if cluster == nil {
				cluster = &Cluster{
					Shards: make([]Shard, a.Config.MaxConcurrency),
				}
				a.clustersCache.Clusters[shard.ClusterID] = cluster
			} else {
				cluster.AvgLatency = 0
				cluster.ShardsUp = 0
				cluster.Shards = make([]Shard, a.Config.MaxConcurrency)
			}
		}
		cluster.AvgLatency += shard.Latency
		a.clustersCache.AvgLatency += shard.Latency
		if time.Since(shard.LastHeartbeat.Time) >= 10*time.Minute {
			shard.Up = false
		} else if shard.Up || time.Since(shard.LastReconnect.Time) <= 10*time.Second {
			a.clustersCache.ShardsUp++
			cluster.ShardsUp++
			shard.Up = true
		}
		cluster.Shards[shard.ShardID%a.Config.MaxConcurrency] = shard
	}

	for _, cluster := range a.clustersCache.Clusters {
		cluster.AvgLatency /= a.Config.MaxConcurrency
		if cluster.ShardsUp > (a.Config.MaxConcurrency / 2) {
			cluster.Up = true
		}
	}
	a.clustersCache.AvgLatency /= a.clustersCache.NumShards
	a.cacheTimestamp = time.Now()
	return &a.clustersCache, nil
}

func (a *API) GetClusters(w http.ResponseWriter, r *http.Request) {
	clusters, err := a.getClustersCached()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		a.Logger.Error("error while getting clusters", slog.Any("error", err))
	}
	render.JSON(w, r, clusters)
}

func (a *API) GetShards(w http.ResponseWriter, r *http.Request) {
	clusters, err := a.getClustersCached()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		a.Logger.Error("error while getting clusters", slog.Any("error", err))
	}
	index, err := strconv.Atoi(chi.URLParam(r, "clusterID"))
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadRequest), http.StatusBadRequest)
	}
	render.JSON(w, r, clusters.Clusters[index].Shards)
}

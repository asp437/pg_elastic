package api

import (
	"github.com/asp437/pg_elastic/server"
	"net/http"
)

// ClusterHealth describes JSON schema for _cluster/health requests
type clusterHealth struct {
	Name                        string  `json:"cluster_name"`
	Status                      string  `json:"status"`
	TimedOut                    bool    `json:"timed_out"`
	NumberOfNodes               int     `json:"number_of_nodes"`
	ActivePrimaryShards         int     `json:"active_primary_shards"`
	ActiveShards                int     `json:"active_shards"`
	RelocatingShards            int     `json:"relocating_shards"`
	InitializingShards          int     `json:"initializing_shards"`
	UnassignedShards            int     `json:"unassigned_shards"`
	DelayedUnassignedShards     int     `json:"delayed_unassigned_shards"`
	PendingTasks                int     `json:"number_of_pending_tasks"`
	InFlightFetch               int     `json:"number_of_in_flight_fetch"`
	TaskMaxWaitingInQueueMillis int     `json:"task_max_waiting_in_queue_millis"`
	ActiveShardsPercent         float32 `json:"active_shards_percent_as_number"`
}

// HealthHandler process a health-check response
func HealthHandler(endpoint string, r *http.Request, server server.PGElasticServer) (interface{}, error) {
	health := clusterHealth{}
	health.Name = "pg_elastic_cluster"
	health.Status = "yellow"
	health.NumberOfNodes = 1
	return health, nil
}

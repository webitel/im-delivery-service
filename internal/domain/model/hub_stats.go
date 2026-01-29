package model

import "time"

type HubStats struct {
	TotalUsers       int           `json:"total_users"`
	TotalConnections int           `json:"total_connections"`
	Uptime           time.Duration `json:"uptime"`
	Shards           []ShardStats  `json:"shards,omitempty"`
}

type ShardStats struct {
	ShardID     int `json:"shard_id"`
	UserCount   int `json:"user_count"`
	ActiveCells int `json:"active_cells"`
}

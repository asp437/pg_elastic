package server

import (
	"github.com/asp437/pg_elastic/db"
	"github.com/asp437/pg_elastic/utils"
)

// PGElasticServer represents a HTTP server which provides an ElasticSearch like API
type PGElasticServer interface {
	GetConfiguration() utils.PGElasticConfig
	GetDBClient() *db.Client
	Start()
}

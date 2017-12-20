package server

import (
	"github.com/asp437/pg_elastic/db"
	"github.com/asp437/pg_elastic/utils"
)

type PGElasticServer interface {
	GetConfiguration() utils.PGElasticConfig
	GetDBClient() *db.Client
	Start()
}

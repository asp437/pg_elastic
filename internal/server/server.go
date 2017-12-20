package server

import (
	"errors"
	"github.com/asp437/pg_elastic/api"
	"github.com/asp437/pg_elastic/db"
	"github.com/asp437/pg_elastic/server"
	"github.com/asp437/pg_elastic/utils"
	"log"
	"net/http"
	"regexp"
	"strconv"
)

type PGElasticServerProto struct {
	handler  *server.ElasticHandler
	config   *utils.PGElasticConfig
	dbclient *db.Client
}

// Create an instance of server. Configuration should be loaded from file configFileName
func InitializeServer(configFileName string) (server.PGElasticServer, error) {
	s := new(PGElasticServerProto)
	s.config = utils.ReadConfig(configFileName)
	s.handler = server.NewElasticHandler(s)
	s.dbclient = db.CreateClient(s.config.PostgresConfig)

	if s.dbclient == nil {
		return nil, errors.New("Database connection is not established")
	}
	err := s.dbclient.InitializeSchema()
	if err != nil {
		return nil, err
	}
	s.configureHandler()
	return s, nil
}

func (s *PGElasticServerProto) GetConfiguration() utils.PGElasticConfig {
	return *s.config
}

func (s *PGElasticServerProto) GetDBClient() *db.Client {
	return s.dbclient
}

// Start the server
func (s *PGElasticServerProto) Start() {
	log.Printf("starting server, listening on port %d\n", s.config.ServerPort)
	http.ListenAndServe(":"+strconv.Itoa(s.config.ServerPort), s.handler)
}

// Configuration of all handlers of the server
func (s *PGElasticServerProto) configureHandler() {
	s.handler.HandleFunc(regexp.MustCompile("^/_cluster/health"), api.HealthHandler, []string{"GET"})
	s.handler.HandleFunc(regexp.MustCompile("^/_bulk"), api.BulkHandler, []string{"GET"})

	s.handler.HandleFunc(regexp.MustCompile("^/[^_][\\d\\w]*/_mapping/[\\d\\w]+"), api.PutTypeMapping, []string{"PUT"})

	s.handler.HandleFunc(regexp.MustCompile("^/[^_][\\d\\w]*"), api.PutIndexHandler, []string{"PUT"})
	s.handler.HandleFunc(regexp.MustCompile("^/[^_][\\d\\w]*"), api.HeadIndexHandler, []string{"HEAD"})

	s.handler.HandleFunc(regexp.MustCompile("^/[^_][\\d\\w]*/_search"), api.FindIndexDocumentHandler, []string{"GET"})
	s.handler.HandleFuncEndpoint(regexp.MustCompile("^_search"), api.FindDocumentHandler, []string{"GET"})

	s.handler.HandleFuncEndpoint(regexp.MustCompile("^[\\d\\w]*"), api.PutDocumentHandler, []string{"PUT", "POST"})
	s.handler.HandleFuncEndpoint(regexp.MustCompile("^[\\d\\w]+"), api.GetDocumentHandler, []string{"GET"})
	s.handler.HandleFuncEndpoint(regexp.MustCompile("^[\\d\\w]+"), api.DeleteDocumentHandler, []string{"DELETE"})
}

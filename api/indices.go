package api

import (
	"fmt"
	"github.com/asp437/pg_elastic/server"
	"github.com/asp437/pg_elastic/utils"
	"io/ioutil"
	"net/http"
	"regexp"
)

type indexPutResponse struct {
	Acknowledged       bool `json:"acknowledged"`
	ShardsAcknowledged bool `json:"shards_acknowledged"`
}

type typePutResponse struct {
	Acknowledged bool `json:"acknowledged"`
}

var putTypeMappingPattern = regexp.MustCompile("/(?P<index>\\w+)/_mapping/(?P<type>\\w+)")
var indexHandlerPattern = regexp.MustCompile("/(?P<index>\\w+)")

func PutIndexHandler(endpoint string, r *http.Request, server server.PGElasticServer) (interface{}, error) {
	indexName := indexHandlerPattern.ReplaceAllString(endpoint, "${index}")
	optionsBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, utils.NewInternalIOError(err.Error())
	}
	options := string(optionsBytes)

	_, err = server.GetDBClient().CreateIndex(indexName, options)
	if err != nil {
		return nil, err
	}

	return indexPutResponse{true, true}, nil
}

func HeadIndexHandler(endpoint string, r *http.Request, server server.PGElasticServer) (interface{}, error) {
	indexName := indexHandlerPattern.ReplaceAllString(endpoint, "${index}")
	indexRecord, err := server.GetDBClient().GetIndex(indexName)
	if err != nil {
		fmt.Println(err)
		return false, err
	}
	return indexRecord != nil, nil
}

func PutTypeMapping(endpoint string, r *http.Request, server server.PGElasticServer) (interface{}, error) {
	indexName := putTypeMappingPattern.ReplaceAllString(endpoint, "${index}")
	typeName := putTypeMappingPattern.ReplaceAllString(endpoint, "${type}")
	optionsBytes, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, utils.NewInternalIOError(err.Error())
	}
	options := string(optionsBytes)

	typeObject, err := server.GetDBClient().GetType(indexName, typeName)
	if err != nil {
		return nil, err
	}

	if typeObject != nil {
		_, err = server.GetDBClient().UpdateTypeOptions(indexName, typeName, options)
	} else {
		_, err = server.GetDBClient().CreateType(indexName, typeName, options)
	}

	if err != nil {
		return nil, err
	}
	return typePutResponse{true}, nil
}

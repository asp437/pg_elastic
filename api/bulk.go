package api

import (
	"encoding/json"
	"github.com/asp437/pg_elastic/db"
	"github.com/asp437/pg_elastic/server"
	"github.com/asp437/pg_elastic/utils"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

type bulkResponse struct {
	Took   int           `json:"took"`
	Errors bool          `json:"errors"`
	Items  []interface{} `json:"items"`
}

type bulkPutCommandResponse struct {
	documentPutResponse
	Status int `json:"status"`
}

type bulkGetCommandResponse struct {
	documentGetResponse
	Status int `json:"status"`
}

type bulkIndexResponse struct {
	Index bulkPutCommandResponse `json:"index"`
}
type bulkCreateResponse struct {
	Create bulkPutCommandResponse `json:"create"`
}
type bulkUpdateResponse struct {
	Update bulkPutCommandResponse `json:"update"`
}
type bulkDeleteResponse struct {
	Delete bulkGetCommandResponse `json:"delete"`
}

func BulkHandler(endpoint string, r *http.Request, server server.PGElasticServer) (response interface{}, err error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, utils.NewInternalIOError(err.Error())
	}

	str := string(body)
	bulkCommands := strings.Split(str, "\n")
	return ProcessBulkQuery(bulkCommands, server)
}

func ProcessBulkQuery(rawQuery []string, server server.PGElasticServer) (interface{}, error) {
	response := bulkResponse{}
	response.Errors = false
	skip := false
	startTime := time.Now()
	for i, command := range rawQuery {
		if skip || strings.Compare(command, "") == 0 {
			skip = false
			continue
		}
		var parsedJson map[string]interface{}
		var documentObject *db.ElasticSearchDocument

		err := json.Unmarshal([]byte(command), &parsedJson)
		if err != nil {
			return nil, utils.NewJSONWrongFormatError(err.Error())
		}

		for k, v := range parsedJson {
			responseCommand := make(map[string]interface{})
			if _, ok := v.(map[string]interface{}); !ok {
				return nil, utils.NewJSONWrongFormatError("Wrong JSON format")
			}
			indexDescriptor := v.(map[string]interface{})
			indexName := indexDescriptor["_index"].(string)
			typeName := indexDescriptor["_type"].(string)
			id := indexDescriptor["_id"].(string)

			switch k {
			case "index":
				documentObject, err := server.GetDBClient().GetDocument(indexName, typeName, id)
				if err != nil {
					response.Errors = true
					return nil, err
				}
				if documentObject == nil {
					documentObject, err = server.GetDBClient().CreateDocument(indexName, typeName, rawQuery[i+1], id)
				} else {
					documentObject, err = server.GetDBClient().UpdateDocument(indexName, typeName, rawQuery[i+1], id)
				}
				skip = true
			case "create":
				documentObject, err = server.GetDBClient().CreateDocument(indexName, typeName, rawQuery[i+1], id)
				skip = true
			case "update":
				documentObject, err = server.GetDBClient().UpdateDocument(indexName, typeName, rawQuery[i+1], id)
				skip = true
			case "delete":
				documentObject, err = server.GetDBClient().DeleteDocument(indexName, typeName, id)
			}
			if err != nil {
				response.Errors = true
				if _, ok := err.(utils.ElasticError); ok {
					responseCommand[k] = utils.NewElasticErrorBulk(err.(utils.ElasticError), indexName, "1", "1").FormatErrorResponse()
				} else {
					return nil, err
				}
			}
			switch k {
			case "index", "create", "update":
				if documentObject != nil {
					command := bulkPutCommandResponse{
						documentPutResponse{
							Shards:  shardInfo{1, 0, 1},
							Index:   indexName,
							Type:    typeName,
							ID:      documentObject.ID,
							Version: documentObject.Version,
							Created: documentObject.Version == 1,
							Result: func() string {
								if documentObject.Version == 1 {
									return "created"
								}
								return "updated"
							}(),
						},
						200,
					}
					responseCommand[k] = command
				} else if err == nil {
					responseCommand[k] = utils.NewElasticErrorBulk(utils.NewInternalError("Unknown error"), indexName, "1", "1").FormatErrorResponse()
					response.Errors = true
				}
			case "delete":
				command := bulkGetCommandResponse{
					documentGetResponse{
						Index: indexName,
						Type:  typeName,
						Found: documentObject != nil,
					},
					200,
				}
				if documentObject != nil {
					command.Version = documentObject.Version
					command.Document = documentObject.Document
					command.ID = documentObject.ID
				}
				responseCommand["delete"] = command
			}
			response.Items = append(response.Items, responseCommand)
		}
	}
	response.Took = (int)(time.Since(startTime).Nanoseconds() / 1000000.0)

	return response, nil
}

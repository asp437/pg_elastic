package api

import (
	"encoding/json"
	"github.com/asp437/pg_elastic/api/search"
	"github.com/asp437/pg_elastic/db"
	"github.com/asp437/pg_elastic/server"
	"github.com/asp437/pg_elastic/utils"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
	"time"
)

type shardInfo struct {
	Total      int `json:"total"`
	Failed     int `json:"failed"`
	Successful int `json:"successful"`
}

type searchHits struct {
	MaxScore float32                  `json:"max_score"`
	Total    int                      `json:"total"`
	Hits     []documentSearchResponse `json:"hits"`
}

type documentPutResponse struct {
	Shards  shardInfo `json:"_shards"`
	Index   string    `json:"_index"`
	Type    string    `json:"_type"`
	ID      string    `json:"_id"`
	Version int       `json:"_version"`
	Created bool      `json:"created"`
	Result  string    `json:"result"`
}

type documentGetResponse struct {
	Index    string      `json:"_index"`
	Type     string      `json:"_type"`
	ID       string      `json:"_id"`
	Version  int         `json:"_version"`
	Found    bool        `json:"found"`
	Document interface{} `json:"_source"`
}

type documentSearchResponse struct {
	Index    string      `json:"_index"`
	Type     string      `json:"_type"`
	ID       string      `json:"_id"`
	Score    float32     `json:"_score"`
	Document interface{} `json:"_source"`
}

type searchResponse struct {
	Took     int        `json:"took"`
	TimedOut bool       `json:"timed_out"`
	Shards   shardInfo  `json:"_shards"`
	Hits     searchHits `json:"hits"`
}

func formatDocumentSearchResponse(index, typeName string, doc db.ElasticSearchDocument) documentSearchResponse {
	return documentSearchResponse{
		Index:    index,
		Type:     typeName,
		ID:       doc.ID,
		Score:    1.0,
		Document: doc.Document,
	}
}

// PutDocumentHandler handles request to put document into storage
func PutDocumentHandler(index, typeName, endpoint string, r *http.Request, s server.PGElasticServer) (interface{}, error) {
	var documentObject *db.ElasticSearchDocument
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, utils.NewInternalIOError(err.Error())
	}
	if strings.Compare(endpoint, "") == 0 {
		documentObject, err = s.GetDBClient().CreateDocument(index, typeName, string(body), "")
		if err != nil {
			return nil, err
		}
	} else {
		documentID := endpoint
		exists, err := s.GetDBClient().IsDocumentExists(index, typeName, documentID)
		if err != nil {
			return nil, err
		}
		if exists {
			documentObject, err = s.GetDBClient().UpdateDocument(index, typeName, string(body), documentID)
		} else {
			documentObject, err = s.GetDBClient().CreateDocument(index, typeName, string(body), documentID)
		}
		if err != nil {
			return nil, err
		}
	}

	response := documentPutResponse{
		Shards:  shardInfo{1, 0, 1},
		Index:   index,
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
	}

	return response, nil
}

// GetDocumentHandler handles request to get document from storage
func GetDocumentHandler(index, typeName, endpoint string, r *http.Request, s server.PGElasticServer) (response interface{}, err error) {
	documentID := endpoint
	documentObject, err := s.GetDBClient().GetDocument(index, typeName, documentID)
	if err != nil {
		return nil, err
	}
	if documentObject != nil {
		response = documentGetResponse{
			Index:    index,
			Type:     typeName,
			ID:       documentObject.ID,
			Version:  documentObject.Version,
			Found:    true,
			Document: documentObject.Document,
		}
	} else {
		response = documentGetResponse{
			Index: index,
			Type:  typeName,
			Found: false,
		}

	}
	return response, nil
}

// DeleteDocumentHandler handles request to delete document from storage
func DeleteDocumentHandler(index, typeName, endpoint string, r *http.Request, s server.PGElasticServer) (response interface{}, err error) {
	documentID := endpoint
	documentObject, err := s.GetDBClient().DeleteDocument(index, typeName, documentID)
	if err != nil {
		return nil, err
	}
	if documentObject != nil {
		response = documentGetResponse{
			Index:    index,
			Type:     typeName,
			ID:       documentObject.ID,
			Version:  documentObject.Version,
			Found:    true,
			Document: documentObject.Document,
		}
	} else {
		response = documentGetResponse{
			Index: index,
			Type:  typeName,
			Found: false,
		}

	}
	return response, nil
}

// FindDocumentHandler handles request to find document on storage
func FindDocumentHandler(indexPattern, typePattern, endpoint string, r *http.Request, s server.PGElasticServer) (response interface{}, err error) {
	startTime := time.Now()
	var parsedQuery interface{}
	indices, err := s.GetDBClient().FindIndices(indexPattern)
	if err != nil {
		return nil, utils.NewInternalError(err.Error())
	}
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, utils.NewInternalIOError(err.Error())
	}

	err = json.Unmarshal(body, &parsedQuery)
	if err != nil {
		return nil, utils.NewJSONWrongFormatError(err.Error())
	}

	for _, index := range indices {
		types, err := s.GetDBClient().FindTypes(index, typePattern)
		if err != nil {
			return nil, utils.NewInternalError(err.Error())
		}
		for _, typeName := range types {
			for k, v := range parsedQuery.(map[string]interface{}) {
				switch k {
				case "query":
					var typeMapping map[string]interface{}

					// Get type mapping from system type record
					docType, err := s.GetDBClient().GetType(index, typeName)
					if err != nil {
						return nil, err
					}
					json.Unmarshal([]byte(docType.Options), &typeMapping)

					query := s.GetDBClient().NewQuery(index, typeName)
					search.ParseSearchQuery(v.(map[string]interface{}), query, typeMapping)
					docs, err := s.GetDBClient().ProcessSearchQuery(index, typeName, query)
					if err != nil {
						return nil, err
					}
					response := searchResponse{
						Took:     1,
						TimedOut: false,
						Shards:   shardInfo{1, 0, 1},
						Hits: searchHits{
							MaxScore: 0,
							Total:    0,
							Hits:     []documentSearchResponse{},
						},
					}
					for _, doc := range docs {
						docResponse := formatDocumentSearchResponse(index, typeName, doc)
						response.Hits.Hits = append(response.Hits.Hits, docResponse)
						if response.Hits.MaxScore < docResponse.Score {
							response.Hits.MaxScore = docResponse.Score
						}
						response.Hits.Total += 1
					}
					response.Took = (int)(time.Since(startTime).Nanoseconds() / 1000000.0)
					return response, nil
				}
			}
		}
	}
	return nil, utils.NewIllegalQueryError("Illegal search query")
}

// FindIndexDocumentHandler handles request to find document of any type on storage
func FindIndexDocumentHandler(endpoint string, r *http.Request, s server.PGElasticServer) (response interface{}, err error) {
	var indexHandlerPattern = regexp.MustCompile("^/(?P<index>\\w+)/_search")
	indexName := indexHandlerPattern.ReplaceAllString(endpoint, "${index}")
	return FindDocumentHandler(indexName, "*", endpoint, r, s)
}

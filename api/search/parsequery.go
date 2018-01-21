package search

import (
	"fmt"
	"github.com/asp437/pg_elastic/db"
	"github.com/asp437/pg_elastic/utils"
)

// ParseSearchQuery parses a query and convert it into db.Query
func ParseSearchQuery(rawQuery map[string]interface{}, query *db.Query, mapping map[string]interface{}) {
	for k, v := range rawQuery {
		switch k {
		case "match_all":
			parseMatchAllQuery(v.(map[string]interface{}), query, mapping)
		case "match":
			parseMatchQuery(v.(map[string]interface{}), query, mapping)
		case "match_phrase":
			parseMatchPhraseQuery(v.(map[string]interface{}), query, mapping)
		case "bool":
			parseBoolQuery(v.(map[string]interface{}), query, mapping)
		}
	}
}

func parseMatchAllQuery(rawQuery map[string]interface{}, query *db.Query, mapping map[string]interface{}) {
}

func parseMatchPhraseQuery(rawQuery map[string]interface{}, query *db.Query, mapping map[string]interface{}) {
	var whereClause string
	for k, v := range rawQuery {
		fieldName := k
		switch v.(type) {
		case string:
			fieldMapping, ok := utils.GetFieldMapping(mapping, fieldName)
			if ok && len(fieldMapping.Analyzer) > 0 {
				whereClause = fmt.Sprintf("to_tsvector('%s', document->'%s') @@ phraseto_tsquery('%s', '%s')", fieldMapping.Analyzer, fieldName, fieldMapping.Analyzer, v)
			} else {
				whereClause = fmt.Sprintf("to_tsvector(document->'%s') @@ phraseto_tsquery('%s')", fieldName, v)
			}
			query.Where(whereClause)
		case map[string]interface{}:
			var queryString, operator string
			for kk, vv := range v.(map[string]interface{}) {
				switch kk {
				case "query":
					queryString = vv.(string)
				case "operator":
					operator = vv.(string)
				}
			}
			operator = operator // Silent not used variable
			fieldMapping, ok := utils.GetFieldMapping(mapping, fieldName)
			if ok && len(fieldMapping.Analyzer) > 0 {
				whereClause = fmt.Sprintf("to_tsvector('%s', document->'%s') @@ phraseto_tsquery('%s', '%s')", fieldMapping.Analyzer, fieldName, fieldMapping.Analyzer, queryString)
			} else {
				whereClause = fmt.Sprintf("to_tsvector(document->'%s') @@ phraseto_tsquery('%s')", fieldName, queryString)
			}
			query.Where(whereClause)
		}
	}
}

func parseMatchQuery(rawQuery map[string]interface{}, query *db.Query, mapping map[string]interface{}) {
	var whereClause string
	for k, v := range rawQuery {
		fieldName := k
		switch v.(type) {
		case string:
			fieldMapping, ok := utils.GetFieldMapping(mapping, fieldName)
			if ok && len(fieldMapping.Analyzer) > 0 {
				whereClause = fmt.Sprintf("to_tsvector('%s', document->'%s') @@ to_tsquery('%s', '%s')", fieldMapping.Analyzer, fieldName, fieldMapping.Analyzer, v)
			} else {
				whereClause = fmt.Sprintf("to_tsvector(document->'%s') @@ to_tsquery('%s')", fieldName, v)
			}
			query.Where(whereClause)
		case map[string]interface{}:
			var queryString, operator string
			for kk, vv := range v.(map[string]interface{}) {
				switch kk {
				case "query":
					queryString = vv.(string)
				case "operator":
					operator = vv.(string)
				}
			}
			operator = operator // Silent not used variable
			fieldMapping, ok := utils.GetFieldMapping(mapping, fieldName)
			if ok && len(fieldMapping.Analyzer) > 0 {
				whereClause = fmt.Sprintf("to_tsvector('%s', document->'%s') @@ to_tsquery('%s', '%s')", fieldMapping.Analyzer, fieldName, fieldMapping.Analyzer, queryString)
			} else {
				whereClause = fmt.Sprintf("to_tsvector(document->'%s') @@ to_tsquery('%s')", fieldName, queryString)
			}
			query.Where(whereClause)
		}
	}
}

func parseBoolQuery(rawQuery map[string]interface{}, query *db.Query, mapping map[string]interface{}) {
	for k, v := range rawQuery {
		switch k {
		case "must":
			query.WhereGroup(func(q *db.Query) (*db.Query, error) {
				ParseSearchQuery(v.(map[string]interface{}), q, mapping)
				return q, nil
			})
		case "filter":
			query.WhereGroup(func(q *db.Query) (*db.Query, error) {
				ParseSearchQuery(v.(map[string]interface{}), q, mapping)
				return q, nil
			})
		case "must_not":
			// TODO: must_not query negation
			query.WhereGroup(func(q *db.Query) (*db.Query, error) {
				ParseSearchQuery(v.(map[string]interface{}), q, mapping)
				return q, nil
			})
		case "should":
			query.WhereOrGroup(func(q *db.Query) (*db.Query, error) {
				ParseSearchQuery(v.(map[string]interface{}), q, mapping)
				return q, nil
			})
		}
	}
}

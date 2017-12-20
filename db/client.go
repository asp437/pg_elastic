package db

import (
	"fmt"
	"github.com/asp437/pg_elastic/utils"
	"github.com/go-pg/pg"
	"github.com/go-pg/pg/orm"
	"strings"
)

type Client struct {
	connection *pg.DB
}

type Query = orm.Query

type IndexRecord struct {
	Name    string
	Options string
}

type TypeRecord struct {
	Name      string
	IndexName string
	Options   string
}

type ElasticSearchDocument struct {
	ID       string
	Document interface{}
	Version  int
}

/*
 * General Client API
 */

// Create an instance of Client with specified parameters
// Doesn't check connection to DB server
func CreateClient(config utils.PostgresConnectionConfig) (result *Client) {
	result = new(Client)
	result.connection = pg.Connect(&pg.Options{
		User:     config.User,
		Addr:     config.ServerAddress,
		Password: config.Password,
		Database: config.DBName,
	})
	return result
}

// Initialize system tables used by pg_elastic
// Doesn't affect existing tables
func (dbc *Client) InitializeSchema() error {
	for _, model := range []interface{}{&IndexRecord{}, &TypeRecord{}} {
		err := dbc.connection.CreateTable(model, &orm.CreateTableOptions{IfNotExists: true})
		if err != nil {
			return err
		}
	}
	return nil
}

// Create Query instance for specified index and type
func (dbc *Client) NewQuery(indexName, typeName string) *Query {
	tableName := fmt.Sprintf("%s_%s", indexName, typeName)
	return dbc.connection.Model().Table(tableName)
}

// Processing ElasticSearch-format query
func (dbc *Client) ProcessSearchQuery(indexName, typeName string, query *Query) ([]ElasticSearchDocument, error) {
	var documentObject []ElasticSearchDocument

	c, err := query.Count()
	if err != nil {
		return nil, utils.NewDBQueryError(err.Error())
	}
	if c == 0 {
		return nil, nil
	}

	err = query.Select(&documentObject)
	if err != nil {
		return nil, utils.NewDBQueryError(err.Error())
	}

	return documentObject, nil
}

/*
 * Indices API
 */

// Create an index record with specified options
func (dbc *Client) CreateIndex(indexName, options string) (*IndexRecord, error) {
	var indexRecord IndexRecord
	indexSelectQuery := dbc.connection.Model(&IndexRecord{}).Where("Name = ?", indexName)
	count, err := indexSelectQuery.Count()
	if err != nil {
		return nil, utils.NewDBQueryError(err.Error())
	}
	if count == 0 {
		indexRecord = IndexRecord{Name: indexName, Options: options}
		err = dbc.connection.Insert(&indexRecord)
		if err != nil {
			return nil, utils.NewDBQueryError(err.Error())
		}
	} else {
		return nil, utils.NewIllegalQueryError("Index already exists")
	}
	indexSelectQuery.Select(&indexRecord)
	return &indexRecord, nil
}

// Get an instance of existing index
func (dbc *Client) GetIndex(indexName string) (*IndexRecord, error) {
	var indexRecord IndexRecord
	indexSelectQuery := dbc.connection.Model(&IndexRecord{}).Where("Name = ?", indexName)
	count, err := indexSelectQuery.Count()
	if err != nil {
		return nil, utils.NewDBQueryError(err.Error())
	}
	if count == 0 {
		return nil, nil
	}
	indexSelectQuery.Select(&indexRecord)
	if err != nil {
		return nil, utils.NewDBQueryError(err.Error())
	}

	return &indexRecord, nil
}

// Search for indicies using name pattern in ElasticSearch wildcard format
func (dbc *Client) FindIndices(indexPattern string) ([]string, error) {
	var records []IndexRecord
	var results []string
	indexPattern = strings.Replace(indexPattern, "?", "_", -1)
	indexPattern = strings.Replace(indexPattern, "*", "%", -1)
	query := dbc.connection.Model(&IndexRecord{}).Where("name LIKE ?", indexPattern)
	err := query.Select(&records)
	if err != nil {
		return nil, err
	}
	for _, v := range records {
		results = append(results, v.Name)
	}
	return results, nil
}

/*
 * Types API
 */

// Create a type record with specified options
func (dbc *Client) CreateType(indexName, typeName, options string) (*TypeRecord, error) {
	var typeRecord TypeRecord
	typeSelectQuery := dbc.connection.Model(&TypeRecord{}).Where("Name = ?", typeName).Where("Index_Name = ?", indexName)
	count, err := typeSelectQuery.Count()
	if err != nil {
		return nil, utils.NewDBQueryError(err.Error())
	}
	if count == 0 {
		typeRecord = TypeRecord{Name: typeName, IndexName: indexName, Options: options}
		err = dbc.connection.Insert(&typeRecord)
		if err != nil {
			return nil, utils.NewDBQueryError(err.Error())
		}
		err = dbc.createDataTable(indexName, typeName)
		if err != nil {
			return nil, utils.NewDBQueryError(err.Error())
		}
	} else {
		return nil, utils.NewIllegalQueryError("Type already exists")
	}
	return &typeRecord, nil
}

// Get an instance of existing type
func (dbc *Client) GetType(indexName, typeName string) (*TypeRecord, error) {
	var typeRecord TypeRecord
	typeSelectQuery := dbc.connection.Model(&TypeRecord{}).Where("Name = ?", typeName).Where("Index_Name = ?", indexName)
	count, err := typeSelectQuery.Count()
	if err != nil {
		return nil, utils.NewDBQueryError(err.Error())
	}
	if count == 0 {
		return nil, nil
	}
	err = typeSelectQuery.Select(&typeRecord)
	if err != nil {
		return nil, utils.NewDBQueryError(err.Error())
	}
	return &typeRecord, nil

}

// Update options for exiting type
func (dbc *Client) UpdateTypeOptions(indexName, typeName, options string) (*TypeRecord, error) {
	_, err := dbc.connection.Model(&TypeRecord{}).Where("Name = ?", typeName).Where("Index_Name = ?", indexName).Set("options = ?", options).Update()
	if err != nil {
		return nil, utils.NewDBQueryError(err.Error())
	}
	return dbc.GetType(indexName, typeName)
}

// Search for indicies using name pattern in ElasticSearch wildcard format
func (dbc *Client) FindTypes(index, typePattern string) ([]string, error) {
	var records []TypeRecord
	var results []string
	typePattern = strings.Replace(typePattern, "?", "_", -1)
	typePattern = strings.Replace(typePattern, "*", "%", -1)
	query := dbc.connection.Model(&TypeRecord{}).Where("name LIKE ?", typePattern).Where("index_name = ?", index)
	err := query.Select(&records)
	if err != nil {
		return nil, err
	}
	for _, v := range records {
		results = append(results, v.Name)
	}
	return results, nil
}

/*
 * Documents API
 */

// Create a new document in database
func (dbc *Client) CreateDocument(indexName, typeName, document string, documentID string) (result *ElasticSearchDocument, err error) {
	index, err := dbc.GetIndex(indexName)
	if err != nil {
		return nil, err
	} else if index == nil {
		_, err = dbc.CreateIndex(indexName, "")
		if err != nil {
			return nil, err
		}
	}

	typeObject, err := dbc.GetType(indexName, typeName)
	if err != nil {
		return nil, err
	} else if typeObject == nil {
		_, err = dbc.CreateType(indexName, typeName, "")
		if err != nil {
			return nil, err
		}
	}

	if len(documentID) == 0 {
		result, err = dbc.insertDocument(indexName, typeName, document)
		if err != nil {
			return nil, err
		}
	} else {
		documentExist, err := dbc.IsDocumentExists(indexName, typeName, documentID)
		if err != nil {
			return nil, err
		}
		if documentExist {
			return nil, utils.NewDBQueryError(fmt.Sprintf("Document with ID %s already exists", documentID))
		} else {
			result, err = dbc.insertDocumentID(indexName, typeName, document, documentID)
			if err != nil {
				return nil, err
			}
		}
	}
	return result, nil
}

// Get document specified by index, type, and ID
func (dbc *Client) GetDocument(indexName, typeName string, documentID string) (*ElasticSearchDocument, error) {
	var documentObject ElasticSearchDocument

	tableName := fmt.Sprintf("%s_%s", indexName, typeName)
	query := dbc.connection.Model().TableExpr(tableName).Where("id = ?", documentID)

	c, err := query.Count()
	if err != nil {
		return nil, utils.NewDBQueryError(err.Error())
	}
	if c == 0 {
		return nil, nil
	}

	err = query.Select(&documentObject)
	if err != nil {
		return nil, utils.NewDBQueryError(err.Error())
	}

	return &documentObject, nil
}

// Check for document existance in database
func (dbc *Client) IsDocumentExists(indexName, typeName string, documentID string) (bool, error) {
	if len(documentID) == 0 {
		return false, nil
	}
	tableName := fmt.Sprintf("%s_%s", indexName, typeName)
	count, err := dbc.connection.Model().TableExpr(tableName).Where("id = ?", documentID).Count()
	if err != nil {
		return false, utils.NewDBQueryError(err.Error())
	}
	return count == 1, nil
}

// Update existing document in database
func (dbc *Client) UpdateDocument(indexName, typeName, document string, documentID string) (result *ElasticSearchDocument, err error) {
	documentExist, err := dbc.IsDocumentExists(indexName, typeName, documentID)
	if err != nil {
		return nil, err
	}
	if len(documentID) != 0 && documentExist {
		tableName := fmt.Sprintf("%s_%s", indexName, typeName)
		_, err := dbc.connection.Model().TableExpr(tableName).Set("document = ?", document).Set("version = version + 1").Where("id = ?", documentID).Update()
		if err != nil {
			return nil, utils.NewDBQueryError(err.Error())
		}
		result, err = dbc.GetDocument(indexName, typeName, documentID)
	} else {
		return nil, utils.NewDBQueryError(fmt.Sprintf("Document with ID %s doesn't exists", documentID))
	}
	return result, nil
}

// Delete existing document in database
func (dbc *Client) DeleteDocument(indexName, typeName string, documentID string) (*ElasticSearchDocument, error) {
	documentObject, err := dbc.GetDocument(indexName, typeName, documentID)
	if err != nil {
		return nil, err
	}
	if documentObject != nil {
		tableName := fmt.Sprintf("%s_%s", indexName, typeName)
		_, err := dbc.connection.Model().TableExpr(tableName).Where("id = ?", documentID).Delete()
		if err != nil {
			return nil, utils.NewDBQueryError(err.Error())
		}
	}
	return documentObject, nil
}

/*
 * Documents processing helpers/internal methods
 */

// Create a document storage table for specified index and type
func (dbc *Client) createDataTable(indexName, typeName string) error {
	queryString := fmt.Sprintf("CREATE SEQUENCE %s_%s_id_seq;", indexName, typeName)
	_, err := dbc.connection.Exec(queryString)
	if err != nil {
		return utils.NewDBQueryError(err.Error())
	}
	queryString = fmt.Sprintf("CREATE TABLE %s_%s(id VARCHAR(128) PRIMARY KEY DEFAULT nextval('%s_%s_id_seq'), document JSONB NOT NULL, version integer);", indexName, typeName, indexName, typeName)
	_, err = dbc.connection.Exec(queryString)
	if err != nil {
		return utils.NewDBQueryError(err.Error())
	}
	return nil
}

// Insert a new document with default ID
func (dbc *Client) insertDocument(indexName, typeName, document string) (*ElasticSearchDocument, error) {
	documentObject := &ElasticSearchDocument{Document: document, Version: 1}
	queryString := fmt.Sprintf("INSERT INTO %s_%s (id, document, version) VALUES(DEFAULT, '%s', %d) RETURNING id;", indexName, typeName, document, 1)
	_, err := dbc.connection.Query(documentObject, queryString)
	if err != nil {
		return nil, utils.NewDBQueryError(err.Error())
	}
	return documentObject, nil
}

// Insert a new document with specified ID
func (dbc *Client) insertDocumentID(indexName, typeName, document string, documentID string) (*ElasticSearchDocument, error) {
	documentObject := &ElasticSearchDocument{documentID, document, 1}
	queryString := fmt.Sprintf("INSERT INTO %s_%s (id, document, version) VALUES(%d, '%s', %d);", indexName, typeName, documentID, document, 1)
	_, err := dbc.connection.Exec(queryString)
	if err != nil {
		return nil, utils.NewDBQueryError(err.Error())
	}
	return documentObject, nil
}

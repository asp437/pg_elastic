package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"github.com/asp437/pg_elastic/db"
	"github.com/asp437/pg_elastic/utils"
	"gopkg.in/olivere/elastic.v5"
	"log"
)

func migrateMapping(postgresqlClient *db.Client, indexName string, typeName string, mapping map[string]interface{}) {
	options, err := json.MarshalIndent(mapping, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Processing mapping for Index: %s, Type: %s\n", indexName, typeName)
	_, err = postgresqlClient.CreateType(indexName, typeName, string(options))
	if err != nil {
		log.Fatal(err)
	}
}

func migrateIndexMappings(elasticClient *elastic.Client, postgresqlClient *db.Client, indexName string) {
	mappings, err := elasticClient.GetMapping().Index(indexName).Do(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	for key, mapping := range mappings[indexName].(map[string]interface{})["mappings"].(map[string]interface{}) {
		processedMapping := map[string]interface{}{
			"mappings": map[string]interface{}{
				key: mapping,
			},
		}
		migrateMapping(postgresqlClient, indexName, key, processedMapping)
	}
}

func migrateIndex(elasticClient *elastic.Client, postgresqlClient *db.Client, indexName string) {
	getResult, err := elasticClient.Search().Index(indexName).Do(context.Background())
	if err != nil {
		log.Fatal(err)
	}

	for _, item := range getResult.Hits.Hits {
		body, err := json.MarshalIndent(item.Source, "", "\t")
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Processing record Index: %s, Type: %s, ID: %s\n", item.Index, item.Type, item.Id)
		_, err = postgresqlClient.CreateDocument(item.Index, item.Type, string(body), item.Id)
		if err != nil {
			log.Fatal(err)
		}
	}
}

func main() {
	postgresqlHost := flag.String("postgresql-host", "localhost:5432", "Host and port of destanation PostgreSQL instance")
	postgresqlUser := flag.String("postgresql-user", "postgres", "User to perform actions on destanation PostgreSQL instance")
	postgresqlPassword := flag.String("postgresql-password", "", "Password for specified user on destanation PostgreSQL instance")
	postgresqlDB := flag.String("postgresql-database", "postgres", "Name of destanation database inside target PostgreSQL instance")
	elasticsearchHost := flag.String("elasticsearch-host", "localhost:9200", "Host and port of source ElasticSearch instance")
	flag.Parse()

	elasticURL := fmt.Sprintf("http://%s", *elasticsearchHost)

	elasticClient, err := elastic.NewClient(elastic.SetURL(elasticURL))
	if err != nil {
		log.Fatal(err)
	}

	postgresConfig := utils.PostgresConnectionConfig{
		ServerAddress: *postgresqlHost,
		User:          *postgresqlUser,
		Password:      *postgresqlPassword,
		DBName:        *postgresqlDB,
	}
	postgresqlClient := db.CreateClient(postgresConfig)
	if postgresqlClient == nil {
		log.Fatal(errors.New("Database connection is not established"))
	}
	err = postgresqlClient.InitializeSchema()
	if err != nil {
		log.Fatal(err)
	}

	indices, err := elasticClient.IndexNames()
	if err != nil {
		log.Fatal(err)
	}

	for _, indexName := range indices {
		migrateIndexMappings(elasticClient, postgresqlClient, indexName)
		migrateIndex(elasticClient, postgresqlClient, indexName)
	}
}

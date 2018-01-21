package utils

import (
	"encoding/json"
	"log"
	"os"
)

// PGElasticConfig represents structure of the pg-elastic configuration
type PGElasticConfig struct {
	ServerPort     int
	PostgresConfig PostgresConnectionConfig
}

// PostgresConnectionConfig represents structure of the pg-elastic Postgres connection configuration
type PostgresConnectionConfig struct {
	ServerAddress string
	User          string
	Password      string
	DBName        string
}

// ReadConfig reads a configuration from the file at path
func ReadConfig(path string) *PGElasticConfig {
	file, _ := os.Open(path)
	decoder := json.NewDecoder(file)
	config := &PGElasticConfig{}
	err := decoder.Decode(config)

	if err != nil {
		log.Fatal(err)
	}

	return config
}

package utils

import (
	"encoding/json"
	"log"
	"os"
)

type PGElasticConfig struct {
	ServerPort     int
	PostgresConfig PostgresConnectionConfig
}

type PostgresConnectionConfig struct {
	ServerAddress string
	User          string
	Password      string
	DBName        string
}

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

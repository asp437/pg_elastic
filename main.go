package main

import (
	"github.com/asp437/pg_elastic/internal/server"
	"log"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Llongfile)
	s, err := server.InitializeServer("pg_elastic_config.json")

	if err != nil {
		log.Fatal(err)
	}

	s.Start()
}

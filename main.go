package main

import (
	"gitlab.com/lbauthdata/server"
)

func main() {

	config := &server.Config{
		Upstreamurl:        "http://localhost:6060",
		ExposedPort:        ":9001",
		PostgresConfig:     "user=keycloak password=password host=172.10.4.6 port=5432 database=lbauth sslmode=disable",
		EnableJSONLogging:  false,
		DisableAllLogging:  false,
		Verbose:            false,
		Opaurl:             "http://localhost:8181/v1/data/authzdata",
		HttpCallTimeoutSec: 10}

	lbdataauthz, _ := server.NewLbDataAuthzProxy(config)
	lbdataauthz.CreateDbConnection()
	lbdataauthz.RunServer()
}

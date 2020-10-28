package server

import (
	"testing"
)

func TestNewLbDataAuthzProxy(t *testing.T) {

	config := &Config{
		Upstreamurl:        "http://localhost:6060",
		ExposedPort:        ":9001",
		PostgresConfig:     "user=kk password=psw host=172.10.4.6 port=5432 database=lbauth sslmode=disable",
		EnableJSONLogging:  false,
		DisableAllLogging:  false,
		Verbose:            false,
		Opaurl:             "http://localhost:8181/v1/data/authzdata",
		HttpCallTimeoutSec: 10}

	_, err := NewLbDataAuthzProxy(config)

	if err != nil {
		t.Fatal(err)
	}
}

func TestCreateDbConnection(t *testing.T) {

}

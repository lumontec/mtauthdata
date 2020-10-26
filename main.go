package main

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
	"lbauthdata/server"
)

func main() {

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	upstreamurl := os.Getenv("UPSTREAMURL")
	exposedport := os.Getenv("EXPOSEDPORT")
	postgresconfig := os.Getenv("POSTGRESCONFIG")
	enablejsonlogging, _ := strconv.ParseBool(os.Getenv("ENABLEJSONLOGGING"))
	disablealllogging, _ := strconv.ParseBool(os.Getenv("DISABLEALLLOGGING"))
	verbose, _ := strconv.ParseBool(os.Getenv("VERBOSE"))
	opaurl := os.Getenv("OPARURL")
	httpcalltimeoutsec, _ := strconv.ParseInt(os.Getenv("HTTPCALLTIIMEOUTSEC"), 10, 0)

	log.Println("ACTIVE ENVS:", "\n",
		"upstreamurl:", upstreamurl, "\n",
		"exposedport:", exposedport, "\n",
		"postgresconfig:", postgresconfig, "\n",
		"enablejsonlogging:", enablejsonlogging, "\n",
		"disablealllogging:", disablealllogging, "\n",
		"verbose:", verbose, "\n",
		"opaurl:", opaurl, "\n",
		"httpcalltimeoutsec:", httpcalltimeoutsec)

	config := &server.Config{
		Upstreamurl:        upstreamurl,    //"http://localhost:6060",
		ExposedPort:        exposedport,    //":9001",
		PostgresConfig:     postgresconfig, //"user=keycloak password=password host=172.10.4.6 port=5432 database=lbauth sslmode=disable",
		EnableJSONLogging:  enablejsonlogging,
		DisableAllLogging:  disablealllogging,
		Verbose:            verbose,
		Opaurl:             opaurl, //"http://localhost:8181/v1/data/authzdata",
		HttpCallTimeoutSec: httpcalltimeoutsec}

	lbdataauthz, _ := server.NewLbDataAuthzProxy(config)
	lbdataauthz.CreateDbConnection()
	lbdataauthz.RunServer()
}

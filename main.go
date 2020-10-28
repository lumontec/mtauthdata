package main

import (
	"log"
	"os"
	"strconv"

	"lbauthdata/permissions"
	"lbauthdata/server"

	"github.com/joho/godotenv"
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
		Upstreamurl:        upstreamurl,
		ExposedPort:        exposedport,
		PostgresConfig:     postgresconfig,
		EnableJSONLogging:  enablejsonlogging,
		DisableAllLogging:  disablealllogging,
		Verbose:            verbose,
		Opaurl:             opaurl,
		HttpCallTimeoutSec: httpcalltimeoutsec}

	proxy, err := server.NewLbDataAuthzProxy(config)
	if err != nil {
		panic(err)
	}

	proxy.Permissions, err = permissions.NewPermissionProvider(postgresconfig)
	if err != nil {
		panic(err)
	}

	proxy.RunServer()
}

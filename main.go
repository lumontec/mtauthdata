package main

import (
	"log"
	"os"
	"strconv"

	"lbauthdata/authz"
	"lbauthdata/logger"
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
	opaurl := os.Getenv("OPARURL")
	levelmodules := os.Getenv("LOG_LEVELMODULES")
	enablejsonlogging, _ := strconv.ParseBool(os.Getenv("LOG_ENABLEJSONLOGGING"))
	disablealllogging, _ := strconv.ParseBool(os.Getenv("LOG_DISABLEALLLOGGING"))
	disablestacktrace, _ := strconv.ParseBool(os.Getenv("LOG_DISABLESTACKTRACE"))
	disablecaller, _ := strconv.ParseBool(os.Getenv("LOG_DISABLECALLER"))
	development, _ := strconv.ParseBool(os.Getenv("LOG_DEVELOPMENT"))
	httpcalltimeoutsec, _ := strconv.ParseInt(os.Getenv("HTTPCALLTIIMEOUTSEC"), 10, 0)

	log.Println("ACTIVE ENVS:", "\n",
		"upstreamurl:", upstreamurl, "\n",
		"exposedport:", exposedport, "\n",
		"postgresconfig:", postgresconfig, "\n",
		"opaurl:", opaurl, "\n",
		"levelmodules:", levelmodules, "\n",
		"enablejsonlogging:", enablejsonlogging, "\n",
		"disablealllogging:", disablealllogging, "\n",
		"disablestacktrace:", disablestacktrace, "\n",
		"disablecaller:", disablecaller, "\n",
		"development:", development, "\n",
		"httpcalltimeoutsec:", httpcalltimeoutsec)

	lconfig := &logger.LoggerConfig{
		LevelModules:      levelmodules,
		EnableJSONLogging: enablejsonlogging,
		DisableAllLogging: disablealllogging,
		DisableStackTrace: disablestacktrace,
		DisableCaller:     disablecaller,
		Development:       development,
	}

	sconfig := &server.Config{
		Upstreamurl:        upstreamurl,
		ExposedPort:        exposedport,
		PostgresConfig:     postgresconfig,
		Opaurl:             opaurl,
		HttpCallTimeoutSec: httpcalltimeoutsec}

	// Configure loggers
	logger.SetLoggerConfig(lconfig)

	// Create new server instance
	proxy, err := server.NewLbDataAuthzProxy(sconfig)
	if err != nil {
		panic(err)
	}

	// Initialize injectable permissionsprovider
	proxy.Permissions, err = permissions.NewDBPermissionProvider(postgresconfig)
	if err != nil {
		panic(err)
	}

	// Initialize injectable authzprovider
	proxy.Authz, err = authz.NewHttpAuthzProvider(sconfig)
	if err != nil {
		panic(err)
	}

	proxy.RunServer()
}

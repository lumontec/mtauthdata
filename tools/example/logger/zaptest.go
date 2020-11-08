package main

import "go.uber.org/zap"

const (
	disablelogging    = false
	disablestacktrace = false
	disablecaller     = true
	jsonlgging        = true
	development       = true
)

func newLogger() (*zap.Logger, error) {
	if disablelogging {
		return zap.NewNop(), nil
	}

	c := zap.NewProductionConfig()
	c.DisableStacktrace = disablestacktrace
	c.DisableCaller = disablecaller
	if !jsonlgging {
		c.Encoding = "console"
	}

	c.Development = development

	c.Level = zap.NewAtomicLevelAt(zap.DebugLevel)

	logger, _ := c.Build()
	return logger, nil
}

func main() {
	log, _ := newLogger()
	log.Debug("ciao", zap.String("chiave", "valore"))
}

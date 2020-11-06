package logger

import (
	"strings"

	"go.uber.org/zap"
)

type level int

const (
	DEBUG level = iota
	INFO
	ERROR
)

type LoggerConfig struct {
	LevelModules      string
	EnableJSONLogging bool
	DisableAllLogging bool
	DisableStackTrace bool
	DisableCaller     bool
	Development       bool
}

type ZapLogger struct {
	modulename string
	zlog       *zap.SugaredLogger
}

var modulelevels map[string]level
var jsonlgging bool
var disablelogging bool
var disablestacktrace bool
var disablecaller bool
var development bool

func SetLoggerConfig(lc *LoggerConfig) {

	jsonlgging = lc.EnableJSONLogging
	disablelogging = lc.DisableAllLogging
	disablestacktrace = lc.DisableStackTrace
	disablecaller = lc.DisableCaller
	development = lc.Development

	modLevArr := strings.Split(lc.LevelModules, ",")
	for _, modlev := range modLevArr {
		keyvalarr := strings.Split(modlev, "=")
		switch keyvalarr[1] {
		case "DEBUG":
			modulelevels[keyvalarr[0]] = DEBUG
			break
		case "INFO":
			modulelevels[keyvalarr[0]] = INFO
			break
		case "ERROR":
			modulelevels[keyvalarr[0]] = ERROR
			break
		}
	}
}

func newLogger(modulename string) (*zap.SugaredLogger, error) {
	if disablelogging {
		return zap.NewNop().Sugar(), nil
	}

	c := zap.NewProductionConfig()
	c.DisableStacktrace = disablestacktrace
	c.DisableCaller = disablecaller
	if !jsonlgging {
		c.Encoding = "console"
	}

	c.Development = development

	switch modulelevels[modulename] {
	case DEBUG:
		c.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		break

	case INFO:
		c.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
		break

	case ERROR:
		c.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
		break
	}

	logger, _ := c.Build()
	return logger.Sugar(), nil
}

func GetLogger(modulename string) *ZapLogger {
	zlogger, _ := newLogger(modulename)
	return &ZapLogger{
		modulename: modulename,
		zlog:       zlogger,
	}
}

func (zl *ZapLogger) Debug(args ...interface{}) {
	zl.zlog.Debug(zl.modulename, args)
}

func (zl *ZapLogger) Info(args ...interface{}) {
	zl.zlog.Info(zl.modulename, args)
}

func (zl *ZapLogger) Error(args ...interface{}) {
	zl.zlog.Error(zl.modulename, args)
}

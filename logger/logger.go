package logger

import (
	"lbauthdata/config"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type level int

const (
	DEBUG level = iota
	INFO
	ERROR
)

type ZapLogger struct {
	modulename string
	zlog       *zap.Logger
}

func KeyVal(key string, val string) zapcore.Field {
	return zap.String(key, val)
}

var modulelevels = map[string]level{}
var jsonlgging bool
var disablelogging bool
var disablestacktrace bool
var disablecaller bool
var development bool

func SetLoggerConfig(lc *config.LoggerConfig) {

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

func newLogger(modulename string) (*zap.Logger, error) {
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
	return logger, nil
}

func GetLogger(modulename string) *ZapLogger {
	zlogger, _ := newLogger(modulename)
	return &ZapLogger{
		modulename: modulename,
		zlog:       zlogger,
	}
}

func (zl *ZapLogger) Debug(msg string, ctx ...zapcore.Field) {
	zl.zlog.Debug(msg, ctx...)
}

func (zl *ZapLogger) Info(msg string, ctx ...zapcore.Field) {
	zl.zlog.Info(msg, ctx...)
}

func (zl *ZapLogger) Error(msg string, ctx ...zapcore.Field) {
	zl.zlog.Error(msg, ctx...)
}

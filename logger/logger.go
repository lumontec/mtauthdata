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

var moduleLevels map[string]level

func SetModulesLogLevels(levelsconfig string) {
	modLevArr := strings.Split(levelsconfig, ",")
	for _, modlev := range modLevArr {
		keyvalarr := strings.Split(modlev, "=")
		switch keyvalarr[1] {
		case "DEBUG":
			moduleLevels[keyvalarr[0]] = DEBUG
			break
		case "INFO":
			moduleLevels[keyvalarr[0]] = INFO
			break
		case "ERROR":
			moduleLevels[keyvalarr[0]] = ERROR
			break
		}
	}
}

type ZapLogger struct {
	modulename string
	zlog       *zap.SugaredLogger
}

func GetLogger(modulename string) *ZapLogger {
	logger, _ := zap.NewProduction()
	defer logger.Sync() // flushes buffer, if any
	sugar := logger.Sugar()
	return &ZapLogger{
		modulename: modulename,
		zlog:       sugar,
	}
}

func (zl *ZapLogger) Debug(args ...interface{}) {
	if moduleLevels[zl.modulename] <= DEBUG {
		zl.zlog.Debug(args)
	}
}

func (zl *ZapLogger) Info(args ...interface{}) {
	if moduleLevels[zl.modulename] <= INFO {
		zl.zlog.Info(args)
	}
}

func (zl *ZapLogger) Error(args ...interface{}) {
	if moduleLevels[zl.modulename] <= ERROR {
		zl.zlog.Error(args)
	}
}

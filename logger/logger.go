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

type Logger struct {
	modulename string
	zlog       *zap.SugaredLogger
}

func (l *Logger) ChildCathegory(cathegory string) *Logger {
	n := *l
	n.zlog = l.zlog.Named(cathegory)
	return &n
}

func (l *Logger) Debug(args ...interface{}) {
	l.zlog.Debug(args...)
}

func (l *Logger) Info(args ...interface{}) {
	l.zlog.Info(args...)
}

func (l *Logger) Error(args ...interface{}) {
	l.zlog.Error(args...)
}

func SetCtx(key string, val string) zapcore.Field {
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

func newZapLogger(rootname string) (*zap.SugaredLogger, error) {
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

	// switch modulelevels[modulename] {
	// case DEBUG:
	// 	c.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	// 	break

	// case INFO:
	// 	c.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	// 	break

	// case ERROR:
	// 	c.Level = zap.NewAtomicLevelAt(zap.ErrorLevel)
	// 	break
	// }

	logger, _ := c.Build()
	return logger.Named(rootname).Sugar(), nil
}

func NewRootLogger() *Logger {
	zlogger, _ := newZapLogger("root")
	return &Logger{
		modulename: "root",
		zlog:       zlogger,
	}
}

// func (zl *ZapLogger) Debug(msg string, ctx ...zapcore.Field) {
// 	zl.zlog.Debug(msg, ctx...)
// }

// func (zl *ZapLogger) Info(msg string, ctx ...zapcore.Field) {
// 	zl.zlog.Info(msg, ctx...)
// }

// func (zl *ZapLogger) Error(msg string, ctx ...zapcore.Field) {
// 	zl.zlog.Error(msg, ctx...)
// }

package interfaces

import (
	"lbauthdata/logger"
	"lbauthdata/model"
)

type PermissionProvider interface {
	GetGroupsPermissions(groupsarray []string, reqId string) (model.GroupPermMappings, error)
}

type AuthzProvider interface {
	GetAuthzDecision(groupmappings string, reqId string) (model.OpaResp, error)
}

type Logger interface {
	Debug(arg ...interface{})
	Info(arg ...interface{})
	Error(arg ...interface{})
	// NewRootLogger(arg ...interface{}) *logger.Logger
	ChildCathegory(cathegory string) *logger.Logger
	// SetLoggerConfig(config *config.LoggerConfig)
}

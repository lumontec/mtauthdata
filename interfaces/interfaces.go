package interfaces

import "lbauthdata/model"

type PermissionProvider interface {
	GetGroupsPermissions(groupsarray []string, reqId string) (model.GroupPermMappings, error)
}

type AuthzProvider interface {
	GetAuthzDecision(groupmappings string, reqId string) (model.OpaResp, error)
}

type Logger interface {
	Debug(arg string)
	Info(arg string)
	Error(arg string)
}

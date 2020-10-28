package interfaces

import "lbauthdata/model"

type PermissionProvider interface {
	GetGroupsPermissions(groupsarray []string) (model.GroupPermMappings, error)
}

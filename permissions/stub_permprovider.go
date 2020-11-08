package permissions

import "lbauthdata/model"

type StubPermissionProvider struct{}

func (sp *StubPermissionProvider) GetGroupsPermissions(groupsarray []string, reqId string) (model.GroupPermMappings, error) {
	// return testgroupmappings, nil
	return model.GroupPermMappings{
		Groups: []model.Mapping{
			{
				Group_uuid: "e694ddf2-1790-addd-0f57-bc23b9d47fa3",
				Permissions: model.Permissions{
					Admin_iots:     false,
					View_iots:      false,
					Configure_iots: false,
					Vpn_iots:       false,
					Webpage_iots:   false,
					Hmi_iots:       false,
					Data_admin:     false,
					Data_read:      false,
					Data_cold_read: false,
					Data_warm_read: false,
					Data_hot_read:  false,
					Services_admin: false,
					Billing_admin:  false,
					Org_admin:      false,
				},
			},
			{
				Group_uuid: "0dbd3c3e-0b44-4a4e-aa32-569f8951dc79",
				Permissions: model.Permissions{
					Admin_iots:     false,
					View_iots:      false,
					Configure_iots: false,
					Vpn_iots:       false,
					Webpage_iots:   false,
					Hmi_iots:       false,
					Data_admin:     false,
					Data_read:      false,
					Data_cold_read: false,
					Data_warm_read: false,
					Data_hot_read:  false,
					Services_admin: false,
					Billing_admin:  false,
					Org_admin:      false,
				},
			},
		},
	}, nil
}

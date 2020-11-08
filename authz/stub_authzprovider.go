package authz

import "lbauthdata/model"

type StubAuthzProvider struct{}

func (sa *StubAuthzProvider) GetAuthzDecision(groupmappings string, reqId string) (model.OpaResp, error) {
	return model.OpaResp{
		Result: model.OpaJudgement{
			Allow:          true,
			Allowed_groups: []string{"e694ddf2-1790-addd-0f57-bc23b9d47fa3", "0dbd3c3e-0b44-4a4e-aa32-569f8951dc79"},
			Read_allowed:   []string{"0dbd3c3e-0b44-4a4e-aa32-569f8951dc79"},
			Cold_allowed:   []string{"0dbd3c3e-0b44-4a4e-aa32-569f8951dc79"},
			Warm_allowed:   []string{},
			Hot_allowed:    []string{},
		},
	}, nil
}

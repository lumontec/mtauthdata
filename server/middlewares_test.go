package server

import (
	"context"
	"encoding/json"
	"fmt"
	"lbauthdata/model"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"

	"go.uber.org/zap"
)

type stubPermissionProvider struct{}

func (sp *stubPermissionProvider) GetGroupsPermissions(groupsarray []string) (model.GroupPermMappings, error) {
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

type stubAuthzProvider struct{}

func (sa *stubAuthzProvider) GetAuthzDecision(groupmappings string) (model.OpaResp, error) {
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

func newFakeConfig() *Config {
	return &Config{
		Upstreamurl:        "http://localhost:6060",
		ExposedPort:        ":9001",
		PostgresConfig:     "user=kk password=psw host=172.10.4.6 port=5432 database=lbauth sslmode=disable",
		EnableJSONLogging:  false,
		DisableAllLogging:  false,
		Verbose:            false,
		Opaurl:             "http://localhost:8181/v1/data/authzdata",
		HttpCallTimeoutSec: 10,
	}
}

func newFakeProxy(config *Config) (*lbDataAuthzProxy, error) {
	logger, err := createLogger(config)
	if err != nil {
		return nil, err
	}

	lbdataauthz := &lbDataAuthzProxy{
		config: config,
		logger: logger,
	}

	// Prepare remote url for request proxying
	lbdataauthz.upstream, err = url.Parse(config.Upstreamurl)
	if err != nil {
		return nil, err
	}

	logger.Info("initializing the service with:", zap.String("upstreamurl:", config.Upstreamurl), zap.String("action", "initializing proxy"))

	return lbdataauthz, nil
}

func TestGroupPermissionsMiddleware(t *testing.T) {

	req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
	res := httptest.NewRecorder()
	cfg := newFakeConfig()
	fl, _ := newFakeProxy(cfg)
	sp := &stubPermissionProvider{}
	fl.Permissions = sp

	permHandler := func(w http.ResponseWriter, r *http.Request) {
		gotpermstring, _ := r.Context().Value("groupmappings").(string)
		var gotpermissions model.GroupPermMappings
		if err := json.Unmarshal([]byte(gotpermstring), &gotpermissions); err != nil {
			panic(err)
		}

		wantpermissions, _ := sp.GetGroupsPermissions([]string{})
		if diff := deep.Equal(gotpermissions, wantpermissions); diff != nil {
			t.Error(diff)
		}
	}

	tim := fl.GroupPermissionsMiddleware(permHandler)
	tim.ServeHTTP(res, req)
}

func TestAuthzEnforcementMiddleware(t *testing.T) {

	cfg := newFakeConfig()
	fl, _ := newFakeProxy(cfg)
	sa := &stubAuthzProvider{}
	fl.Authz = sa
	sp := &stubPermissionProvider{}
	fl.Permissions = sp

	testpermissions, _ := sp.GetGroupsPermissions([]string{})

	testPermArrbytes, err := json.Marshal(testpermissions)

	if err != nil {
		// l.logger.Error("error unmarshalling groupsArrbytes:", zap.String("error:", err.Error()), zap.String("reqid:", reqId))
		panic(err)
	}

	req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
	req = req.WithContext(context.WithValue(req.Context(), "groupmappings", string(testPermArrbytes)))

	res := httptest.NewRecorder()

	authzHandler := func(w http.ResponseWriter, r *http.Request) {

		gottempstrings, _ := r.Context().Value("grouptemps").([]string)
		wanttempstrings := []string{"group:0dbd3c3e-0b44-4a4e-aa32-569f8951dc79:temp:read", "group:0dbd3c3e-0b44-4a4e-aa32-569f8951dc79:temp:cold"}

		if diff := deep.Equal(gottempstrings, wanttempstrings); diff != nil {
			t.Error(diff)
		}
	}

	tim := fl.AuthzEnforcementMiddleware(authzHandler)
	tim.ServeHTTP(res, req)
}

func TestTagsFilteringMiddleware(t *testing.T) {

	cfg := newFakeConfig()
	fl, _ := newFakeProxy(cfg)

	cases := []struct {
		inQuery   string
		gtemps    []string
		wantQuery string
	}{
		{"/tags/autoComplete/tags", []string{"group:0dbd3c3e-0b44-4a4e-aa32-569f8951dc79:temp:read", "group:0dbd3c3e-0b44-4a4e-aa32-569f8951dc79:temp:cold"}, "&expr=data:pr:ext:acl:grouptemp=~(^group:0dbd3c3e-0b44-4a4e-aa32-569f8951dc79:temp:read$|^group:0dbd3c3e-0b44-4a4e-aa32-569f8951dc79:temp:cold$)"},
		{"/tags/autoComplete/tags", []string{"group:0dbd3c3e-0b44-4a4e-aa32-569f8951dc79:temp:read"}, "&expr=data:pr:ext:acl:grouptemp=~(^group:0dbd3c3e-0b44-4a4e-aa32-569f8951dc79:temp:read$)"},
		{"/tags/autoComplete/tags", []string{}, "&expr=data:pr:ext:acl:grouptemp=~()"},
	}

	for _, tc := range cases {
		tname, _ := json.Marshal(tc.gtemps)
		t.Run(string(tname), func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "http://localhost"+tc.inQuery, nil)
			req = req.WithContext(context.WithValue(req.Context(), "grouptemps", tc.gtemps))
			res := httptest.NewRecorder()

			tagsFilterHandler := func(w http.ResponseWriter, r *http.Request) {
				fmt.Println("test response:", r.URL.RawQuery)
				assert.Equal(t, tc.wantQuery, r.URL.RawQuery)
			}

			tim := fl.TagsFilteringMiddleware(tagsFilterHandler)
			tim.ServeHTTP(res, req)
		})
	}
}

func TestRenderFilteringMiddleware(t *testing.T) {

	cfg := newFakeConfig()
	fl, _ := newFakeProxy(cfg)

	cases := []struct {
		inQuery  string
		gtemps   []string
		gotQuery string
	}{
		{"/render?target=demotags.iot1.metric0&from=-5min&until=now&format=json&maxDataPoints=653", []string{"group:0dbd3c3e-0b44-4a4e-aa32-569f8951dc79:temp:read", "group:0dbd3c3e-0b44-4a4e-aa32-569f8951dc79:temp:cold"}, "target=demotags.iot1.metric0&from=-5min&until=now&format=json&maxDataPoints=653&expr=data:pr:ext:acl:grouptemp=~(^group:0dbd3c3e-0b44-4a4e-aa32-569f8951dc79:temp:read$|^group:0dbd3c3e-0b44-4a4e-aa32-569f8951dc79:temp:cold$)"},
		{"/render?target=demotags.iot1.metric0&from=-5min&until=now&format=json&maxDataPoints=653", []string{"group:0dbd3c3e-0b44-4a4e-aa32-569f8951dc79:temp:cold"}, "target=demotags.iot1.metric0&from=-5min&until=now&format=json&maxDataPoints=653&expr=data:pr:ext:acl:grouptemp=~(^group:0dbd3c3e-0b44-4a4e-aa32-569f8951dc79:temp:cold$)"},
		{"/render?target=demotags.iot1.metric0&from=-5min&until=now&format=json&maxDataPoints=653", []string{}, "target=demotags.iot1.metric0&from=-5min&until=now&format=json&maxDataPoints=653&expr=data:pr:ext:acl:grouptemp=~()"},
	}

	for _, tc := range cases {
		tname, _ := json.Marshal(tc.gtemps)
		t.Run(string(tname), func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "http://localhost/"+tc.inQuery, nil)
			req = req.WithContext(context.WithValue(req.Context(), "grouptemps", tc.gtemps))
			res := httptest.NewRecorder()

			tagsFilterHandler := func(w http.ResponseWriter, r *http.Request) {
				assert.Equal(t, tc.gotQuery, r.URL.RawQuery)
			}

			tim := fl.TagsFilteringMiddleware(tagsFilterHandler)
			tim.ServeHTTP(res, req)
		})
	}
}

package server

import (
	"context"
	"encoding/json"
	"fmt"
	"lbauthdata/authz"
	"lbauthdata/config"
	"lbauthdata/model"
	"lbauthdata/permissions"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/go-test/deep"
	"github.com/stretchr/testify/assert"

	"go.uber.org/zap"
)

func newFakeConfig() *config.ServerConfig {
	return &config.ServerConfig{
		Upstreamurl:        "http://localhost:6060",
		ExposedPort:        ":9001",
		PostgresConfig:     "user=kk password=psw host=172.10.4.6 port=5432 database=lbauth sslmode=disable",
		Opaurl:             "http://localhost:8181/v1/data/authzdata",
		HttpCallTimeoutSec: 10,
	}
}

func NewFakeProxy(config *config.ServerConfig) (*lbDataAuthzProxy, error) {

	lbdataauthz := &lbDataAuthzProxy{
		config: config,
	}

	// Prepare remote url for request proxying
	upstream, err := url.Parse(config.Upstreamurl)
	if err != nil {
		return nil, err
	}

	lbdataauthz.upstream = upstream

	slog.Info("initializing the service with:", zap.String("upstreamurl:", config.Upstreamurl), zap.String("action", "initializing proxy"))

	return lbdataauthz, nil
}

func TestGroupPermissionsMiddleware(t *testing.T) {

	req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
	res := httptest.NewRecorder()
	cfg := newFakeConfig()
	fl, _ := NewFakeProxy(cfg)
	sp := &permissions.StubPermissionProvider{}
	fl.Permissions = sp

	permHandler := func(w http.ResponseWriter, r *http.Request) {
		gotpermstring, _ := r.Context().Value("groupmappings").(string)
		var gotpermissions model.GroupPermMappings
		if err := json.Unmarshal([]byte(gotpermstring), &gotpermissions); err != nil {
			panic(err)
		}

		wantpermissions, _ := sp.GetGroupsPermissions([]string{}, "testid")
		if diff := deep.Equal(gotpermissions, wantpermissions); diff != nil {
			t.Error(diff)
		}
	}

	tim := fl.GroupPermissionsMiddleware(permHandler)
	tim.ServeHTTP(res, req)
}

func TestAuthzEnforcementMiddleware(t *testing.T) {

	cfg := newFakeConfig()
	fl, _ := NewFakeProxy(cfg)
	sa := &authz.StubAuthzProvider{}
	fl.Authz = sa
	sp := &permissions.StubPermissionProvider{}
	fl.Permissions = sp

	testpermissions, _ := sp.GetGroupsPermissions([]string{}, "testid")

	testPermArrbytes, err := json.Marshal(testpermissions)

	if err != nil {
		slog.Error("error unmarshalling groupsArrbytes:", zap.String("error:", err.Error()))
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
	fl, _ := NewFakeProxy(cfg)

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
	fl, _ := NewFakeProxy(cfg)

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

package server

import (
	"fmt"
	"lbauthdata/model"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"go.uber.org/zap"
)

type fakePermissionProvider struct{}

func (fp *fakePermissionProvider) GetGroupsPermissions(groupsarray []string) (model.GroupPermMappings, error) {
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
	permHandler := func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("ciao")
	}

	req := httptest.NewRequest(http.MethodGet, "http://localhost", nil)
	// req = req.WithContext(context.WithValue(req.Context(), "some-key", "123ABC"))

	res := httptest.NewRecorder()

	permHandler(res, req)

	cfg := newFakeConfig()
	fl, _ := newFakeProxy(cfg)
	fp := &fakePermissionProvider{}
	fl.Permissions = fp

	tim := fl.GroupPermissionsMiddleware(permHandler)
	tim.ServeHTTP(res, req)
}

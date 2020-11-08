package authz

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"lbauthdata/config"
	"lbauthdata/logger"
	"lbauthdata/model"

	"go.uber.org/zap"
)

var log = logger.GetLogger("authz")

type AuthzClient struct {
	httpclient *http.Client
	opaurl     string
}

func NewHttpAuthzProvider(config *config.ServerConfig) (*AuthzClient, error) {
	log.Info("Creating database connection:", zap.String("dbconfig:", config.PostgresConfig))

	// Initializing http client
	httpclient := &http.Client{
		Timeout: time.Second * time.Duration(config.HttpCallTimeoutSec),
	}

	return &AuthzClient{
		httpclient: httpclient,
		opaurl:     config.Opaurl}, nil
}

func (ac *AuthzClient) GetAuthzDecision(stringgroupmappings string, reqId string) (model.OpaResp, error) {
	opaurl, err := url.Parse(ac.opaurl)
	if err != nil {
		return model.OpaResp{}, fmt.Errorf("could not validate opa url: %w", err)
	}

	req, err := http.NewRequest("POST", opaurl.String(), strings.NewReader(`{ "input" :`+stringgroupmappings+`}`))
	// req.Header.Set("X-Auth-Username", "admin")
	req.Header.Set("Content-Type", "application/json")
	// req.Header.Set("Accept", "application/json")
	resp, err := ac.httpclient.Do(req)
	if err != nil {
		return model.OpaResp{}, fmt.Errorf("opa http req failed: %w", err)
	}

	data, err := ioutil.ReadAll(resp.Body)
	log.Debug("OPA judgement:", zap.String("response:", string(data)), zap.String("reqid:", reqId))

	if err != nil {
		return model.OpaResp{}, fmt.Errorf("opa read resp failed: %w", err)
	}

	var opaResp model.OpaResp
	if err := json.Unmarshal(data, &opaResp); /*json.NewDecoder(resp.Body).Decode(&orgResp);*/ err != nil {
		return model.OpaResp{}, fmt.Errorf("opa resp unmarshal failed: %w", err)
	}

	return opaResp, nil
}

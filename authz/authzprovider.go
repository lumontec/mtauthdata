package authz

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"lbauthdata/logger"
	"lbauthdata/model"
	"lbauthdata/server"

	"go.uber.org/zap"
)

var log = logger.GetLogger("authz")

type AuthzClient struct {
	httpclient *http.Client
	opaurl     string
}

func NewHttpAuthzProvider(config *server.Config) (*AuthzClient, error) {
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
		log.Error("could not validate opa url:", zap.String("reqid:", reqId))
		panic(err)
	}

	req, err := http.NewRequest("POST", opaurl.String(), strings.NewReader(`{ "input" :`+stringgroupmappings+`}`))
	// req.Header.Set("X-Auth-Username", "admin")
	req.Header.Set("Content-Type", "application/json")
	// req.Header.Set("Accept", "application/json")
	resp, err := ac.httpclient.Do(req)
	if err != nil {
		log.Error("opa call failed:", zap.String("error:", err.Error()), zap.String("reqid:", reqId))
		panic(err)
	}

	data, err := ioutil.ReadAll(resp.Body)
	log.Info("OPA judgement:", zap.String("response:", string(data)), zap.String("reqid:", reqId))

	if err != nil {
		log.Error("opa call failed:", zap.String("error:", err.Error()), zap.String("reqid:", reqId))
		panic(err)
	}

	var opaResp model.OpaResp
	if err := json.Unmarshal(data, &opaResp); /*json.NewDecoder(resp.Body).Decode(&orgResp);*/ err != nil {
		log.Error("opa resp unmarshal failed:", zap.String("error:", err.Error()), zap.String("reqid:", reqId))
		panic(err)
	}

	return opaResp, nil
}

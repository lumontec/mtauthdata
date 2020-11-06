package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"

	"lbauthdata/expr"
	"lbauthdata/logger"
)

//--------------------
//	MIDDLEWARE CHAIN |
//--------------------
//
//  Httpreq -> GroupPermissionsMiddleware -> groupmappings
//									|
//									v
//  groupmappings -> AuthzEnforcementMiddleware -> grouptemps
//									|
//                              | /tags | _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _| /render |
//                                  |                                                 |
//									v                                                 v
//  grouptemps -> TagsFilteringMiddleware -> rawquery         grouptemps -> RenderFilteringMiddleware -> rawquery
//                                  |                                                 |
//                                  |_ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _ _|
//                                                       |
//                                                       v
//	                                    rawquery -> proxyhandler
//

var (
	ErrGmapExtract = errors.New("cannot extract groupmappings from ctx")
	ErrGtmpExtract = errors.New("cannot extract grouptemps from ctx")
)

var gplog = logger.GetLogger("middlewares.grouppermissions")
var azlog = logger.GetLogger("middlewares.authorization")
var tflog = logger.GetLogger("middlewares.tagsfilter")
var rflog = logger.GetLogger("middlewares.renderfilter")
var mtlog = logger.GetLogger("middlewares.mtproxy")

func (l *lbDataAuthzProxy) GroupPermissionsMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqId := middleware.GetReqID(r.Context())

		groupsarray := []string{
			"e694ddf2-1790-addd-0f57-bc23b9d47fa3",
			"0dbd3c3e-0b44-4a4e-aa32-569f8951dc79",
			"5033357b-25f3-0124-180c-51029be60114",
			"521db0c7-78e9-36b8-a95b-da4ba8fe7f9e"}

		gplog.Debug("gathering permissions from db for groups", zap.Strings("groups:", groupsarray), zap.String("reqid:", reqId))

		groupsArr, err := l.Permissions.GetGroupsPermissions(groupsarray, reqId)
		if err != nil {
			gplog.Error("error gathering permissions:", zap.String("error:", err.Error()), zap.String("reqid:", reqId))
			http.Error(w, http.StatusText(500), 500)
			return
		}

		groupsArrbytes, err := json.Marshal(groupsArr)
		if err != nil {
			gplog.Error("error unmarshalling groupsArrbytes:", zap.String("error:", err.Error()), zap.String("reqid:", reqId))
			http.Error(w, http.StatusText(500), 500)
			return
		}

		ctx := r.Context()
		ctx = context.WithValue(ctx, "groupmappings", string(groupsArrbytes))
		r = r.WithContext(ctx)

		// Will pass groupmappings inside the request context to enforcement
		// ctx := context.WithValue(r.Context(), groupsArrbytes, "groupmappings")

		h.ServeHTTP(w, r)
	})
}

func (l *lbDataAuthzProxy) AuthzEnforcementMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		reqId := middleware.GetReqID(r.Context())
		stringgroupmappings, ok := r.Context().Value("groupmappings").(string)

		if !ok {
			gplog.Error(zap.Error(ErrGmapExtract), zap.String("reqid:", reqId))
			http.Error(w, http.StatusText(400), 400)
			return
		}

		gplog.Debug("enforcing authorization for context:", zap.String("context:", stringgroupmappings), zap.String("reqid:", reqId), zap.String("opaurl:", l.config.Opaurl))

		opaResp, _ := l.Authz.GetAuthzDecision(stringgroupmappings, reqId)

		if opaResp.Result.Allow == false {
			gplog.Info("user is NOT ALLOWED to access data", zap.String("reqid:", reqId))
			http.Error(w, http.StatusText(400), 400)
			return
		} else {

			gplog.Debug("user is allowed to access data, will generate grouptemps", zap.String("reqid:", reqId))

			grouptemps := []string{}
			for _, group := range opaResp.Result.Read_allowed {
				grouptemps = append(grouptemps, "group:"+group+":temp:read")
			}
			for _, group := range opaResp.Result.Cold_allowed {
				grouptemps = append(grouptemps, "group:"+group+":temp:cold")
			}
			for _, group := range opaResp.Result.Warm_allowed {
				grouptemps = append(grouptemps, "group:"+group+":temp:warm")
			}
			for _, group := range opaResp.Result.Hot_allowed {
				grouptemps = append(grouptemps, "group:"+group+":temp:hot")
			}

			gplog.Debug("generated grouptemps", zap.Strings("grouptemps:", grouptemps), zap.String("reqid:", reqId))

			ctx := r.Context()
			ctx = context.WithValue(ctx, "grouptemps", grouptemps)
			r = r.WithContext(ctx)
		}

		h.ServeHTTP(w, r)
	})
}

func (l *lbDataAuthzProxy) TagsFilteringMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		reqId := middleware.GetReqID(r.Context())

		grouptemps, ok := r.Context().Value("grouptemps").([]string)

		if !ok {
			tflog.Error(zap.Error(ErrGtmpExtract), zap.String("reqid:", reqId))
			http.Error(w, http.StatusText(500), 500)
			return
		}

		// grouptemps := []string{"group:dom:e34ba21c74c289ba894b75ae6c76d22f:temp:warm", "group:ou:e34ba21c74c289ba894b75ae6c76d22f:temp:warm"}

		tflog.Debug("pre-filter request /tags:", zap.String("RawQuery:", r.URL.RawQuery), zap.String("reqid:", reqId))

		grouptempfilters := ""

		if len(grouptemps) > 0 {
			for _, grouptemp := range grouptemps {
				grouptempfilters += "^" + grouptemp + "$|"
			}
		}

		grouptempfilters = strings.TrimSuffix(grouptempfilters, "|")

		switch r.URL.Path {
		case "/tags/autoComplete/values":
			parts := strings.Split(r.URL.RawQuery, "&")
			if parts[0] != "tag=name" {
				subparts := strings.Split(parts[0], "=")
				r.URL.RawQuery = "tag=data:pu:int:cust:" + subparts[1] + "&" + parts[1]
			}
		case "/tags/autoComplete/tags":
			if r.URL.RawQuery != "" && r.URL.RawQuery != "tagPrefix=n" && r.URL.RawQuery != "tagPrefix=na" && r.URL.RawQuery != "tagPrefix=nam" && r.URL.RawQuery != "tagPrefix=name" {
				parts := strings.Split(r.URL.RawQuery, "=")
				r.URL.RawQuery = parts[0] + "=data:pu:int:cust:" + parts[1]
			}
		}

		r.URL.RawQuery += "&expr=data:pr:ext:acl:grouptemp=~(" + grouptempfilters + ")"

		tflog.Debug("filtered request /tags:", zap.String("RawQuery:", r.URL.RawQuery), zap.String("reqid:", reqId))
		h.ServeHTTP(w, r)
	})
}

func (l *lbDataAuthzProxy) RenderFilteringMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		reqId := middleware.GetReqID(r.Context())
		rflog.Debug("pre-filter request /render:", zap.String("RawQuery:", r.URL.RawQuery), zap.String("reqid:", reqId))

		grouptemps, ok := r.Context().Value("grouptemps").([]string)

		if !ok {
			rflog.Error(ErrGtmpExtract, zap.String("reqid:", reqId))
			http.Error(w, http.StatusText(500), 500)
			return
		}

		// grouptemps := []string{"group:dom:e34ba21c74c289ba894b75ae6c76d22f:temp:warm", "group:ou:e34ba21c74c289ba894b75ae6c76d22f:temp:warm"}

		grouptempfilters := ""

		if len(grouptemps) > 0 {
			for _, grouptemp := range grouptemps {
				grouptempfilters += "^" + grouptemp + "$|"
			}
		}

		urlParsed, err := url.ParseQuery(r.URL.RawQuery)
		if err != nil {
			rflog.Error("error parsing rawquery:", zap.String("error:", err.Error()), zap.String("reqid:", reqId))
			http.Error(w, http.StatusText(500), 500)
			return
		}

		exprs, err := expr.ParseMany(urlParsed["target"])
		if err != nil {
			rflog.Error("error parsing target:", zap.String("error:", err.Error()), zap.String("reqid:", reqId))
			http.Error(w, http.StatusText(500), 500)
			panic(err)
		}

		targetstr := ""

		for _, expr := range exprs {
			targetstr += expr.ApplyQueryFilters("\"data:pr:ext:acl:grouptemp=~(" + grouptempfilters + ")\"")
		}

		urlParsed.Del("target")            // Delete target key
		urlParsed.Add("target", targetstr) // Adds recomputed target

		r.URL.RawQuery = urlParsed.Encode()
		rflog.Debug("filtered request /render:", zap.String("RawQuery:", r.URL.RawQuery), zap.String("reqid:", reqId))

		h.ServeHTTP(w, r)
	})
}

func (l *lbDataAuthzProxy) ProxyHandler(w http.ResponseWriter, r *http.Request) {
	reqId := middleware.GetReqID(r.Context())
	mtlog.Debug("sending request to metrictank", zap.String("reqid:", reqId))
	l.reverseproxy.ServeHTTP(w, r)
}

package server

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/middleware"
	"github.com/jackc/pgtype"

	"gitlab.com/lbauthdata/expr"
	"gitlab.com/lbauthdata/model"
	"go.uber.org/zap"
)

func (l *lbDataAuthzProxy) GroupPermissionsMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqId := middleware.GetReqID(r.Context())

		groupsarray := []string{
			"e694ddf2-1790-addd-0f57-bc23b9d47fa3",
			"0dbd3c3e-0b44-4a4e-aa32-569f8951dc79",
			"5033357b-25f3-0124-180c-51029be60114",
			"521db0c7-78e9-36b8-a95b-da4ba8fe7f9e"}

		l.logger.Info("gathering permissions from db for groups", zap.Strings("groups:", groupsarray), zap.String("reqid:", reqId))

		groupsquery := ""

		// Generating query string
		for index, group := range groupsarray {
			groupsquery += " group_uuid = '" + group + "'"
			if index < (len(groupsarray) - 1) {
				groupsquery += " OR"
			}
		}

		l.logger.Info("groupsquery", zap.String("query:", groupsquery))

		// var id int64
		// var group_uuid pgtype.UUID
		// var role_uuid pgtype.UUID
		// Send the query to the server. The returned rows MUST be closed
		// before conn can be used again.
		rows, err := l.dbconn.Query(
			`SELECT
			roles_group_mapping.group_uuid,
			bool_or (roles.admin_iots) AS admin_iots,
			bool_or (roles.view_iots) AS view_iots,
			bool_or (roles.configure_iots) AS configure_iots,
			bool_or (roles.vpn_iots) AS vpn_iots,
			bool_or (roles.webpage_iots) AS webpage_iots,
			bool_or (roles.hmi_iots) AS hmi_iots,
			bool_or (roles.data_admin) AS data_admin,
			bool_or (roles.data_read) AS data_read,
			bool_or (roles.data_cold_read) AS data_cold_read,
			bool_or (roles.data_warm_read) AS data_warm_read,
			bool_or (roles.data_hot_read) AS data_hot_read,
			bool_or (roles.services_admin) AS services_admin,
			bool_or (roles.billing_admin) AS billing_admin,
			bool_or (roles.org_admin) AS org_admin
		FROM	roles_group_mapping
		INNER JOIN roles ON roles_group_mapping.role_uuid = roles.uuid AND (` + groupsquery + `) GROUP BY roles_group_mapping.group_uuid;`)
		if err != nil {
			l.logger.Error("during db query:", zap.String("error:", err.Error()), zap.String("reqid:", reqId))
			panic(err)
		}
		// rows.Close is called by rows.Next when all rows are read
		// or an error occurs in Next or Scan. So it may optionally be
		// omitted if nothing in the rows.Next loop can panic. It is
		// safe to close rows multiple times.
		defer rows.Close()

		// var sum int32

		groupsArr := model.GroupPermMappings{}

		// Iterate through the result set
		for rows.Next() {

			groupMap := model.Mapping{}

			var group_uuid pgtype.UUID
			var admin_iots pgtype.Bool
			var view_iots pgtype.Bool
			var configure_iots pgtype.Bool
			var vpn_iots pgtype.Bool
			var webpage_iots pgtype.Bool
			var hmi_iots pgtype.Bool
			var data_admin pgtype.Bool
			var data_read pgtype.Bool
			var data_cold_read pgtype.Bool
			var data_warm_read pgtype.Bool
			var data_hot_read pgtype.Bool
			var services_admin pgtype.Bool
			var billing_admin pgtype.Bool
			var org_admin pgtype.Bool

			err = rows.Scan(
				&group_uuid,
				&admin_iots,
				&view_iots,
				&configure_iots,
				&vpn_iots,
				&webpage_iots,
				&hmi_iots,
				&data_admin,
				&data_read,
				&data_cold_read,
				&data_warm_read,
				&data_hot_read,
				&services_admin,
				&billing_admin,
				&org_admin)

			if err != nil {
				l.logger.Error("error scanning rows:", zap.String("error:", err.Error()), zap.String("reqid:", reqId))
				panic(err)
			}

			group_uuid.AssignTo(&groupMap.Group_uuid)
			admin_iots.AssignTo(&groupMap.Permissions.Admin_iots)
			view_iots.AssignTo(&groupMap.Permissions.View_iots)
			configure_iots.AssignTo(&groupMap.Permissions.Configure_iots)
			vpn_iots.AssignTo(&groupMap.Permissions.Vpn_iots)
			webpage_iots.AssignTo(&groupMap.Permissions.Webpage_iots)
			hmi_iots.AssignTo(&groupMap.Permissions.Hmi_iots)
			data_admin.AssignTo(&groupMap.Permissions.Data_admin)
			data_read.AssignTo(&groupMap.Permissions.Data_read)
			data_cold_read.AssignTo(&groupMap.Permissions.Data_cold_read)
			data_warm_read.AssignTo(&groupMap.Permissions.Data_warm_read)
			data_hot_read.AssignTo(&groupMap.Permissions.Data_hot_read)
			services_admin.AssignTo(&groupMap.Permissions.Services_admin)
			billing_admin.AssignTo(&groupMap.Permissions.Billing_admin)
			org_admin.AssignTo(&groupMap.Permissions.Org_admin)

			groupsArr.Groups = append(groupsArr.Groups, groupMap)
			// sum += id
		}

		groupsArrbytes, err := json.Marshal(groupsArr)
		if err != nil {
			l.logger.Error("error unmarshalling groupsArrbytes:", zap.String("error:", err.Error()), zap.String("reqid:", reqId))
			panic(err)
		}

		l.logger.Info("permissions were retreived for groups", zap.String("reqid:", reqId))

		// Any errors encountered by rows.Next or rows.Scan will be returned here
		if rows.Err() != nil {
			l.logger.Error("error during rows next:", zap.String("error:", rows.Err().Error()), zap.String("reqid:", reqId))
			panic(rows.Err())
		}

		// Take the context out from the request
		ctx := r.Context()

		// Get new context with key-value "params" -> "httprouter.Params"
		ctx = context.WithValue(ctx, "groupmappings", string(groupsArrbytes))

		// Get new http.Request with the new context
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
			err := errors.New("could not extract value groupmappings from context")
			l.logger.Error("could not extract value from context:", zap.String("reqid:", reqId))
			panic(err)
		}

		l.logger.Info("enforcing authorization for context:", zap.String("context:", stringgroupmappings), zap.String("reqid:", reqId), zap.String("opaurl:", l.config.Opaurl))

		opaurl, err := url.Parse(l.config.Opaurl)
		if err != nil {
			l.logger.Error("could not validate opa url:", zap.String("reqid:", reqId))
			panic(err)
		}

		req, err := http.NewRequest("POST", opaurl.String(), strings.NewReader(`{ "input" :`+stringgroupmappings+`}`))
		// req.Header.Set("X-Auth-Username", "admin")
		req.Header.Set("Content-Type", "application/json")
		// req.Header.Set("Accept", "application/json")
		resp, err := l.httpclient.Do(req)
		if err != nil {
			l.logger.Error("opa call failed:", zap.String("error:", err.Error()), zap.String("reqid:", reqId))
			panic(err)
		}

		data, err := ioutil.ReadAll(resp.Body)
		l.logger.Info("OPA judgement:", zap.String("response:", string(data)), zap.String("reqid:", reqId))

		if err != nil {
			l.logger.Error("opa call failed:", zap.String("error:", err.Error()), zap.String("reqid:", reqId))
			panic(err)
		}

		var opaResp model.OpaResp
		if err := json.Unmarshal(data, &opaResp); /*json.NewDecoder(resp.Body).Decode(&orgResp);*/ err != nil {
			l.logger.Error("opa resp unmarshal failed:", zap.String("error:", err.Error()), zap.String("reqid:", reqId))
			panic(err)
		}

		if opaResp.Result.Allow == false {
			l.logger.Info("user is NOT ALLOWED to access data", zap.String("reqid:", reqId))
			http.Error(w, http.StatusText(400), 400)
			return

		} else {

			l.logger.Info("user is allowed to access data, will generate grouptemps", zap.String("reqid:", reqId))

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

			l.logger.Info("generated grouptemps", zap.Strings("grouptemps:", grouptemps), zap.String("reqid:", reqId))

			// Take the context out from the request
			ctx := r.Context()

			// Get new context with key-value "params" -> "httprouter.Params"
			ctx = context.WithValue(ctx, "grouptemps", grouptemps)

			// Get new http.Request with the new context
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
			err := errors.New("could not extract value grouptemps from context")
			l.logger.Error("could not extract value from context:", zap.String("reqid:", reqId))
			panic(err)
		}

		// grouptemps := []string{"group:dom:e34ba21c74c289ba894b75ae6c76d22f:temp:warm", "group:ou:e34ba21c74c289ba894b75ae6c76d22f:temp:warm"}

		l.logger.Info("pre-filter request /tags:", zap.String("RawQuery:", r.URL.RawQuery), zap.String("reqid:", reqId))

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

		l.logger.Info("filtered request /tags:", zap.String("RawQuery:", r.URL.RawQuery), zap.String("reqid:", reqId))
		h.ServeHTTP(w, r)
	})
}

func (l *lbDataAuthzProxy) RenderFilteringMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		reqId := middleware.GetReqID(r.Context())
		l.logger.Info("pre-filter request /render:", zap.String("RawQuery:", r.URL.RawQuery), zap.String("reqid:", reqId))

		grouptemps, ok := r.Context().Value("grouptemps").([]string)

		if !ok {
			err := errors.New("could not extract value grouptemps from context")
			l.logger.Error("could not extract value from context:", zap.String("reqid:", reqId))
			panic(err)
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
			panic(err)
		}

		exprs, err := expr.ParseMany(urlParsed["target"])

		if err != nil {
			panic(err)
		}

		targetstr := ""

		for _, expr := range exprs {
			targetstr += expr.ApplyQueryFilters("\"data:pr:ext:acl:grouptemp=~(" + grouptempfilters + ")\"")
		}

		urlParsed.Del("target")            // Delete target key
		urlParsed.Add("target", targetstr) // Adds recomputed target

		r.URL.RawQuery = urlParsed.Encode()
		l.logger.Info("filtered request /render:", zap.String("RawQuery:", r.URL.RawQuery), zap.String("reqid:", reqId))

		h.ServeHTTP(w, r)
	})
}

func (l *lbDataAuthzProxy) ProxyHandler(w http.ResponseWriter, r *http.Request) {
	reqId := middleware.GetReqID(r.Context())
	l.logger.Info("sending request to metrictank", zap.String("reqid:", reqId))
	l.reverseproxy.ServeHTTP(w, r)
}

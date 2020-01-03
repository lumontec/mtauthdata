package server

import (
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/go-chi/chi/middleware"

	"gitlab.com/lbauthdata/expr"
	"go.uber.org/zap"
)

//func (l *lbDataAuthzProxy) LogMiddleware(h http.HandlerFunc) http.HandlerFunc {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		//		l.logger.Info("LOGGING INPUT:", zap.Reflect("request:", r))
//		log.Println("request:", r)
//		h.ServeHTTP(w, r)
//	})
//}

//func GroupPermissionsMiddleware(h http.HandlerFunc) http.HandlerFunc {
//	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
//		log.Println("GATHERING PERMSSIONS FROM USER GROUPS:")
//
//		// initialize db connection
//		conn, err := pgx.Connect(context.Background(), "user=keycloak password=password host=172.10.4.6 port=5432 database=lbauth sslmode=disable")
//		if err != nil {
//			fmt.Fprintf(os.Stderr, "Unable to connection to database: %v\n", err)
//			os.Exit(1)
//		}
//		defer conn.Close(context.Background())
//
//		// var id int64
//		// var group_uuid pgtype.UUID
//		// var role_uuid pgtype.UUID
//		// Send the query to the server. The returned rows MUST be closed
//		// before conn can be used again.
//		rows, err := conn.Query(context.Background(),
//			`SELECT
//			roles_group_mapping.group_uuid,
//			bool_or (roles.admin_iots) AS admin_iots,
//			bool_or (roles.view_iots) AS view_iots,
//			bool_or (roles.configure_iots) AS configure_iots,
//			bool_or (roles.vpn_iots) AS vpn_iots,
//			bool_or (roles.webpage_iots) AS webpage_iots,
//			bool_or (roles.hmi_iots) AS hmi_iots,
//			bool_or (roles.data_admin) AS data_admin,
//			bool_or (roles.data_read) AS data_read,
//			bool_or (roles.data_cold_read) AS data_cold_read,
//			bool_or (roles.data_warm_read) AS data_warm_read,
//			bool_or (roles.data_hot_read) AS data_hot_read,
//			bool_or (roles.services_admin) AS services_admin,
//			bool_or (roles.billing_admin) AS billing_admin,
//			bool_or (roles.org_admin) AS org_admin
//		FROM	roles_group_mapping
//		INNER JOIN roles ON roles_group_mapping.role_uuid = roles.uuid AND (
//			group_uuid = 'e694ddf2-1790-addd-0f57-bc23b9d47fa3' OR
//			group_uuid = '0dbd3c3e-0b44-4a4e-aa32-569f8951dc79' OR
//			group_uuid = '5033357b-25f3-0124-180c-51029be60114' OR
//			group_uuid = '521db0c7-78e9-36b8-a95b-da4ba8fe7f9e' )
//		GROUP BY roles_group_mapping.group_uuid;`)
//		if err != nil {
//			fmt.Println(err)
//		}
//		// rows.Close is called by rows.Next when all rows are read
//		// or an error occurs in Next or Scan. So it may optionally be
//		// omitted if nothing in the rows.Next loop can panic. It is
//		// safe to close rows multiple times.
//		defer rows.Close()
//
//		// var sum int32
//
//		groupsArr := model.GroupPermMappings{}
//
//		// Iterate through the result set
//		for rows.Next() {
//
//			groupMap := model.Mapping{}
//
//			var group_uuid pgtype.UUID
//			var admin_iots pgtype.Bool
//			var view_iots pgtype.Bool
//			var configure_iots pgtype.Bool
//			var vpn_iots pgtype.Bool
//			var webpage_iots pgtype.Bool
//			var hmi_iots pgtype.Bool
//			var data_admin pgtype.Bool
//			var data_read pgtype.Bool
//			var data_cold_read pgtype.Bool
//			var data_warm_read pgtype.Bool
//			var data_hot_read pgtype.Bool
//			var services_admin pgtype.Bool
//			var billing_admin pgtype.Bool
//			var org_admin pgtype.Bool
//
//			err = rows.Scan(
//				&group_uuid,
//				&admin_iots,
//				&view_iots,
//				&configure_iots,
//				&vpn_iots,
//				&webpage_iots,
//				&hmi_iots,
//				&data_admin,
//				&data_read,
//				&data_cold_read,
//				&data_warm_read,
//				&data_hot_read,
//				&services_admin,
//				&billing_admin,
//				&org_admin)
//
//			if err != nil {
//				fmt.Println(err)
//			}
//
//			group_uuid.AssignTo(&groupMap.Group_uuid)
//			admin_iots.AssignTo(&groupMap.Permissions.Admin_iots)
//			view_iots.AssignTo(&groupMap.Permissions.View_iots)
//			configure_iots.AssignTo(&groupMap.Permissions.Configure_iots)
//			vpn_iots.AssignTo(&groupMap.Permissions.Vpn_iots)
//			webpage_iots.AssignTo(&groupMap.Permissions.Webpage_iots)
//			hmi_iots.AssignTo(&groupMap.Permissions.Hmi_iots)
//			data_admin.AssignTo(&groupMap.Permissions.Data_admin)
//			data_read.AssignTo(&groupMap.Permissions.Data_read)
//			data_cold_read.AssignTo(&groupMap.Permissions.Data_cold_read)
//			data_warm_read.AssignTo(&groupMap.Permissions.Data_warm_read)
//			data_hot_read.AssignTo(&groupMap.Permissions.Data_hot_read)
//			services_admin.AssignTo(&groupMap.Permissions.Services_admin)
//			billing_admin.AssignTo(&groupMap.Permissions.Billing_admin)
//			org_admin.AssignTo(&groupMap.Permissions.Org_admin)
//
//			groupsArr.Groups = append(groupsArr.Groups, groupMap)
//			// sum += id
//		}
//
//		groupsArrbytes, err := json.Marshal(groupsArr)
//		if err != nil {
//			panic(err)
//		}
//
//		fmt.Println(string(groupsArrbytes))
//
//		// Any errors encountered by rows.Next or rows.Scan will be returned here
//		if rows.Err() != nil {
//			fmt.Println(err)
//		}
//
//		log.Println(r)
//		h.ServeHTTP(w, r)
//	})
//}

func (l *lbDataAuthzProxy) TagsFilteringMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		reqid := middleware.GetReqID(r.Context())
		log.Println(r.Context())
		l.logger.Info("pre-filter request /tags:", zap.String("RawQuery:", r.URL.RawQuery), zap.String("reqid:", reqId))

		grouptemps := []string{"group:dom:e34ba21c74c289ba894b75ae6c76d22f:temp:warm", "group:ou:e34ba21c74c289ba894b75ae6c76d22f:temp:warm"}

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

		l.logger.Info("filtered request /tags:", zap.String("RawQuery:", r.URL.RawQuery))
		h.ServeHTTP(w, r)
	})
}

func (l *lbDataAuthzProxy) RenderFilteringMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		l.logger.Info("pre-filter request /render:", zap.String("RawQuery:", r.URL.RawQuery))

		grouptemps := []string{"group:dom:e34ba21c74c289ba894b75ae6c76d22f:temp:warm", "group:ou:e34ba21c74c289ba894b75ae6c76d22f:temp:warm"}

		grouptempfilters := ""

		if len(grouptemps) > 0 {
			for _, grouptemp := range grouptemps {
				grouptempfilters += "^" + grouptemp + "$|"
			}
		}

		urlParsed, err := url.ParseQuery(r.URL.RawQuery)

		if err != nil {
			log.Fatal(err)
		}

		exprs, err := expr.ParseMany(urlParsed["target"])

		if err != nil {
			log.Fatal(err)
		}

		targetstr := ""

		for _, expr := range exprs {
			targetstr += expr.ApplyQueryFilters("\"data:pr:ext:acl:grouptemp=~(" + grouptempfilters + ")\"")
		}

		urlParsed.Del("target")            // Delete target key
		urlParsed.Add("target", targetstr) // Adds recomputed target

		r.URL.RawQuery = urlParsed.Encode()

		l.logger.Info("filtered request /render:", zap.String("RawQuery:", r.URL.RawQuery))
		h.ServeHTTP(w, r)
	})
}

func (l *lbDataAuthzProxy) ProxyHandler(w http.ResponseWriter, r *http.Request) {
	l.logger.Info("sending request to metrictank")
	l.reverseproxy.ServeHTTP(w, r)
}

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4"

	"gitlab.com/lbauthdata/expr"
	"gitlab.com/lbauthdata/model"
)

func main() {

	// Prepare remote url for request proxying
	remote, err := url.Parse("http://localhost:6060")
	if err != nil {
		panic(err)
	}

	// initialize db connection
	conn, err := pgx.Connect(context.Background(), "user=keycloak password=password host=172.10.4.6 port=5432 database=lbauth sslmode=disable")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connection to database: %v\n", err)
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	// var id int64
	// var group_uuid pgtype.UUID
	// var role_uuid pgtype.UUID
	// Send the query to the server. The returned rows MUST be closed
	// before conn can be used again.
	rows, err := conn.Query(context.Background(),
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
		INNER JOIN roles ON roles_group_mapping.role_uuid = roles.uuid AND (
			group_uuid = 'e694ddf2-1790-addd-0f57-bc23b9d47fa3' OR 
			group_uuid = '0dbd3c3e-0b44-4a4e-aa32-569f8951dc79' OR 
			group_uuid = '5033357b-25f3-0124-180c-51029be60114' OR
			group_uuid = '521db0c7-78e9-36b8-a95b-da4ba8fe7f9e' )
		GROUP BY roles_group_mapping.group_uuid;`)
	if err != nil {
		fmt.Println(err)
	}
	// rows.Close is called by rows.Next when all rows are read
	// or an error occurs in Next or Scan. So it may optionally be
	// omitted if nothing in the rows.Next loop can panic. It is
	// safe to close rows multiple times.
	defer rows.Close()

	// var sum int32

	// Iterate through the result set
	for rows.Next() {

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
			fmt.Println(err)
		}

		fmt.Println(rows.Values())
		uuid_value, _ := group_uuid.Value()
		vpn_value, _ := vpn_iots.Value()
		fmt.Println(uuid_value)
		fmt.Println(vpn_value)
		// sum += id
	}

	// Any errors encountered by rows.Next or rows.Scan will be returned here
	if rows.Err() != nil {
		fmt.Println(err)
	}

	return

	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.ModifyResponse = CleanResponse
	http.HandleFunc("/tags/autoComplete/tags", LogMiddleware(TagsFilteringMiddleware(ProxyMiddleware(proxy))))
	http.HandleFunc("/tags/autoComplete/values", LogMiddleware(TagsFilteringMiddleware(ProxyMiddleware(proxy))))
	http.HandleFunc("/render", LogMiddleware(RenderFilteringMiddleware(ProxyMiddleware(proxy))))
	err = http.ListenAndServe(":9001", nil)
	if err != nil {
		panic(err)
	}
}

func LogMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("LOGGING INPUT REQUEST:")
		log.Println(r)
		h.ServeHTTP(w, r)
	})
}

func GroupPermissionsMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("GATHERING PERMSSIONS FROM USER GROUPS:")

		log.Println(r)
		h.ServeHTTP(w, r)
	})
}

func TagsFilteringMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("FILTERING REQUEST TAGS:")

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

		log.Println(r)
		h.ServeHTTP(w, r)
	})
}

func RenderFilteringMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("FILTERING REQUEST RENDER:")

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

		log.Println(r.URL.RawQuery)

		h.ServeHTTP(w, r)
	})
}

func ProxyMiddleware(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("SENDING REQUEST TO METRICTANK:")
		log.Println(r)
		p.ServeHTTP(w, r)
	}
}

func CleanResponse(r *http.Response) error {
	log.Println("CLEANING RESPONSE:")

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Error reading response body %s\n", err)
		return err
	}

	var jsonResp []byte

	switch r.Request.URL.Path {
	case "/render":
		var mtRespRender model.Series

		if err := json.Unmarshal(b, &mtRespRender); err != nil {
			fmt.Println(err)
		}

		cleanRender(&mtRespRender[0])

		log.Println(mtRespRender)

		jsonResp, err = json.Marshal(mtRespRender)
		if err != nil {
			log.Println(err)
		}

	case "/tags/autoComplete/tags":
		var mtRespTags model.Tags

		if err := json.Unmarshal(b, &mtRespTags); err != nil {
			fmt.Println(err)
		}

		err, mtRespTagsClean := cleanTags(mtRespTags)

		if err != nil {
			fmt.Println(err)
		}

		cleanTags(mtRespTags)

		log.Println(mtRespTags)

		jsonResp, err = json.Marshal(mtRespTagsClean)
		if err != nil {
			log.Println(err)
		}

	case "/tags/autoComplete/values":
		var mtRespTags model.Tags

		if err := json.Unmarshal(b, &mtRespTags); err != nil {
			fmt.Println(err)
		}

		err, mtRespTagsClean := cleanTags(mtRespTags)

		if err != nil {
			fmt.Println(err)
		}

		cleanTags(mtRespTags)

		log.Println(mtRespTags)

		jsonResp, err = json.Marshal(mtRespTagsClean)
		if err != nil {
			log.Println(err)
		}

		// defalut:
		// 	log.Fatal("Response type not matched")
	}

	buf := bytes.NewBufferString("")
	buf.Write(jsonResp)
	r.Body = ioutil.NopCloser(buf)
	r.Header["Content-Length"] = []string{fmt.Sprint(buf.Len())}
	return nil

	// var responseContent []interface{}
	// err := parseResponse(r, &responseContent)
	// if err != nil {
	// 	return err
	// }

	// log.Println(responseContent)
}

func cleanRender(mtResp *model.Serie) error {
	cleantarget := ""

	semistr := strings.Split(mtResp.Target, ";")
	for _, semis := range semistr {
		colsemistr := strings.Split(semis, ":")
		for i := 0; i < len(colsemistr); i++ {
			switch colsemistr[i] {
			case "pu":
				continue
			case "pr":
				continue
			case "data":
				continue
			case "temp":
				i++ // jump also next filed
				continue
			case "ext":
				continue
			case "int":
				continue
			case "acl":
				continue
			case "ou":
				i++ // jump also next filed
				continue
			case "cust":
				cleantarget += ";" + colsemistr[i+1]
				i++ // jump also next filed
				continue
			default: // all the unmatched slices are left unmatched
				eqcolssemistr := strings.Split(colsemistr[i], "=")
				for j := 0; j < len(eqcolssemistr); j++ {
					switch eqcolssemistr[j] {
					case "grouptemp":
						j++ // jump also the value
						continue
					case "temp":
						j++ // jump also the value
						continue
					case "creator":
						cleantarget += ";creator="
						continue
					default: // all the unmatched slices are left unmatched
						cleantarget += eqcolssemistr[j]
						break
					}
				}
				break
			}

		}
	}

	mtResp.Target = cleantarget

	// Cleaning tags
	for k, v := range mtResp.Tags {
		tagstr := strings.Split(k, ":")
		for i := 0; i < len(tagstr); i++ {
			switch tagstr[i] {
			case "name":
				continue
			case "data":
				continue
			case "ext":
				continue
			case "int":
				continue
			case "pu":
				continue
			case "cust":
				delete(mtResp.Tags, k)
				mtResp.Tags[tagstr[i+1]] = v
				continue
			case "pr":
				delete(mtResp.Tags, k)
				break
			case "acl":
				delete(mtResp.Tags, k)
				break

			default:
				delete(mtResp.Tags, k)
				break
			}
		}
	}

	return nil
}

func cleanTags(mtResp model.Tags) (err error, cleantags []string) {

	// Cleaning tags
	for _, tag := range mtResp {
		tagstr := strings.Split(tag, ":")
		for j := 0; j < len(tagstr); j++ {
			switch tagstr[j] {
			case "name":
				cleantags = append(cleantags, tagstr[j])
				continue
			case "data":
				continue
			case "ext":
				continue
			case "int":
				continue
			case "pu":
				continue
			case "cust":
				// cleantags = append(cleantags, tagstr[j])
				continue
			case "pr":
				continue
			case "acl":
				continue
			case "creator":
				continue
			case "temp":
				continue
			case "grouptemp":
				continue

			default:
				cleantags = append(cleantags, tagstr[j])
				continue
			}
		}
	}

	return nil, cleantags
}

func parseResponse(res *http.Response, unmarshalStruct *interface{}) error {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	res.Body.Close()

	res.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	return json.Unmarshal(body, unmarshalStruct)
}

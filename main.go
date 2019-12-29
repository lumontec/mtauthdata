package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"gitlab.com/lbauthdata/expr"
)

func main() {
	remote, err := url.Parse("http://localhost:6060")
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.ModifyResponse = CleanResponse
	http.HandleFunc("/tags/autoComplete/tags", LogMiddleware(TagsFiltering(ProxyMiddleware(proxy))))
	http.HandleFunc("/tags/autoComplete/values", LogMiddleware(TagsFiltering(ProxyMiddleware(proxy))))
	http.HandleFunc("/render", LogMiddleware(RenderFiltering(ProxyMiddleware(proxy))))
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

func TagsFiltering(h http.HandlerFunc) http.HandlerFunc {
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

		r.URL.RawQuery += "&expr=data:pr:ext:acl:grouptemp=~(" + grouptempfilters + ")"

		log.Println(r)
		h.ServeHTTP(w, r)
	})
}

func RenderFiltering(h http.HandlerFunc) http.HandlerFunc {
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

	type Point []float64

	// type Point struct {
	// 	Val float64 `json:"val,omitempty"`
	// 	Ts  uint32  `json:"ts,omitempty"`
	// }

	type Serie struct {
		Target     string            `json:"target,omitempty"` // for fetched data, set from models.Req.Target, i.e. the metric graphite key. for function output, whatever should be shown as target string (legend)
		Datapoints []Point           `json:"datapoints,omitempty"`
		Tags       map[string]string `json:"tags,omitempty"` // Must be set initially via call to `SetTags()`
		Interval   uint32            `json:"interval,omitempty"`
		QueryPatt  string            `json:"queryPatt,omitempty"` // to tie series back to request it came from. e.g. foo.bar.*, or if series outputted by func it would be e.g. scale(foo.bar.*,0.123456)
		QueryFrom  uint32            `json:"queryFrom,omitempty"` // to tie series back to request it came from
		QueryTo    uint32            `json:"queryTo,omitempty"`   // to tie series back to request it came from
	}

	type Series []Serie

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		fmt.Printf("Error reading response body %s\n", err)
		return err
	}

	var mtResp Series

	if err := json.Unmarshal(b, &mtResp); err != nil {
		fmt.Println(err)
		//fmt.Errorf(err.Error())
	}

	log.Println(mtResp)

	// // Cleaning target
	// s := strings.Split(mtResp[0].Target, ";")
	// mtResp[0].Target = s[0]

	// Cleaning target

	cleantarget := ""

	semistr := strings.Split(mtResp[0].Target, ";")
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

	mtResp[0].Target = cleantarget

	// Cleaning tags
	for k, v := range mtResp[0].Tags {
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
				delete(mtResp[0].Tags, k)
				mtResp[0].Tags[tagstr[i+1]] = v
				continue
			case "pr":
				delete(mtResp[0].Tags, k)
				break
			case "acl":
				delete(mtResp[0].Tags, k)
				break

			default:
				delete(mtResp[0].Tags, k)
				break
			}
		}
	}

	log.Println(mtResp)

	jsonData, err := json.Marshal(mtResp)
	if err != nil {
		log.Println(err)
	}

	buf := bytes.NewBufferString("")
	buf.Write(jsonData)
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

func parseResponse(res *http.Response, unmarshalStruct *interface{}) error {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}
	res.Body.Close()

	res.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	return json.Unmarshal(body, unmarshalStruct)
}

package main

import (
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

		grouptemps := []string{"group:dom:e34ba21c74c289ba894b75ae6c76d22f:temp:hot", "group:ou:e34ba21c74c289ba894b75ae6c76d22f:temp:cold"}

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

		grouptemps := []string{"group:dom:e34ba21c74c289ba894b75ae6c76d22f:temp:hot", "group:ou:e34ba21c74c289ba894b75ae6c76d22f:temp:cold"}

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

		rawquery := ""

		for _, expr := range exprs {
			rawquery += expr.ApplyQueryFilters("\"data:pr:ext:acl:grouptemp=~(" + grouptempfilters + ")\"")
		}

		r.URL.RawQuery = "render?target=" + rawquery

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
	log.Println("MODIFYING RESPONSE FROM METRICTANK:")
	log.Println(r)

	return nil
}

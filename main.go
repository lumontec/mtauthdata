package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {
	remote, err := url.Parse("http://localhost:8080")
	if err != nil {
		panic(err)
	}

	proxy := httputil.NewSingleHostReverseProxy(remote)
	proxy.ModifyResponse = UpdateResponse
	http.HandleFunc("/", LogMiddleware(ProxyMiddleware(proxy)))
	err = http.ListenAndServe(":9001", nil)
	if err != nil {
		panic(err)
	}
}

func ProxyMiddleware(p *httputil.ReverseProxy) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Println("SENDING REQUEST TO METRICTANK:")
		log.Println(r)
		p.ServeHTTP(w, r)
	}
}

func LogMiddleware(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Println("LOGGING INPUT REQUEST:")
		log.Println(r)
		h.ServeHTTP(w, r) // call ServeHTTP on the original handler
	})
}

func UpdateResponse(r *http.Response) error {
	log.Println("MODIFYING RESPONSE FROM METRICTANK:")
	log.Println(r)
	return nil
}

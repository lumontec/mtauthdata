package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"strings"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"

	"lbauthdata/interfaces"
	"lbauthdata/model"

	"go.uber.org/zap"
)

type Config struct {
	Upstreamurl        string
	ExposedPort        string
	PostgresConfig     string
	EnableJSONLogging  bool
	DisableAllLogging  bool
	Verbose            bool
	Opaurl             string
	HttpCallTimeoutSec int64
}

type lbDataAuthzProxy struct {
	config       *Config
	upstream     *url.URL
	logger       *zap.Logger
	reverseproxy *httputil.ReverseProxy
	Permissions  interfaces.PermissionProvider
	Authz        interfaces.AuthzProvider
	// httpclient   *http.Client
}

func NewLbDataAuthzProxy(config *Config) (*lbDataAuthzProxy, error) {
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

	lbdataauthz.reverseproxy = httputil.NewSingleHostReverseProxy(lbdataauthz.upstream)
	lbdataauthz.reverseproxy.ModifyResponse = lbdataauthz.CleanResponse

	return lbdataauthz, nil
}

func (l *lbDataAuthzProxy) RunServer() error {

	l.logger.Info("starting the service...", zap.String("port:", l.config.ExposedPort))
	r := l.createServerRouting()

	err := http.ListenAndServe(l.config.ExposedPort, r)
	if err != nil {
		panic(err)
	}

	return nil
}

func (l *lbDataAuthzProxy) createServerRouting() chi.Router {
	r := chi.NewRouter()

	// A good base middleware stack
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer) // This middleware avoids crash on panics ! By now I want to crash

	// Set a timeout value on the request context (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	r.Use(middleware.Timeout(60 * time.Second))

	r.Mount("/debug", middleware.Profiler())

	r.Get("/tags/autoComplete/tags",
		l.GroupPermissionsMiddleware(
			l.AuthzEnforcementMiddleware(
				l.TagsFilteringMiddleware(
					l.ProxyHandler))))

	r.Get("/tags/autoComplete/values",
		l.GroupPermissionsMiddleware(
			l.AuthzEnforcementMiddleware(
				l.TagsFilteringMiddleware(
					l.ProxyHandler))))

	r.Get("/render",
		l.GroupPermissionsMiddleware(
			l.AuthzEnforcementMiddleware(
				l.RenderFilteringMiddleware(
					l.ProxyHandler))))

	return r
}

func createLogger(config *Config) (*zap.Logger, error) {
	if config.DisableAllLogging {
		return zap.NewNop(), nil
	}

	c := zap.NewProductionConfig()
	c.DisableStacktrace = true
	c.DisableCaller = true
	// are we enabling json logging?
	if !config.EnableJSONLogging {
		c.Encoding = "console"
	}
	// are we running verbose mode?
	if config.Verbose {
		c.DisableCaller = false
		c.Development = true
		c.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	}

	return c.Build()
}

func (l *lbDataAuthzProxy) CleanResponse(r *http.Response) error {

	reqId := middleware.GetReqID(r.Request.Context())

	b, err := ioutil.ReadAll(r.Body)
	if err != nil {
		l.logger.Error("Error reading MT response body", zap.String("error", err.Error()))
		return err
	}

	var jsonResp []byte

	switch r.Request.URL.Path {
	case "/render":
		var mtRespRender model.Series

		if err := json.Unmarshal(b, &mtRespRender); err != nil {
			l.logger.Error("Error unmarshalling MT response body", zap.String("error", err.Error()))
			return err
		}

		l.logger.Info("pre-clean response:", zap.Any("/render", mtRespRender), zap.String("reqid:", reqId))

		cleanRender(&mtRespRender[0])

		l.logger.Info("cleaned response:", zap.Any("/render", mtRespRender), zap.String("reqid:", reqId))

		jsonResp, err = json.Marshal(mtRespRender)
		if err != nil {
			l.logger.Error("Error marshalling json render response", zap.String("error", err.Error()))
			return err
		}

	case "/tags/autoComplete/tags":
		var mtRespTags model.Tags

		if err := json.Unmarshal(b, &mtRespTags); err != nil {
			l.logger.Error("Error unmarshalling MT response body", zap.String("error", err.Error()))
			return err
		}

		l.logger.Info("pre-clean response:", zap.Any("/tags/autoComplete/tags", mtRespTags), zap.String("reqid:", reqId))

		err, mtRespTagsClean := cleanTags(mtRespTags)
		if err != nil {
			l.logger.Error("Error cleaning MT response tag keys", zap.String("error", err.Error()))
			return err
		}

		l.logger.Info("cleaned response:", zap.Any("/tags/autoComplete/tags", mtRespTagsClean), zap.String("reqid:", reqId))

		jsonResp, err = json.Marshal(mtRespTagsClean)
		if err != nil {
			l.logger.Error("Error unmarshalling MT response body", zap.String("error", err.Error()))
			return err
		}

	case "/tags/autoComplete/values":
		var mtRespTags model.Tags

		if err := json.Unmarshal(b, &mtRespTags); err != nil {
			l.logger.Error("Error unmarshalling MT response body", zap.String("error", err.Error()))
			return err
		}

		l.logger.Info("pre-clean response:", zap.Any("/tags/autoComplete/values", mtRespTags), zap.String("reqid:", reqId))

		err, mtRespTagsClean := cleanTags(mtRespTags)
		if err != nil {
			l.logger.Error("Error cleaning MT response tag values", zap.String("error", err.Error()))
		}

		l.logger.Info("cleaned response:", zap.Any("/tags/autoComplete/values", mtRespTagsClean), zap.String("reqid:", reqId))

		jsonResp, err = json.Marshal(mtRespTagsClean)
		if err != nil {
			l.logger.Error("Error unmarshalling MT response body", zap.String("error", err.Error()))
			return err
		}

		//	defalut:
		//		l.logger.Error("Error unmarshalling MT response body", zap.String("error", err.Error()))
		//		return nil
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

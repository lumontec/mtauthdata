package server

import (
	"testing"
)

func TestNewLbDataAuthzProxy(t *testing.T) {

	config := &Config{
		Upstreamurl:        "http://localhost:6060",
		ExposedPort:        ":9001",
		PostgresConfig:     "user=kk password=psw host=172.10.4.6 port=5432 database=lbauth sslmode=disable",
		EnableJSONLogging:  false,
		DisableAllLogging:  false,
		Verbose:            false,
		Opaurl:             "http://localhost:8181/v1/data/authzdata",
		HttpCallTimeoutSec: 10}

	_, err := NewLbDataAuthzProxy(config)

	if err != nil {
		t.Fatal(err)
	}
}

// func testCleanTags(t *testing.T) {

// 	cases := []struct {
// 		intags   model.Tags
// 		wantags  []string
// 	}{
// 		{"/render?target=demotags.iot1.metric0&from=-5min&until=now&format=json&maxDataPoints=653", []string{"group:0dbd3c3e-0b44-4a4e-aa32-569f8951dc79:temp:read", "group:0dbd3c3e-0b44-4a4e-aa32-569f8951dc79:temp:cold"}, "target=demotags.iot1.metric0&from=-5min&until=now&format=json&maxDataPoints=653&expr=data:pr:ext:acl:grouptemp=~(^group:0dbd3c3e-0b44-4a4e-aa32-569f8951dc79:temp:read$|^group:0dbd3c3e-0b44-4a4e-aa32-569f8951dc79:temp:cold$)"},
// 		{"/render?target=demotags.iot1.metric0&from=-5min&until=now&format=json&maxDataPoints=653", []string{"group:0dbd3c3e-0b44-4a4e-aa32-569f8951dc79:temp:cold"}, "target=demotags.iot1.metric0&from=-5min&until=now&format=json&maxDataPoints=653&expr=data:pr:ext:acl:grouptemp=~(^group:0dbd3c3e-0b44-4a4e-aa32-569f8951dc79:temp:cold$)"},
// 		{"/render?target=demotags.iot1.metric0&from=-5min&until=now&format=json&maxDataPoints=653", []string{}, "target=demotags.iot1.metric0&from=-5min&until=now&format=json&maxDataPoints=653&expr=data:pr:ext:acl:grouptemp=~()"},
// 	}

// 	for _, tc := range cases {
// 		tname, _ := json.Marshal(tc.intags)
// 		t.Run(string(tname), func(t *testing.T) {

// 			assert.Equal(t, tc.gotquery, r.URL.RawQuery)
// 		}
// 	}

// }

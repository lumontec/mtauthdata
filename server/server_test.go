package server

import (
	"encoding/json"
	"lbauthdata/model"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewLbDataAuthzProxy(t *testing.T) {

	config := &Config{
		Upstreamurl:        "http://localhost:6060",
		ExposedPort:        ":9001",
		PostgresConfig:     "user=kk password=psw host=172.10.4.6 port=5432 database=lbauth sslmode=disable",
		Opaurl:             "http://localhost:8181/v1/data/authzdata",
		HttpCallTimeoutSec: 10}

	_, err := NewLbDataAuthzProxy(config)

	if err != nil {
		t.Fatal(err)
	}
}

func TestCleanTags(t *testing.T) {

	cases := []struct {
		inTags   model.Tags
		wantTags []string
	}{
		{model.Tags{"name:test"}, []string{"name", "test"}},
		{model.Tags{"data:test"}, []string{"test"}},
		{model.Tags{"ext:test"}, []string{"test"}},
		{model.Tags{"int:test"}, []string{"test"}},
		{model.Tags{"pu:test"}, []string{"test"}},
		{model.Tags{"cust:test"}, []string{"test"}},
		{model.Tags{"pr:test"}, []string{"test"}},
		{model.Tags{"acl:test"}, []string{"test"}},
		{model.Tags{"creator:test"}, []string{"test"}},
		{model.Tags{"temp:test"}, []string{"test"}},
		{model.Tags{"grouptemp:test"}, []string{"test"}},
	}
	for _, tc := range cases {
		tname, _ := json.Marshal(tc.inTags)
		t.Run(string(tname), func(t *testing.T) {
			_, gotTags := cleanTags(tc.inTags)
			assert.Equal(t, tc.wantTags, gotTags)
		})
	}

}

func TestCleanRender(t *testing.T) {

	cases := []struct {
		inSerie   model.Serie
		wantSerie model.Serie
	}{
		{
			inSerie: model.Serie{
				Target:     "demotags.iot1.metric0",
				Datapoints: []model.Point{},
				Tags: map[string]string{
					"name": "demotags.iot1.metric0",
				},
				Interval:  0,
				QueryPatt: "",
				QueryFrom: 0,
				QueryTo:   1000,
			},
			wantSerie: model.Serie{
				Target:     "demotags.iot1.metric0",
				Datapoints: []model.Point{},
				Tags: map[string]string{
					"name": "demotags.iot1.metric0",
				},
				Interval:  0,
				QueryPatt: "",
				QueryFrom: 0,
				QueryTo:   1000,
			},
		},
	}

	for _, tc := range cases {
		tname, _ := json.Marshal(tc.inSerie)
		t.Run(string(tname), func(t *testing.T) {
			gotSerie, _ := cleanRender(tc.inSerie)
			assert.Equal(t, tc.inSerie, gotSerie)
		})
	}
}

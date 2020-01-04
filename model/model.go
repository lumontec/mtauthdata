package model

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

type GroupPermMappings struct {
	Groups []Mapping `json:"groups"`
}

type Mapping struct {
	Group_uuid  string      `json:"group_uuid"`
	Permissions Permissions `json:"permissions"`
}

type Permissions struct {
	Admin_iots     bool `json:"admin_iots"`
	View_iots      bool `json:"view_iots"`
	Configure_iots bool `json:"configure_iots"`
	Vpn_iots       bool `json:"vpn_iots"`
	Webpage_iots   bool `json:"webpage_iots"`
	Hmi_iots       bool `json:"hmi_iots"`
	Data_admin     bool `json:"data_admin"`
	Data_read      bool `json:"data_read"`
	Data_cold_read bool `json:"data_cold_read"`
	Data_warm_read bool `json:"data_warm_read"`
	Data_hot_read  bool `json:"data_hot_read"`
	Services_admin bool `json:"services_admin"`
	Billing_admin  bool `json:"billing_admin"`
	Org_admin      bool `json:"org_admin"`
}

type Series []Serie

type Tags []string

type OpaResp struct {
	Result OpaJudgement `json:"result"`
}

type OpaJudgement struct {
	Allow          bool     `json:"allow"`
	Allowed_groups []string `json:"allowed_groups"`
	Read_allowed   []string `json:"read_allowed"`
	Cold_allowed   []string `json:"cold_allowed"`
	Warm_allowed   []string `json:"warm_allowed"`
	Hot_allowed    []string `json:"hot_allowed"`
}

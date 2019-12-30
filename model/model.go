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

type Series []Serie

type Tags []string


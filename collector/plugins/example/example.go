package example

import (
	"bes-agent/collector"
	"bes-agent/common/metric"
	"bes-agent/common/plugin"
)

// NewExample XXX
func NewExample(conf plugin.InitConfig) plugin.Plugin {
	return &Example{}
}

// Example XXX
type Example struct {
}

// Check XXX
func (e *Example) Check(agg metric.Aggregator) error {
	return nil
}

func init() {
	collector.Add("example", NewExample)
}

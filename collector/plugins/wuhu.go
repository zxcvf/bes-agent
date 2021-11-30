package plugins

import (
	"bes-agent/collector"
	"bes-agent/common/metric"
	"bes-agent/common/plugin"
	"fmt"
)

// NewRedis XXX
func NewWuhu(conf plugin.InitConfig) plugin.Plugin {
	return &wuhu{}
}

// Redis XXX
type wuhu struct {
}

// Check XXX
func (r *wuhu) Check(agg metric.Aggregator) error {
	fmt.Println("    wuhu >. <")
	x := make(map[string]interface{})
	y := make([]string, 0)
	agg.AddMetrics("gauge", "asd", x, y, "")
	agg.AddMetrics("gauge2", "asd", x, y, "")
	return nil
}

func init() {
	collector.Add("wuhu", NewWuhu)
}

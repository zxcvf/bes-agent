package statsd

import (
	"fmt"
	"time"

	"bes-agent/common/api"
	"bes-agent/common/config"
	"bes-agent/common/emitter"
	"bes-agent/common/log"
)

// Reporter XXX
type Reporter struct {
	*emitter.Emitter

	api  *api.API
	conf *config.Config
}

// NewReporter creates a new instance of Reporter.
func NewReporter(conf *config.Config) *Reporter {
	emitter := emitter.NewEmitter("Statsd")
	api := api.NewAPI(conf.GetForwarderAddrWithScheme(), conf.GlobalConfig.LicenseKey, 5*time.Second)

	r := &Reporter{
		Emitter: emitter,
		api:     api,
		conf:    conf,
	}
	r.Emitter.Parent = r

	return r
}

// Post sends the metrics to Forwarder API.
func (r *Reporter) Post(metrics []interface{}) error {
	start := time.Now()
	payload := Payload{}
	payload.Series = metrics

	err := r.api.SubmitMetrics(&payload)
	elapsed := time.Since(start)
	if err == nil {
		fmt.Printf("Reporter Post batch of %d metrics in %s", len(metrics), elapsed)

		log.Debugf("Post batch of %d metrics in %s",
			len(metrics), elapsed)
	}
	return err
}

package server

import (
	"context"
	"time"

	"github.com/altstory/go-metrics"
	"github.com/altstory/go-runner"
)

var (
	httpMetrics struct {
		QPS, ProcTime, MaxProcTime, Count, Failure *metrics.Metric
	}
	serverMetrics struct {
		Goroutine, Panic *metrics.Metric
	}
)

func init() {
	runner.OnStart(func(ctx context.Context) error {
		initMetrics()
		return nil
	})
}

func initMetrics() {
	httpMetrics.QPS = metrics.Define(&metrics.Def{
		Category: "api_qps",
		Method:   metrics.Sum,
		Duration: time.Second,
	})
	httpMetrics.ProcTime = metrics.Define(&metrics.Def{
		Category: "api_proc_time",
		Method:   metrics.Average,
	})
	httpMetrics.MaxProcTime = metrics.Define(&metrics.Def{
		Category: "api_max_proc_time",
		Method:   metrics.Maximum,
	})
	httpMetrics.Count = metrics.Define(&metrics.Def{
		Category: "api_count",
		Method:   metrics.Sum,
	})
	httpMetrics.Failure = metrics.Define(&metrics.Def{
		Category: "api_failure",
		Method:   metrics.Sum,
	})

	serverMetrics.Goroutine = metrics.Define(&metrics.Def{
		Category: "server_goroutine",
		Method:   metrics.Maximum,
	})
	serverMetrics.Panic = metrics.Define(&metrics.Def{
		Category: "server_panic",
		Method:   metrics.Sum,
	})
}

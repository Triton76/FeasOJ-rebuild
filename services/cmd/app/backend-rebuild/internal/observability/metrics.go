package observability

import (
	"sync/atomic"
	"time"
)

var queueConsumeTotal atomic.Uint64
var writebackSuccessTotal atomic.Uint64
var writebackFailureTotal atomic.Uint64
var scoreboardQuerySuccessTotal atomic.Uint64
var scoreboardQueryFailureTotal atomic.Uint64
var scoreboardLatencyTotalMicros atomic.Uint64

func IncQueueConsume() {
	queueConsumeTotal.Add(1)
}

func IncWritebackSuccess() {
	writebackSuccessTotal.Add(1)
}

func IncWritebackFailure() {
	writebackFailureTotal.Add(1)
}

func ObserveScoreboardQueryLatency(d time.Duration, success bool) {
	if success {
		scoreboardQuerySuccessTotal.Add(1)
	} else {
		scoreboardQueryFailureTotal.Add(1)
	}
	if d > 0 {
		scoreboardLatencyTotalMicros.Add(uint64(d.Microseconds()))
	}
}

func Snapshot() map[string]uint64 {
	return map[string]uint64{
		"queue_consume_total":            queueConsumeTotal.Load(),
		"writeback_success_total":        writebackSuccessTotal.Load(),
		"writeback_failure_total":        writebackFailureTotal.Load(),
		"scoreboard_query_success_total": scoreboardQuerySuccessTotal.Load(),
		"scoreboard_query_failure_total": scoreboardQueryFailureTotal.Load(),
		"scoreboard_latency_total_micros": scoreboardLatencyTotalMicros.Load(),
	}
}

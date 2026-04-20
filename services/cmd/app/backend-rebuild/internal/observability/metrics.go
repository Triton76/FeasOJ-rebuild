package observability

import (
	"sync"
	"sync/atomic"
	"time"
)

var (
	queueConsumeTotal     int64
	queueEnqueueTotal     int64
	queueRetryTotal       int64
	queueDLQTotal         int64
	writebackSuccessTotal int64
	writebackFailureTotal int64
	writebackLatencySumMS int64
	writebackTotal        int64

	scoreboardMu           sync.Mutex
	scoreboardQueryTotal   int64
	scoreboardQuerySuccess int64
	scoreboardLatencySumMS int64
)

func IncQueueConsume() {
	atomic.AddInt64(&queueConsumeTotal, 1)
}

func IncQueueEnqueue() {
	atomic.AddInt64(&queueEnqueueTotal, 1)
}

func IncQueueRetry() {
	atomic.AddInt64(&queueRetryTotal, 1)
}

func IncQueueDLQ() {
	atomic.AddInt64(&queueDLQTotal, 1)
}

func IncWritebackSuccess() {
	atomic.AddInt64(&writebackSuccessTotal, 1)
}

func IncWritebackFailure() {
	atomic.AddInt64(&writebackFailureTotal, 1)
}

func ObserveWritebackLatency(d time.Duration) {
	ms := d.Milliseconds()
	if ms < 0 {
		ms = 0
	}
	atomic.AddInt64(&writebackLatencySumMS, ms)
	atomic.AddInt64(&writebackTotal, 1)
}

func ObserveScoreboardQueryLatency(d time.Duration, success bool) {
	ms := d.Milliseconds()
	if ms < 0 {
		ms = 0
	}

	scoreboardMu.Lock()
	defer scoreboardMu.Unlock()
	scoreboardQueryTotal++
	if success {
		scoreboardQuerySuccess++
	}
	scoreboardLatencySumMS += ms
}

func Snapshot() map[string]any {
	scoreboardMu.Lock()
	total := scoreboardQueryTotal
	success := scoreboardQuerySuccess
	latencySum := scoreboardLatencySumMS
	scoreboardMu.Unlock()

	avgLatency := int64(0)
	if total > 0 {
		avgLatency = latencySum / total
	}
	writebackTotalCount := atomic.LoadInt64(&writebackTotal)
	writebackLatencyAvg := int64(0)
	if writebackTotalCount > 0 {
		writebackLatencyAvg = atomic.LoadInt64(&writebackLatencySumMS) / writebackTotalCount
	}

	return map[string]any{
		"queue_consume_total":      atomic.LoadInt64(&queueConsumeTotal),
		"queue_enqueue_total":      atomic.LoadInt64(&queueEnqueueTotal),
		"queue_retry_total":        atomic.LoadInt64(&queueRetryTotal),
		"queue_dlq_total":          atomic.LoadInt64(&queueDLQTotal),
		"writeback_success_total":  atomic.LoadInt64(&writebackSuccessTotal),
		"writeback_failure_total":  atomic.LoadInt64(&writebackFailureTotal),
		"writeback_avg_ms":         writebackLatencyAvg,
		"scoreboard_query_total":   total,
		"scoreboard_query_success": success,
		"scoreboard_query_avg_ms":  avgLatency,
	}
}

package queue

import (
	"FeasOJ/app/backend-rebuild/internal/observability"
	"FeasOJ/app/backend-rebuild/internal/ports"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQConfig struct {
	URL        string
	Exchange   string
	MainQueue  string
	RetryQueue string
	DLQ        string
	Prefetch   int
	MaxRetries int
	Backoff    []time.Duration
}

type RabbitMQSubmissionQueue struct {
	conn *amqp.Connection
	ch   *amqp.Channel
	cfg  RabbitMQConfig

	mu      sync.Mutex
	pending map[int64]uint64
	seen    map[int64]struct{}
}

func NewRabbitMQSubmissionQueue(cfg RabbitMQConfig) (*RabbitMQSubmissionQueue, error) {
	if cfg.URL == "" {
		return nil, fmt.Errorf("rabbitmq url is required")
	}
	if cfg.Exchange == "" {
		cfg.Exchange = "judge.submission.exchange"
	}
	if cfg.MainQueue == "" {
		cfg.MainQueue = "judge.submission.main"
	}
	if cfg.RetryQueue == "" {
		cfg.RetryQueue = "judge.submission.retry"
	}
	if cfg.DLQ == "" {
		cfg.DLQ = "judge.submission.dlq"
	}
	if cfg.Prefetch <= 0 {
		cfg.Prefetch = 1
	}
	if len(cfg.Backoff) == 0 {
		cfg.Backoff = []time.Duration{2 * time.Second, 5 * time.Second, 15 * time.Second, 30 * time.Second, 60 * time.Second}
	}

	conn, err := amqp.Dial(cfg.URL)
	if err != nil {
		return nil, err
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	if err := ch.Confirm(false); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}
	if err := ch.Qos(cfg.Prefetch, 0, false); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}
	if err := declareTopology(ch, cfg); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return nil, err
	}

	return &RabbitMQSubmissionQueue{
		conn:    conn,
		ch:      ch,
		cfg:     cfg,
		pending: make(map[int64]uint64),
		seen:    make(map[int64]struct{}),
	}, nil
}

func declareTopology(ch *amqp.Channel, cfg RabbitMQConfig) error {
	if err := ch.ExchangeDeclare(cfg.Exchange, amqp.ExchangeDirect, true, false, false, false, nil); err != nil {
		return err
	}
	if _, err := ch.QueueDeclare(cfg.MainQueue, true, false, false, false, amqp.Table{"x-dead-letter-exchange": cfg.Exchange, "x-dead-letter-routing-key": "judge.submission.dead"}); err != nil {
		return err
	}
	if _, err := ch.QueueDeclare(cfg.RetryQueue, true, false, false, false, amqp.Table{"x-dead-letter-exchange": cfg.Exchange, "x-dead-letter-routing-key": "judge.submission.enqueue"}); err != nil {
		return err
	}
	if _, err := ch.QueueDeclare(cfg.DLQ, true, false, false, false, nil); err != nil {
		return err
	}
	if err := ch.QueueBind(cfg.MainQueue, "judge.submission.enqueue", cfg.Exchange, false, nil); err != nil {
		return err
	}
	if err := ch.QueueBind(cfg.RetryQueue, "judge.submission.retry", cfg.Exchange, false, nil); err != nil {
		return err
	}
	if err := ch.QueueBind(cfg.DLQ, "judge.submission.dead", cfg.Exchange, false, nil); err != nil {
		return err
	}
	return nil
}

func (q *RabbitMQSubmissionQueue) Close() error {
	if q == nil {
		return nil
	}
	var firstErr error
	if q.ch != nil {
		firstErr = q.ch.Close()
	}
	if q.conn != nil {
		if err := q.conn.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

func (q *RabbitMQSubmissionQueue) Enqueue(ctx context.Context, job ports.SubmissionJob) error {
	if q == nil {
		return ports.ErrNotImplemented
	}
	q.mu.Lock()
	if _, ok := q.seen[job.SubmissionID]; ok {
		q.mu.Unlock()
		return nil
	}
	q.seen[job.SubmissionID] = struct{}{}
	q.mu.Unlock()

	body, err := json.Marshal(job)
	if err != nil {
		return err
	}
	confirm := q.ch.NotifyPublish(make(chan amqp.Confirmation, 1))
	if err := q.ch.PublishWithContext(ctx, q.cfg.Exchange, "judge.submission.enqueue", false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		MessageId:    strconv.FormatInt(job.SubmissionID, 10),
		Timestamp:    time.Now().UTC(),
		Body:         body,
		Headers:      amqp.Table{"attempt": int32(1)},
	}); err != nil {
		return err
	}
	select {
	case ack := <-confirm:
		if !ack.Ack {
			return errors.New("rabbitmq publisher confirm nack")
		}
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(5 * time.Second):
		return errors.New("rabbitmq publisher confirm timeout")
	}
	observability.IncQueueEnqueue()
	observability.LogJSON("submission.enqueued", map[string]any{"submission_id": job.SubmissionID, "queue_name": q.cfg.MainQueue, "attempt": 1})
	return nil
}

func (q *RabbitMQSubmissionQueue) Dequeue(ctx context.Context) (*ports.SubmissionJob, error) {
	if q == nil {
		return nil, ports.ErrNotImplemented
	}
	msg, ok, err := q.ch.Get(q.cfg.MainQueue, false)
	if err != nil {
		if errors.Is(err, amqp.ErrClosed) {
			return nil, err
		}
		return nil, ports.ErrQueueEmpty
	}
	if !ok {
		return nil, ports.ErrQueueEmpty
	}
	var job ports.SubmissionJob
	if err := json.Unmarshal(msg.Body, &job); err != nil {
		_ = msg.Reject(false)
		return nil, err
	}
	q.mu.Lock()
	q.pending[job.SubmissionID] = msg.DeliveryTag
	q.mu.Unlock()
	observability.LogJSON("submission.dequeue", map[string]any{"submission_id": job.SubmissionID, "queue_name": q.cfg.MainQueue, "attempt": messageAttempt(msg.Headers)})
	return &job, nil
}

func (q *RabbitMQSubmissionQueue) Acknowledge(ctx context.Context, submissionID int64) error {
	if q == nil {
		return ports.ErrNotImplemented
	}
	q.mu.Lock()
	tag, ok := q.pending[submissionID]
	if ok {
		delete(q.pending, submissionID)
	}
	q.mu.Unlock()
	if !ok {
		return ports.ErrNotFound
	}
	if err := q.ch.Ack(tag, false); err != nil {
		return err
	}
	observability.LogJSON("submission.ack", map[string]any{"submission_id": submissionID, "queue_name": q.cfg.MainQueue})
	return nil
}

func (q *RabbitMQSubmissionQueue) Size(ctx context.Context) (int, error) {
	if q == nil {
		return 0, ports.ErrNotImplemented
	}
	info, err := q.ch.QueueInspect(q.cfg.MainQueue)
	if err != nil {
		return 0, err
	}
	return info.Messages, nil
}

func messageAttempt(headers amqp.Table) int {
	if headers == nil {
		return 1
	}
	if v, ok := headers["attempt"]; ok {
		switch n := v.(type) {
		case int32:
			return int(n)
		case int64:
			return int(n)
		case int:
			return n
		case byte:
			return int(n)
		}
	}
	return 1
}

type RetryPolicy struct {
	MaxRetries int
	Backoff    []time.Duration
	QueueName  string
	DLQName    string
}

func (p RetryPolicy) BackoffFor(attempt int) time.Duration {
	if len(p.Backoff) == 0 {
		p.Backoff = []time.Duration{2 * time.Second, 5 * time.Second, 15 * time.Second, 30 * time.Second, 60 * time.Second}
	}
	if attempt <= 0 {
		return p.Backoff[0]
	}
	if attempt >= len(p.Backoff) {
		return p.Backoff[len(p.Backoff)-1]
	}
	return p.Backoff[attempt]
}

type RetryHandler struct {
	channel *amqp.Channel
	policy  RetryPolicy
}

func NewRetryHandler(channel *amqp.Channel, policy RetryPolicy) *RetryHandler {
	return &RetryHandler{channel: channel, policy: policy}
}

func (h *RetryHandler) HandleRetry(ctx context.Context, job ports.SubmissionJob, attempt int) error {
	if h == nil || h.channel == nil {
		return ports.ErrNotImplemented
	}
	if h.policy.MaxRetries > 0 && attempt >= h.policy.MaxRetries {
		return h.sendDLQ(ctx, job, attempt)
	}
	time.Sleep(h.policy.BackoffFor(attempt))
	return h.publish(ctx, job, attempt+1)
}

func (h *RetryHandler) publish(ctx context.Context, job ports.SubmissionJob, attempt int) error {
	body, err := json.Marshal(job)
	if err != nil {
		return err
	}
	if err := h.channel.PublishWithContext(ctx, "judge.submission.exchange", "judge.submission.retry", false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		MessageId:    strconv.FormatInt(job.SubmissionID, 10),
		Timestamp:    time.Now().UTC(),
		Body:         body,
		Headers:      amqp.Table{"attempt": int32(attempt)},
	}); err != nil {
		return err
	}
	observability.IncQueueRetry()
	observability.LogJSON("submission.retry", map[string]any{"submission_id": job.SubmissionID, "attempt": attempt})
	return nil
}

func (h *RetryHandler) sendDLQ(ctx context.Context, job ports.SubmissionJob, attempt int) error {
	body, err := json.Marshal(job)
	if err != nil {
		return err
	}
	if err := h.channel.PublishWithContext(ctx, "judge.submission.exchange", "judge.submission.dead", false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		MessageId:    strconv.FormatInt(job.SubmissionID, 10),
		Timestamp:    time.Now().UTC(),
		Body:         body,
		Headers:      amqp.Table{"attempt": int32(attempt)},
	}); err != nil {
		return err
	}
	observability.IncQueueDLQ()
	observability.LogJSON("submission.dead_letter", map[string]any{"submission_id": job.SubmissionID, "attempt": attempt})
	return nil
}

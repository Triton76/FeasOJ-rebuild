//go:build integration
// +build integration

package queue

import (
	"FeasOJ/app/backend-rebuild/internal/ports"
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"
)

func TestRabbitMQSubmissionQueue_Live(t *testing.T) {
	url := os.Getenv("BACKEND_REBUILD_RABBITMQ_URL")
	if url == "" {
		url = "amqp://guest:guest@localhost:5672/"
	}

	suffix := strings.NewReplacer("/", ".", " ", ".").Replace(t.Name()) + fmt.Sprintf(".%d", time.Now().UnixNano())
	cfg := RabbitMQConfig{
		URL:        url,
		Exchange:   "it.exchange." + suffix,
		MainQueue:  "it.main." + suffix,
		RetryQueue: "it.retry." + suffix,
		DLQ:        "it.dlq." + suffix,
		Prefetch:   1,
	}

	q, err := NewRabbitMQSubmissionQueue(cfg)
	if err != nil {
		t.Fatalf("create queue: %v", err)
	}
	defer q.Close()

	job := ports.SubmissionJob{
		SubmissionID: 991001,
		UserID:       "u-live",
		ProblemID:    991,
		ContestID:    0,
		Language:     "go",
		SourceCode:   "package main",
	}

	if err := q.Enqueue(context.Background(), job); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	size, err := q.Size(context.Background())
	if err != nil {
		t.Fatalf("size after enqueue: %v", err)
	}
	if size != 1 {
		t.Fatalf("expected queue size 1, got %d", size)
	}

	if err := q.Enqueue(context.Background(), job); err != nil {
		t.Fatalf("duplicate enqueue should be idempotent: %v", err)
	}
	size, err = q.Size(context.Background())
	if err != nil {
		t.Fatalf("size after duplicate enqueue: %v", err)
	}
	if size != 1 {
		t.Fatalf("expected queue size to remain 1, got %d", size)
	}

	deq, err := q.Dequeue(context.Background())
	if err != nil {
		t.Fatalf("dequeue: %v", err)
	}
	if deq.SubmissionID != job.SubmissionID {
		t.Fatalf("expected submission_id %d, got %d", job.SubmissionID, deq.SubmissionID)
	}
	if err := q.Close(); err != nil {
		t.Fatalf("close before recovery check: %v", err)
	}

	q, err = NewRabbitMQSubmissionQueue(cfg)
	if err != nil {
		t.Fatalf("reopen queue: %v", err)
	}
	defer q.Close()
	size, err = q.Size(context.Background())
	if err != nil {
		t.Fatalf("size after reopen: %v", err)
	}
	if size != 1 {
		t.Fatalf("expected queue size 1 after recovery, got %d", size)
	}
	deq, err = q.Dequeue(context.Background())
	if err != nil {
		t.Fatalf("dequeue after recovery: %v", err)
	}
	if deq.SubmissionID != job.SubmissionID {
		t.Fatalf("expected recovered submission_id %d, got %d", job.SubmissionID, deq.SubmissionID)
	}
	if err := q.Acknowledge(context.Background(), job.SubmissionID); err != nil {
		t.Fatalf("acknowledge: %v", err)
	}
	size, err = q.Size(context.Background())
	if err != nil {
		t.Fatalf("size after ack: %v", err)
	}
	if size != 0 {
		t.Fatalf("expected queue size 0 after ack, got %d", size)
	}
}

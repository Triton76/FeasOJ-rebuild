//go:build integration
// +build integration

package queue

import "testing"

// Reliability tests are gated behind the integration build tag because they
// require a running RabbitMQ instance and the durable queue implementation.
func TestRabbitMQReliabilitySuitePlaceholder(t *testing.T) {
	t.Skip("integration-only: enable once RabbitMQ durable queue adapter is implemented")
}

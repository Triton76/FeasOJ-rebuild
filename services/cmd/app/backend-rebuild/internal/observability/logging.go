package observability

import (
	"encoding/json"
	"log"
	"time"
)

func LogJSON(event string, fields map[string]any) {
	payload := map[string]any{
		"event": event,
		"ts":    time.Now().UTC().Format(time.RFC3339Nano),
	}
	for k, v := range fields {
		payload[k] = v
	}

	line, err := json.Marshal(payload)
	if err != nil {
		log.Printf("[backend-rebuild] event=%s log_json_marshal_failed err=%v", event, err)
		return
	}
	log.Printf("%s", line)
}

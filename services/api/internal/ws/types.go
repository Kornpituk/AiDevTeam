package ws

import "encoding/json"

// Message represents a WebSocket message with a type and data payload.
type Message struct {
	Type string      `json:"type"`
	Data interface{} `json:"data"`
}

// Message types for real-time updates.
const (
	TypeRunUpdate     = "run_update"
	TypeStepUpdate    = "step_update"
	TypeMessageNew    = "message_new"
	TypeApprovalUpdate = "approval_update"
	TypeToolCallUpdate = "tool_call_update"
)

// NewEvent creates a JSON-encoded Message for broadcasting.
func NewEvent(eventType string, data interface{}) []byte {
	msg := Message{Type: eventType, Data: data}
	bytes, err := json.Marshal(msg)
	if err != nil {
		return []byte(`{"type":"error","data":"failed to marshal event"}`)
	}
	return bytes
}

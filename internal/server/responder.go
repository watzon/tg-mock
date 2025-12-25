// internal/server/responder.go
package server

import (
	"sync/atomic"
	"time"

	"github.com/watzon/tg-mock/gen"
)

// Responder generates appropriate responses for Bot API methods
type Responder struct {
	messageIDCounter int64
}

// NewResponder creates a new Responder instance
func NewResponder() *Responder {
	return &Responder{}
}

// nextMessageID returns a unique, incrementing message ID
func (r *Responder) nextMessageID() int64 {
	return atomic.AddInt64(&r.messageIDCounter, 1)
}

// Generate produces an appropriate response based on the method's return type
func (r *Responder) Generate(spec gen.MethodSpec, params map[string]interface{}) (interface{}, error) {
	if len(spec.Returns) == 0 {
		return true, nil
	}

	returnType := spec.Returns[0]

	switch returnType {
	case "Boolean":
		return true, nil
	case "User":
		return r.generateUser(params), nil
	case "Message":
		return r.generateMessage(params), nil
	case "Array of Update":
		return []interface{}{}, nil
	case "File":
		return r.generateFile(params), nil
	default:
		// Return a basic success response
		return map[string]interface{}{}, nil
	}
}

// generateUser creates a User response with bot fixture data
func (r *Responder) generateUser(params map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"id":         int64(123456789),
		"is_bot":     true,
		"first_name": "TestBot",
		"username":   "test_bot",
	}
}

// generateMessage creates a Message response reflecting input parameters
func (r *Responder) generateMessage(params map[string]interface{}) map[string]interface{} {
	chatID := int64(1)
	if id, ok := params["chat_id"].(int64); ok {
		chatID = id
	} else if id, ok := params["chat_id"].(float64); ok {
		chatID = int64(id)
	}

	msg := map[string]interface{}{
		"message_id": r.nextMessageID(),
		"date":       time.Now().Unix(),
		"chat": map[string]interface{}{
			"id":   chatID,
			"type": "private",
		},
	}

	// Reflect input parameters
	if text, ok := params["text"].(string); ok {
		msg["text"] = text
	}

	return msg
}

// generateFile creates a File response for getFile method
func (r *Responder) generateFile(params map[string]interface{}) map[string]interface{} {
	fileID := "placeholder"
	if id, ok := params["file_id"].(string); ok {
		fileID = id
	}

	// Ensure file_id is at least 8 characters for unique_id generation
	uniqueIDSuffix := fileID
	if len(fileID) >= 8 {
		uniqueIDSuffix = fileID[:8]
	}

	return map[string]interface{}{
		"file_id":        fileID,
		"file_unique_id": "unique_" + uniqueIDSuffix,
		"file_size":      1024,
		"file_path":      "documents/file.txt",
	}
}

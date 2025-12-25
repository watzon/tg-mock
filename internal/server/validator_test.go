// internal/server/validator_test.go
package server

import (
	"testing"

	"github.com/watzon/tg-mock/gen"
)

func TestValidateRequest(t *testing.T) {
	v := NewValidator()

	tests := []struct {
		name    string
		method  string
		params  map[string]interface{}
		wantErr bool
	}{
		{
			name:   "sendMessage valid",
			method: "sendMessage",
			params: map[string]interface{}{
				"chat_id": 123,
				"text":    "Hello",
			},
			wantErr: false,
		},
		{
			name:   "sendMessage missing required",
			method: "sendMessage",
			params: map[string]interface{}{
				"chat_id": 123,
			},
			wantErr: true,
		},
		{
			name:    "getMe no params required",
			method:  "getMe",
			params:  map[string]interface{}{},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spec := gen.Methods[tt.method]
			err := v.Validate(spec, tt.params)
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

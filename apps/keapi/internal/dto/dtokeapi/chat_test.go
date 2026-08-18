package dtokeapi

import (
	"testing"

	"github.com/insmtx/corekg/apps/kellm/models/kellmtype"
	"github.com/ygpkg/yg-go/apis/apiobj"
)

func TestValidChatCompletionsEmptyMessagesRequiresSession(t *testing.T) {
	tests := []struct {
		name      string
		sessionID uint
		messages  []kellmtype.Message
		wantValid bool
		wantMsg   string
	}{
		{
			name:      "empty messages with session",
			sessionID: 123,
			wantValid: true,
		},
		{
			name:      "empty messages without session",
			wantValid: false,
			wantMsg:   "keapi_empty_messages",
		},
		{
			name:      "non-empty messages still require content",
			sessionID: 123,
			messages:  []kellmtype.Message{{Role: "user"}},
			wantValid: false,
			wantMsg:   "keapi_empty_messages",
		},
		{
			name: "non-empty messages with content",
			messages: []kellmtype.Message{{
				Role: "user",
				Content: kellmtype.MessageContent{
					Text: "hello",
				},
			}},
			wantValid: true,
		},
		{
			name: "system message alone is not valid user content",
			messages: []kellmtype.Message{{
				Role: "system",
				Content: kellmtype.MessageContent{
					Text: "answer in Chinese",
				},
			}},
			wantValid: false,
			wantMsg:   "keapi_empty_messages",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := &ChatCompletionsRequest{}
			req.Request.SessionID = tt.sessionID
			req.Request.Messages = tt.messages
			resp := &apiobj.BaseResponse{}

			if got := req.ValidChatCompletions(resp); got != tt.wantValid {
				t.Fatalf("ValidChatCompletions() = %v, want %v", got, tt.wantValid)
			}
			if resp.Message != tt.wantMsg {
				t.Fatalf("resp.Message = %q, want %q", resp.Message, tt.wantMsg)
			}
		})
	}
}

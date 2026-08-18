package agent

import (
	"strings"
	"testing"

	"github.com/cloudwego/eino/schema"
)

func TestMarshalDebugJSONPrefixesMarshalError(t *testing.T) {
	got := marshalDebugJSON(make(chan int))
	if !strings.HasPrefix(got, "marshal_error:") {
		t.Fatalf("expected marshal_error prefix, got %q", got)
	}
}

func TestNextModelInputLogStartLogsChangedSuffix(t *testing.T) {
	agent := &ReActAgent{}
	systemMsg := schema.SystemMessage("system")
	userMsg := schema.UserMessage("question")
	nextStepMsg := schema.UserMessage("next step")

	if got := agent.nextModelInputLogStart([]*schema.Message{systemMsg, userMsg}); got != 0 {
		t.Fatalf("first round should log from 0, got %d", got)
	}
	if got := agent.nextModelInputLogStart([]*schema.Message{systemMsg, userMsg, nextStepMsg}); got != 2 {
		t.Fatalf("second round should skip unchanged prefix, got %d", got)
	}

	toolMsg := schema.ToolMessage("tool result", "tool-call-1", schema.WithToolName("forest_search_tool"))
	current := []*schema.Message{systemMsg, userMsg, toolMsg, nextStepMsg}
	if got := agent.nextModelInputLogStart(current); got != 2 {
		t.Fatalf("changed middle message should log changed suffix, got %d", got)
	}
	if got := agent.nextModelInputLogStart(current); got != len(current) {
		t.Fatalf("unchanged round should skip all messages, got %d", got)
	}
}

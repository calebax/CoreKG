package parser

import (
	"strings"
	"testing"
)

func TestJSONParserParse(t *testing.T) {
	inst := NewJSONParser()

	input1 := "让我查询一下\n```\n{\"name\": \"search\", \"arguments\": {\"query\": \"test\"}}\n```\n结果如下"
	input2 := `{"name": "get_weather", "arguments": "{\"location\": \"北京\"}"}`

	toolCalls1, content1, err := inst.Parse(input1)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(toolCalls1) != 1 {
		t.Fatalf("toolCalls1 len = %d, want 1", len(toolCalls1))
	}
	if toolCalls1[0].Name != "search" {
		t.Fatalf("toolCalls1[0].Name = %q, want search", toolCalls1[0].Name)
	}
	if content1 != "让我查询一下\n\n结果如下" && content1 != "让我查询一下\n结果如下" && content1 != "让我查询一下 结果如下" {
		t.Fatalf("content1 = %q, want remaining non-JSON text", content1)
	}

	toolCalls2, content2, err := inst.Parse(input2)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(toolCalls2) != 1 {
		t.Fatalf("toolCalls2 len = %d, want 1", len(toolCalls2))
	}
	if toolCalls2[0].Name != "get_weather" {
		t.Fatalf("toolCalls2[0].Name = %q, want get_weather", toolCalls2[0].Name)
	}
	if content2 != "" {
		t.Fatalf("content2 = %q, want empty", content2)
	}
}

func TestStreamJSONParser(t *testing.T) {
	inst := NewStreamJSONParser()

	chunks := []string{
		"让我查询一下天气，",
		"先调用工具。\n```json\n",
		"{\"name\": \"get_weather\", ",
		"\"arguments\": \"{\\\"location\\\": \\\"西安\\\"}\"}",
		"\n```\n处理完毕",
	}

	var allCalls []ToolCall
	var emitted string
	for _, chunk := range chunks {
		calls, output, _ := inst.Add(chunk)
		allCalls = append(allCalls, calls...)
		emitted += output
	}

	flushCalls, remaining := inst.Flush()
	allCalls = append(allCalls, flushCalls...)
	emitted += remaining

	if len(allCalls) != 1 {
		t.Fatalf("allCalls len = %d, want 1", len(allCalls))
	}
	if allCalls[0].Name != "get_weather" {
		t.Fatalf("allCalls[0].Name = %q, want get_weather", allCalls[0].Name)
	}
	if !strings.Contains(emitted, "让我查询一下天气") && emitted != "处理完毕" {
		t.Fatalf("emitted = %q, want buffered plain text", emitted)
	}
}

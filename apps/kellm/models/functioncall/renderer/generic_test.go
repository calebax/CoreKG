package renderer

import (
	"strings"
	"testing"

	"github.com/insmtx/corekg/apps/kellm/models/kellmtype"
)

func TestGenericRendererRenderToolCalls(t *testing.T) {
	inst := &GenericRenderer{}
	messages := []kellmtype.Message{
		{
			Role: "user",
			Content: kellmtype.MessageContent{
				Text: "今天西安天气怎么样？",
			},
		},
	}
	tools := []kellmtype.Tool{
		{
			Type: "function",
			Function: kellmtype.Function{
				Name:        "get_weather",
				Description: "获取指定城市的天气信息",
				Parameters: kellmtype.JSONSchema{
					"type": "object",
					"properties": map[string]any{
						"location": map[string]any{
							"type":        "string",
							"description": "城市名称，例如：北京、上海",
						},
						"unit": map[string]any{
							"type":        "string",
							"description": "温度单位",
							"enum":        []string{"celsius", "fahrenheit"},
						},
					},
					"required": []string{"location"},
				},
			},
		},
	}

	prompt, err := inst.Render(messages, tools)
	if err != nil {
		t.Fatalf("Render failed: %v", err)
	}
	if !strings.Contains(prompt, "### 错误示例（禁止这样做）") {
		t.Fatalf("prompt missing error example section: %q", prompt)
	}
	if !strings.Contains(prompt, "### 完整格式模板") {
		t.Fatalf("prompt missing full template section: %q", prompt)
	}
	if !strings.Contains(prompt, `\"`) {
		t.Fatalf("prompt missing escaped quote guidance: %q", prompt)
	}
}

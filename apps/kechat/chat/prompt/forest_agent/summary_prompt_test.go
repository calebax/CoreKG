package prompt

import (
	"strings"
	"testing"
)

func TestBuildForestAgentSummarySystemPromptReferenceEnabled(t *testing.T) {
	got := BuildForestAgentSummarySystemPrompt(SummaryPromptOptions{
		EnableReference: true,
	})

	if !strings.Contains(got, "## 七、来源引用标签规则") {
		t.Fatal("expected reference generation rules to be injected")
	}
	if !strings.Contains(got, "引用白名单") {
		t.Fatal("expected reference whitelist rules to be injected")
	}
	if !strings.Contains(got, "{{.taskHistory}}") {
		t.Fatal("expected task history template to be preserved")
	}
}

func TestBuildForestAgentSummarySystemPromptReferenceDisabled(t *testing.T) {
	got := BuildForestAgentSummarySystemPrompt(SummaryPromptOptions{
		EnableReference: false,
	})

	for _, forbidden := range []string{
		"## 七、来源引用标签规则",
		"引用白名单",
		"{Reference",
		"file_id",
		"sequence",
		"chunk_id",
		"数字fileID",
	} {
		if strings.Contains(got, forbidden) {
			t.Fatalf("disabled reference prompt should not contain %q", forbidden)
		}
	}
	if !strings.Contains(got, "当前未启用来源标记能力") {
		t.Fatal("expected disabled output mode to be injected")
	}
	if strings.Contains(got, "不得输出任何 {Reference ...} 标签") {
		t.Fatal("expected reference generation rules not to be injected")
	}
}

func TestBuildForestAgentSummarySystemPromptExtraPromptWrapped(t *testing.T) {
	got := BuildForestAgentSummarySystemPrompt(SummaryPromptOptions{
		EnableReference: true,
		ExtraPrompt:     "请使用简洁风格。",
	})

	if !strings.Contains(got, "<custom_summary_prompt>\n请使用简洁风格。\n</custom_summary_prompt>") {
		t.Fatal("expected extra prompt to be wrapped")
	}
	if !strings.Contains(got, "以下内容是用户侧补充要求，适用边界如下") {
		t.Fatal("expected extra prompt guardrails to be included")
	}
}

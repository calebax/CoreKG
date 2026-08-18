package file

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/stretchr/testify/assert"
	"github.com/ygpkg/yg-go/logs"
)

func TestAnalysisFileTool(t *testing.T) {
	ctx := context.Background()

	aiKey := ""

	if len(aiKey) == 0 {
		t.Skip("DS_KEY not set, skipping TestGenCode")
	}

	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		APIKey:  aiKey,
		Model:   "deepseek-v3",
		Timeout: 300 * time.Second,
		BaseURL: "https://api.example.com/v3/llm.chat",
	})

	tool, err := NewAnalysisFileTool(ctx, &AnalysisConfig{
		Model: chatModel,
	})
	assert.NoError(t, err)

	req := &AnalysisRequest{
		FileName: "sample.txt",
		FileURL:  "https://example.com:58081/dotpen/analyser_md/251230/f7882e45693a45f782ce3ac774d6cf25/content.md",
		Focus:    "提取关键数据",
	}
	reqBody, _ := json.Marshal(req)

	out, err := tool.InvokableRun(ctx, string(reqBody))
	assert.NoError(t, err)

	var resp AnalysisResponse
	assert.NoError(t, json.Unmarshal([]byte(out), &resp))

	jsonResp, _ := json.Marshal(resp)
	logs.InfoContextf(ctx, "resp: %s", string(jsonResp))

}

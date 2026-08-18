package reranksearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/models/chatagent"
	"github.com/insmtx/corekg/apps/kechat/models/chatclient"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/ygpkg/yg-go/apis/constants"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
)

// GetRerank rerank
func GetRerank(ctx context.Context, question string, docs []string) (*RerankResponse, error) {
	cfg, err := GetRerankConfig(ctx)
	if err != nil {
		logs.ErrorContextf(ctx, "get rerank config err:%v", err)
		return nil, err
	}
	emptyDocCount := 0
	for _, doc := range docs {
		if strings.TrimSpace(doc) == "" {
			emptyDocCount++
		}
	}
	logs.InfoContextf(ctx, "GetRerank request stats: question=%q, docs_len=%d, empty_doc_count=%d", question, len(docs), emptyDocCount)
	reqbody := map[string]interface{}{
		"model":  cfg.ModelName,
		"text_1": question,
		"text_2": docs,
	}
	bodyBytes, err := json.Marshal(reqbody)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetRerank] marshal request body error: %v", err)
		return nil, fmt.Errorf("GetRerank json marshal error: %v", err)
	}
	// 创建请求
	req, err := http.NewRequest("POST", cfg.Url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		logs.ErrorContextf(ctx, "[GetRerank] create request error: %v", err)
		return nil, fmt.Errorf("GetRerank create request error: %v", err)
	}
	// 设置 Header
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.Key)
	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetRerank] do request error: %v", err)
		return nil, fmt.Errorf("GetRerank do request error: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logs.ErrorContextf(ctx, "[GetRerank] request failed with status code: %v, err: %v", resp.StatusCode, string(body))
		return nil, fmt.Errorf("GetRerank request status code:%v, err: %v", resp.StatusCode, string(body))
	}
	var respBody *RerankResponse
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return nil, err
	}
	if respBody.Data == nil {
		logs.WarnContextf(ctx, "[DEBUG][chunk-empty] GetRerank: response data is nil, empty rerank result")
		return nil, fmt.Errorf("GetRerank no rerank data")
	}
	logs.InfoContextf(ctx, "[DEBUG][chunk-empty] GetRerank success: input_docs=%d, output_data=%d", len(docs), len(respBody.Data))
	return respBody, nil
}

// RerankResponse 定义响应结构
type RerankResponse struct {
	Data []Rerank `json:"data"`
}

type Rerank struct {
	Index  int     `json:"index"`
	Score  float64 `json:"score"`
	Object string  `json:"object"`
}

func initEbConfig(ctx context.Context) error {
	cfg := &RerankConfig{}
	err := settings.GetYaml("knowledge", "rerank", cfg)
	if err != nil {
		logs.ErrorContextf(ctx, "read rerank config error:%v", err)
		return err
	}
	rerankConfig = cfg
	return nil
}

var rerankConfig *RerankConfig

func GetRerankConfig(ctx context.Context) (*RerankConfig, error) {
	if rerankConfig == nil {
		err := initEbConfig(ctx)
		if err != nil {
			logs.ErrorContextf(ctx, "get rerankconfig err:%v", err)
			return nil, err
		}
	}
	return rerankConfig, nil
}

type RerankConfig struct {
	Url       string `yaml:"url"`
	Key       string `yaml:"key"`
	ModelName string `yaml:"model_name"`
}

// UserQueryRewrite 问题改写
func UserQueryRewrite(ctx context.Context, question string) (string, error) {
	req := &chattype.ChatRequestBody{
		Stream: false,
		Model:  chatagent.GetAgentI18nName(ctx, ctx.Value(constants.CtxKeyLang).(string), global.ChatAgentUserQueryRewrite),
		ChatOptions: chattype.ChatOptions{
			Input: []chattype.Input{
				{Name: "input1", Value: question},
			},
		},
	}
	ginctx := &gin.Context{}
	w, err := chatclient.NewInternalChat(ginctx, ctx.Value(constants.CtxKeyRequestID).(string), "", 1, req)
	if err != nil {
		logs.ErrorContextf(ctx, "failed to create internal chat: %v", err)
		return "", err
	}
	res, err := w.AgentChatInternal(nil)
	if err != nil {
		logs.ErrorContextf(ctx, "agent chat error: %v", err)
		return "", err
	}
	return res.Content, nil
}

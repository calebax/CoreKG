package essearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/insmtx/corekg/apps/ketask/models/ragtypes"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
)

// const url = "https://example.com:53081/v1/embeddings"

var embeddingConfig *EmbeddingConfig

// GetEmbedding 获取一段话的embeding值
func GetEmbedding(question string) (ragtypes.Embedding, error) {
	reqbody := map[string]interface{}{
		"model": embeddingConfig.ModelName,
		"input": question,
	}
	bodyBytes, err := json.Marshal(reqbody)
	if err != nil {
		return nil, err
	}
	// 创建请求
	req, err := http.NewRequest("POST", embeddingConfig.Url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	// 设置 Header
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+embeddingConfig.Key)
	// 发送请求
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("request status code:%v", resp.StatusCode)
	}
	// 解析响应
	var respBody EmbeddingData
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return nil, err
	}
	if respBody.Data == nil {
		return nil, fmt.Errorf("no embedding data")
	}
	// 目前只有一个只计算一个句子的embedding
	if len(respBody.Data[0].Embedding) == 0 {
		return nil, fmt.Errorf("no embedding data")
	}
	return respBody.Data[0].Embedding, nil
}

func InitEbConfig(ctx context.Context) error {
	cfg := &EmbeddingConfig{}
	err := settings.GetYaml("knowledge", "embedding", cfg)
	if err != nil {
		logs.ErrorContextf(ctx, "read embedding config error:%v", err)
		return err
	}
	embeddingConfig = cfg
	return nil
}

type EmbeddingData struct {
	Data []struct {
		Index     int                `json:"index"`
		Object    string             `json:"object"`
		Embedding ragtypes.Embedding `json:"embedding"`
	} `json:"data"`
}

type EmbeddingConfig struct {
	Url       string `yaml:"url"`
	Key       string `yaml:"key"`
	ModelName string `yaml:"model_name"`
}

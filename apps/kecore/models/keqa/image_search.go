package keqa

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"path/filepath"

	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
	"github.com/insmtx/corekg/pkgs/agentclient"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/pkgs/utils"
	"github.com/ygpkg/yg-go/logs"
)

// ImageQueryFile 图片搜索
func ImageQueryFile(ctx context.Context, uin, forestID uint, url string) ([]*foresttype.KnownowForestFile, error) {
	forestInfo, err := forest.GetForestByID(ctx, forestID)
	if err != nil {
		logs.ErrorContextf(ctx, "get forest info failed: %v", err)
		return nil, err
	}
	fileIds, err := ImageSearchFile(ctx, forestInfo, url)
	if err != nil {
		logs.ErrorContextf(ctx, "image search failed: %v", err)
		return nil, err
	}
	if fileIds == nil {
		logs.WarnContextf(ctx, "file ids is nil")
		return nil, nil
	}

	// ========================= filter ban list ==========================
	result, err := forest.NewAccessProvider(ctx, &forest.ContextModel{
		ResourceType: foresttype.ResourceTypeForestFile,
		ScopeType:    foresttype.ScopeTypeUser,
		ScopeID:      uin,
		Action:       foresttype.ActionBan,
	}).Action(ctx)
	if err != nil {
		logs.ErrorContextf(ctx, "filter ban list failed: %v", err)
		return nil, err
	}
	var fIDs []uint

	banSet := utils.ToMap(result.BanList, func(v uint) uint { return v })
	for _, v := range fileIds {
		if _, ok := banSet[v]; ok {
			continue
		}
		fIDs = append(fIDs, v)
	}

	// ========================== filter ban list ==========================

	fileList, err := forest.GetForestFileByIDs(fIDs)
	if err != nil {
		logs.ErrorContextf(ctx, "get file list failed: %v", err)
		return nil, err
	}
	return fileList, nil
}

// ImageSearchFile 搜索对应知识森林文件
func ImageSearchFile(ctx context.Context, forest_info *foresttype.KnownowForest, url string) ([]uint, error) {
	// 图片调用多模态
	image_content, err := DoImageParseRequest(ctx, url)
	if err != nil {
		logs.ErrorContextf(ctx, "[image_search] [DoImageParseRequest] error: %v", err)
		return nil, err
	}
	wrapper, err := essearch.NewEsSearchWrapper(ctx, forest_info.EsIndex(), image_content, []uint{forest_info.ID}, nil)
	if err != nil {
		logs.ErrorContextf(ctx, "[image_search] [NewEsSearchWrapper] error: %v", err)
		return nil, err
	}
	esdata, err := wrapper.SearchQuestionChunk()
	if err != nil {
		logs.ErrorContextf(ctx, "[image_search] [SearchQuestionChunk] error: %v", err)
		return nil, err
	}
	file_ids := []uint{}
	for _, hit := range esdata.Hits.Hits {
		file_ids = append(file_ids, hit.Source.FileID)
	}
	return file_ids, err
}

// DoImageParseRequest 多模态解析
func DoImageParseRequest(ctx context.Context, url string) (string, error) {
	cfg, err := agentclient.GetLLMConfig(ctx, global.SettingGroupKnowledge, global.SettingKeyLlmImageParse)
	if err != nil {
		logs.ErrorContextf(ctx, "[image_search] [GetLLMConfig] error: %v", err)
		return "", err
	}
	image_base64, err := ImageUrlTOBase64(url)
	if err != nil {
		logs.ErrorContextf(ctx, "[image_parse] [image_url_to_base64] error: %v", err)
		return "", err
	}
	requestBody := map[string]interface{}{
		"model": cfg.ModelName,
		"messages": []interface{}{
			map[string]interface{}{
				"role": "user",
				"content": []interface{}{
					map[string]string{"type": "text", "text": "帮我详细描述图片中的内容"},
					map[string]interface{}{
						"type": "image_url",
						"image_url": map[string]string{
							"url": image_base64,
						},
					},
				},
			},
		},
		"stream": false,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("POST", cfg.BaseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+cfg.APIKEY)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logs.ErrorContextf(ctx, "Request err:", err)
		return "", err
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logs.ErrorContextf(ctx, "Received non-OK HTTP status: %s, body: %s", resp.Status, string(body))
		return "", fmt.Errorf("received non-OK HTTP status: %s", resp.Status)
	}
	var res agentclient.ChatResponseBody
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		logs.ErrorContextf(ctx, "Unmarshal response err:%v", err)
		return "", err
	}

	if len(res.Choices) == 0 {
		return "", fmt.Errorf("no choices found")
	}
	if len(res.Choices[0].Message.Content) <= 0 {
		return "", fmt.Errorf("no content found")
	}

	return res.Choices[0].Message.Content, nil
}

func ImageUrlTOBase64(url string) (string, error) {
	// 发送 HTTP 请求下载文件
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	// 读取响应体
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	// 转换为 base64
	encoded := base64.StdEncoding.EncodeToString(data)
	en_url := "data:" + mime.TypeByExtension(filepath.Ext(url)) + ";base64," + encoded
	return en_url, nil
}

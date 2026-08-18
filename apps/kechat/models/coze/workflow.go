package coze

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/juju/errors"

	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
)

const (
	WorkflowChatUrl = "/v1/workflow/ygrun"
)

// WorkflowChat 获取coze页面ID
func WorkflowChat(ctx context.Context, token, workflowID string, parameters chattype.InputList) (*WorkflowChatResponse, error) {
	cozeUrl, err := settings.GetText("corekg", "coze_url")
	if err != nil {
		logs.ErrorContextf(ctx, "get coze url err %s", err.Error())
		return nil, err
	}
	// cozeUrl = "http://localhost:8888"
	// 2. 创建请求（使用同一个 client，自动携带 Cookie）
	pluginURL, err := url.JoinPath(cozeUrl, WorkflowChatUrl)
	if err != nil {
		logs.ErrorContextf(ctx, "join coze pluginURL url err %s", err.Error())
		return nil, err
	}
	params := map[string]interface{}{}
	for _, v := range parameters {
		params[v.Name] = v.Value
	}

	reqBody := map[string]interface{}{
		"workflow_id": workflowID,
		"is_async":    false,
		"parameters":  params,
	}
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		logs.ErrorContextf(ctx, "marshal request body error, %s", err.Error())
		return nil, err
	}

	req, err := http.NewRequest("POST", pluginURL, bytes.NewBuffer(jsonData))
	if err != nil {
		logs.ErrorContextf(ctx, "create request error, %s", err.Error())
		return nil, err
	}

	// 3. 设置必要请求头（保持与登录时一致的浏览器环境）
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	// req.Header.Set("Cookie", "i18next=zh-CN; session_key="+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		body, _ := io.ReadAll(resp.Body)
		logs.ErrorContextf(ctx, "client.Do error, %s,body:%s", err.Error(), string(body))
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("es query failed: %v, error: %s", resp.StatusCode, string(body))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		logs.ErrorContextf(ctx, "error reading body: %v", err)
		return nil, err
	}

	logs.InfoContextf(ctx, "workflow body %s", string(body))
	// 解析JSON响应
	var res WorkflowChatResponse
	if err = json.Unmarshal(body, &res); err != nil {
		logs.ErrorContextf(ctx, "unmarshal WorkflowChatResponse error: %v", err)
		return nil, err
	}

	if res.Code != 0 {
		logs.ErrorContextf(ctx, "workflow code: %d,msg: %s", res.Code, res.Msg)
		return nil, fmt.Errorf("workflow code: %d,msg: %s", res.Code, res.Msg)
	}

	extractAndReplaceData(&res)

	return &res, nil
}

type WorkflowChatResponse struct {
	Code  int    `json:"code"`
	Msg   string `json:"msg"`
	Data  string `json:"data"`
	Token int    `json:"token"`
	// DataStruct WorkflowChatData `json:"data_struct"`
}

type WrappedOutput struct {
	ContentType  int    `json:"content_type"`
	Data         string `json:"data"`
	TypeForModel int    `json:"type_for_model"`
}

func extractAndReplaceData(res *WorkflowChatResponse) error {
	// 尝试解析为包装格式
	var wrapped WrappedOutput
	if err := json.Unmarshal([]byte(res.Data), &wrapped); err != nil {
		return nil
	}

	// 检查是否为预期coze OpenAPIRunFlow返回的包装格式，返回文本的情况 提取对应data
	if wrapped.ContentType == 1 && wrapped.TypeForModel == 2 {
		// 用内部的 data 替换外层的 data
		res.Data = wrapped.Data
	}

	return nil
}

type WorkflowChatData struct {
	Data         string `json:"data"`
	TypeForModel int    `json:"type_for_model"`
	ContentType  int    `json:"content_type"`
}

type Workflowitem struct {
	WorkflowID string `json:"workflow_id"`
	SpaceID    string `json:"space_id"`
	ExecuteID  string `json:"execute_id"`
	Version    string `json:"version"`
}

type CozeWorkflowCanvasResp struct {
	Data struct {
		Workflow struct {
			SchemaJSON string `json:"schema_json"`
		} `json:"workflow"`
	} `json:"data"`
	Code int `json:"code"`
}

type WorkflowSchema struct {
	Nodes []struct {
		Data struct {
			NodeMeta struct {
				Title string `json:"title"`
			} `json:"nodeMeta"`
			TriggerParameters []WorkflowField `json:"outputs"`
		} `json:"data"`
	} `json:"nodes"`
}

type WorkflowField struct {
	Type     string `json:"type"`
	Name     string `json:"name"`
	Required bool   `json:"required"`
}

func (w *Workflowitem) GetWorkflowCanvas(ctx context.Context, cozeUrl, token string) (wf []WorkflowField, err error) {
	payload := strings.NewReader(`
{
    "space_id": "` + w.SpaceID + `",
    "workflow_id": "` + w.WorkflowID + `"
}
`)

	pluginURL := cozeUrl + "/api/workflow_api/canvas"
	req, err := http.NewRequest("POST", pluginURL, payload)
	if err != nil {
		logs.ErrorContextf(ctx, "create request error, %v", err)
		return
	}

	// 3. 设置必要请求头（保持与登录时一致的浏览器环境）
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logs.ErrorContextf(ctx, "client.Do error, %v", err)
		return
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("es query failed: %v, error: %s", resp.StatusCode, string(body))
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println(err)
		return
	}
	var result CozeWorkflowCanvasResp
	if err = json.Unmarshal(body, &result); err != nil {
		logs.ErrorContextf(ctx, "json.NewDecoder error:  %v", err)
		return
	}
	if result.Code != 0 {
		logs.ErrorContextf(ctx, "API: %s \n resp: %v", pluginURL, result)
	}
	// 解析JSON响应
	var schema WorkflowSchema
	if err = json.Unmarshal([]byte(result.Data.Workflow.SchemaJSON), &schema); err != nil {
		logs.ErrorContextf(ctx, "unmarshal WorkflowSchema error: %v", err)
		return
	}
	for _, v := range schema.Nodes {
		if v.Data.NodeMeta.Title == "开始" {
			wf = v.Data.TriggerParameters
		}
	}
	return wf, nil
}

type WorkflowTestRunResp struct {
	Code int `json:"code"`
	Data struct {
		ExecuteID string `json:"execute_id"`
	} `json:"data"`
}

type WorkflowTestRunField struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func (w *Workflowitem) WorkflowTestRun(ctx context.Context, cozeUrl, token string, fields []WorkflowTestRunField) error {
	fieldStr := ""
	for i, v := range fields {
		if i != len(fields)-1 {
			fieldStr = fieldStr + "\"" + v.Name + "\": \"" + v.Value + "\", \n"
		} else {
			fieldStr += fmt.Sprintf("\"%s\": \"%s\"", v.Name, v.Value)
		}
	}
	payload := strings.NewReader(`
{
    "workflow_id": "` + w.WorkflowID + `",
    "input": {
        ` + fieldStr + `
    },
    "space_id": "` + w.SpaceID + `",
    "commit_id": ""
}
`)

	// 2. 创建请求（使用同一个 client，自动携带 Cookie）
	pluginURL := cozeUrl + "/api/workflow_api/test_run"
	req, err := http.NewRequest("POST", pluginURL, payload)
	if err != nil {
		logs.ErrorContextf(ctx, "create request error, %v", err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logs.ErrorContextf(ctx, "client.Do error, %v", err)
		return err
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logs.ErrorContextf(ctx, "client.Do error, %s", string(body))
		return err
	}
	defer resp.Body.Close()
	var cozeResp WorkflowTestRunResp
	if err = json.NewDecoder(resp.Body).Decode(&cozeResp); err != nil {
		logs.ErrorContextf(ctx, "json.NewDecoder error:  %v", err)
		return err
	}
	if cozeResp.Code != 0 {
		logs.ErrorContextf(ctx, "API: %s \n resp: %v", pluginURL, cozeResp)
		return errors.New("workflow test run failed, code: " + strconv.Itoa(cozeResp.Code))
	}
	w.ExecuteID = cozeResp.Data.ExecuteID
	return nil
}

type WorkflowGetProcessResp struct {
	Code int `json:"code"`
	Data struct {
		Status []struct {
			NodeType  string `json:"NodeType"`
			RawOutput string `json:"raw_output"`
		} `json:"nodeResults"`
	} `json:"data"`
}

func (w *Workflowitem) WorkflowGetProcess(ctx context.Context, cozeUrl, token string) (output string, code int, err error) {
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	maxWaitTime := 5 * time.Minute
	deadline := time.Now().Add(maxWaitTime)
	url := cozeUrl + "/api/workflow_api/get_process?workflow_id=" + w.WorkflowID + "&space_id=" + w.SpaceID + "&execute_id=" + w.ExecuteID + "&need_async=true"
	var (
		req *http.Request
		res *http.Response
	)
	for {
		if time.Now().After(deadline) {
			return "", 0, errors.New("workflow execution timeout after " + maxWaitTime.String())
		}
		select {
		case <-ctx.Done():
			return "", 0, ctx.Err()
		default:
		}
		req, err = http.NewRequest("GET", url, nil)
		if err != nil {
			logs.ErrorContextf(ctx, "create request error, %v", err)
			return "", 0, err
		}
		req.Header.Set("Authorization", "Bearer "+token)
		res, err = client.Do(req)
		if err != nil {
			logs.ErrorContextf(ctx, "client.Do error, %v", err)
			return
		}
		if res.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(res.Body)
			logs.ErrorContextf(ctx, "workflow get process status not ok,  %s", string(body))
			return "", 0, fmt.Errorf("es query failed: %v, error: %s", res.StatusCode, string(body))
		}

		var cozeResp WorkflowGetProcessResp
		if err = json.NewDecoder(res.Body).Decode(&cozeResp); err != nil {
			logs.ErrorContextf(ctx, "json.NewDecoder error:  %v", err)
			return
		}
		if cozeResp.Code != 0 {
			logs.ErrorContextf(ctx, "API: %s \n resp: %v", url, cozeResp)
			return "", cozeResp.Code, nil
		}

		res.Body.Close()

		for _, v := range cozeResp.Data.Status {
			if v.NodeType == "End" && v.RawOutput != "" {
				// var data map[string]string
				// if err = json.Unmarshal([]byte(v.RawOutput), &data); err != nil {
				// 	logs.ErrorContextf(ctx, "parse RawOutput failed: %v", err)
				// 	return
				// }
				// var lastValue string
				// for _, val := range data {
				// 	lastValue = val
				// }
				return v.RawOutput, 0, nil
			}
		}
		time.Sleep(time.Second)
	}
}

type WorkflowPublishResp struct {
	Code int `json:"code"`
	Data struct {
		Success bool `json:"success"`
	} `json:"data"`
}

func (w *Workflowitem) WorkflowPublish(ctx context.Context, cozeUrl, token, version string) (success bool, err error) {
	version, err = IncreaseVersion(version)
	if err != nil {
		logs.ErrorContextf(ctx, "increase version error, %v", err)
		return false, err
	}
	payload := strings.NewReader(`
{
  "workflow_id": "` + w.WorkflowID + `",
  "space_id": "` + w.SpaceID + `",
  "has_collaborator": false,
  "force": true,
  "workflow_version": "` + version + `",
  "version_description": "` + version + `"
}
`)

	pluginURL := cozeUrl + "/api/workflow_api/publish"
	req, err := http.NewRequest("POST", pluginURL, payload)
	if err != nil {
		logs.ErrorContextf(ctx, "create request error, %v", err)
		return false, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logs.ErrorContextf(ctx, "client.Do error, %v", err)
		return false, err
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logs.ErrorContextf(ctx, "client.Do error, %s", string(body))
		return false, err
	}
	defer resp.Body.Close()
	var cozeResp WorkflowPublishResp
	if err = json.NewDecoder(resp.Body).Decode(&cozeResp); err != nil {
		logs.ErrorContextf(ctx, "json.NewDecoder error:  %v", err)
		return false, err
	}
	if cozeResp.Code != 0 {
		logs.ErrorContextf(ctx, "API: %s \n resp: %v", pluginURL, cozeResp)
		return false, errors.New("workflow test run failed, code: " + strconv.Itoa(cozeResp.Code))
	}
	success = cozeResp.Data.Success
	w.Version = version
	return success, nil
}

func IncreaseVersion(version string) (string, error) {
	v := strings.TrimPrefix(version, "v")

	parts := strings.Split(v, ".")
	if len(parts) == 0 {
		return "", fmt.Errorf("invalid version format: %s", version)
	}

	lastIndex := len(parts) - 1
	num, err := strconv.Atoi(parts[lastIndex])
	if err != nil {
		return "", fmt.Errorf("invalid version number: %s", parts[lastIndex])
	}
	num++
	parts[lastIndex] = strconv.Itoa(num)

	newVersion := "v" + strings.Join(parts, ".")
	return newVersion, nil
}

// CreateCozeWorkflowAPI 创建coze工作流
func CreateCozeWorkflowAPI(ctx *gin.Context, desc, name, spaceID, token, cozeUrl string) (string, error) {

	payload := strings.NewReader(`
{
    "desc": "` + desc + `",
    "flow_mode": 0,
    "icon_uri": "default_icon/default_workflow_icon.png",
	"name": "` + name + `",
    "space_id": "` + spaceID + `"
}
`)

	pluginURL := cozeUrl + "/api/workflow_api/create"
	req, err := http.NewRequest("POST", pluginURL, payload)
	if err != nil {
		logs.ErrorContextf(ctx, "create request error, %s", err)
		return "", err
	}

	// 3. 设置必要请求头（保持与登录时一致的浏览器环境）
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logs.ErrorContextf(ctx, "client.Do error, %s", err)
		return "", err
	}
	defer resp.Body.Close()
	var wkResp CreateCozeWorkflowAPIResponse
	if err = json.NewDecoder(resp.Body).Decode(&wkResp); err != nil {
		logs.ErrorContextf(ctx, "json.NewDecoder error:  %s", err)
		return "", err
	}
	if wkResp.Code != 0 {
		response, _ := io.ReadAll(resp.Body)
		logs.ErrorContextf(ctx, "API: %s \n resp: %v", pluginURL, string(response))
	}
	return wkResp.Data.WorkflowID, nil
}

// DeleteCozeWorkflow 删除coze工作流
func DeleteCozeWorkflow(ctx *gin.Context, workflowID, spaceID, token, cozeUrl string) error {

	payload := strings.NewReader(`
{
    "action": 1,
	"workflow_id": "` + workflowID + `",
	"space_id": "` + spaceID + `"
}
`)

	pluginURL := cozeUrl + "/api/workflow_api/delete"
	req, err := http.NewRequest("POST", pluginURL, payload)
	if err != nil {
		logs.ErrorContextf(ctx, "create request error, %s", err)
		return err
	}

	// 3. 设置必要请求头（保持与登录时一致的浏览器环境）
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logs.ErrorContextf(ctx, "client.Do error, %s", err)
		return err
	}
	defer resp.Body.Close()
	body := resp.Body
	var dwResp DeleteCozeWorkflowResponse
	if err = json.NewDecoder(body).Decode(&dwResp); err != nil {
		logs.ErrorContextf(ctx, "json.NewDecoder error:  %s", err)
		return err
	}
	if dwResp.Code != 0 {
		response, _ := io.ReadAll(body)
		logs.WarnContextf(ctx, "API: %s \n req: %v", pluginURL, payload)
		logs.WarnContextf(ctx, "API: %s \n resp: %v", pluginURL, string(response))
		return errors.New("deleted coze workflow failed")
	}
	return nil
}

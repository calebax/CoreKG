package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/insmtx/corekg/pkgs/task"
	"github.com/ygpkg/yg-go/logs"
)

// joinURL joins a base URL with a path, ensuring proper formatting.
func joinURL(baseURL, path string) string {
	return strings.TrimSuffix(baseURL, "/") + "/" + strings.TrimPrefix(path, "/")
}

// GetPendingTestResponse represents the response structure for pending tasks.
type GetPendingTestResponse struct {
	Response struct {
		TaskID  uint   `json:"task_id"`
		Payload string `json:"payload"` // 任务内容
	}
}

// GetPendingTask retrieves a pending task for the worker.
func GetPendingTask(ctx context.Context, taskPayload interface{}) (uint, error) {
	url := joinURL(baseURL, "v3/knowledge.GetPendingTask")
	// 构造请求数据
	reqData := map[string]interface{}{
		"Request": map[string]interface{}{
			"task_type": taskType,
			"worker_id": workerID,
		},
	}

	reqBody, err := json.Marshal(reqData)
	if err != nil {
		logs.ErrorContextf(ctx, "Error marshalling JSON: %v", err)
		return 0, err
	}

	// 创建请求
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		logs.ErrorContextf(ctx, "Error making POST request: %v", err)
		return 0, err
	}
	defer resp.Body.Close()

	var response GetPendingTestResponse
	// 解析 JSON
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		logs.ErrorContextf(ctx, "Error decoding JSON response[%+v]: %v", resp, err)
		return 0, err
	}
	if resp.StatusCode != http.StatusOK {
		logs.ErrorContextf(ctx, "Received non-OK HTTP status: %s", resp.Status)
		return 0, fmt.Errorf("received non-OK HTTP status: %s", resp.Status)
	}
	// 检查任务ID是否为0
	if response.Response.TaskID == 0 {
		logs.InfoContextf(ctx, "No pending task found for worker %v", workerID)
		return 0, nil
	}
	err = json.Unmarshal([]byte(response.Response.Payload), taskPayload)
	if err != nil {
		logs.ErrorContextf(ctx, "Error unmarshalling task payload[%+v]: %v", resp, err)
		return 0, err
	}
	return response.Response.TaskID, nil
}

func CallBackTask(ctx context.Context, taskID uint, status task.TaskStatus, errMsg string, rst interface{}) error {
	url := joinURL(baseURL, "v3/knowledge.TaskCallBack")
	// 构造请求数据
	rstMsg := ""
	if rst != nil {
		rstMsgBytes, err := json.Marshal(rst)
		if err != nil {
			logs.ErrorContextf(ctx, "Error marshalling result: %v", err)
		}
		rstMsg = string(rstMsgBytes)
	}
	reqData := map[string]interface{}{
		"Request": map[string]interface{}{
			"task_id":       taskID,
			"worker_id":     workerID,
			"status":        status,
			"result":        rstMsg,
			"error_message": errMsg,
		},
	}
	logs.DebugContextf(ctx, "CallBackTask send request: %v", reqData)

	reqBody, err := json.Marshal(reqData)
	if err != nil {
		logs.ErrorContextf(ctx, "Error marshalling request body JSON: %v", err)
		return err
	}

	// 创建请求
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		logs.ErrorContextf(ctx, "Error making POST request: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		logs.ErrorContextf(ctx, "Received non-OK HTTP status: %s, body: %s", resp.Status, string(body))
		return fmt.Errorf("received non-OK HTTP status: %s", resp.Status)
	}
	return nil
}

// DownloadFile downloads a file from the given URL and saves it to the specified filepath.
func DownloadFile(ctx context.Context, url string, filepath string) error {
	logs.InfoContextf(ctx, "Downloading file from %s to %s", url, filepath)
	resp, err := http.Get(url)
	if err != nil {
		logs.ErrorContextf(ctx, "Error downloading file from %s: %v", url, err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		logs.ErrorContextf(ctx, "Failed to download file: %s, status: %s, body: %s", url, resp.Status, string(body))
		return fmt.Errorf("failed to download file: %s, status: %s", url, resp.Status)
	}

	out, err := os.Create(filepath)
	if err != nil {
		logs.ErrorContextf(ctx, "Error creating local file: %v", err)
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		logs.ErrorContextf(ctx, "Error saving downloaded file: %v", err)
		return err
	}

	logs.InfoContextf(ctx, "File downloaded successfully to %s", out.Name())
	return nil
}

func checkURL(url string) error {
	cli := &http.Client{
		Timeout: 1 * time.Second, // 设置超时时间
	}
	_, err := cli.Head(url)
	if err != nil {
		return fmt.Errorf("error checking URL %s: %w", url, err)
	}

	return nil
}

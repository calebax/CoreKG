package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/ygpkg/yg-go/lifecycle"
)

var runningTask = new(sync.Map)

func healthCheckRoutine() {
	for {
		select {
		case <-lifecycle.Std().C():
			return
		case <-time.After(15 * time.Second):
			HealthCheck()
		}
	}
}

// HealthCheck performs a health check by sending a request to the knowledge base API.
func HealthCheck() {
	url := joinURL(baseURL, "v3/knowledge.CheckInstance")

	// 构造请求数据
	data := map[string]interface{}{
		"Request": map[string]interface{}{
			"worker_id": workerID,
			"task_type": taskType,
		},
	}

	jsonData, _ := json.Marshal(data)

	// 创建请求
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, _ := ioutil.ReadAll(resp.Body)

	fmt.Println("HealthCheck Status:", resp.Status)
	fmt.Println("HealthCheck Response:", string(body))
}

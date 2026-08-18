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

func HealthCheck() {
	url := joinURL(baseURL, "v3/knowledge.CheckInstance")

	data := map[string]interface{}{
		"Request": map[string]interface{}{
			"worker_id": workerID,
			"task_type": taskType,
		},
	}

	jsonData, _ := json.Marshal(data)

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)

	fmt.Println("HealthCheck Status:", resp.Status)
	fmt.Println("HealthCheck Response:", string(body))
}

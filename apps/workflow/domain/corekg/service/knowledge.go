/*
 * Copyright 2025 coze-dev Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package service

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"

	"github.com/ygpkg/yg-go/logs"
	"github.com/insmtx/corekg/apps/workflow/utils/mysql/coresettings"
)

type coreKGKnowledgeSVC struct {
}

func NewCoreKGKnowledgeSVC() CoreKGKnowledge {
	svc := &coreKGKnowledgeSVC{}
	return svc
}

func (ck *coreKGKnowledgeSVC) GetKnowledgeFiles(request *GetKnowledgeFilesRequest) (response *GetKnowledgeFilesResponse, err error) {
	corekgUrl, err := coresettings.GetCoreKGUrl()
	if err != nil {
		logs.Errorf("get corekg url fail: %v", err)
		return nil, err
	}
	url := corekgUrl + "/v3/forest.ListFile"
	payload := strings.NewReader(`
		{
			"cmd": "/v3/forest.ListFile",
			"env": "prod",
			"version": "v1.10.9",
			"request": {
				"forest_id": ` + strconv.Itoa(int(request.CoreKGKnowledgeID)) + `,
				"list_all": true,
				"limit": 0,
				"offset": 0,
				"orderBy": [
					"updated_at desc"
				],
				"filters": [
					{
						"field": "parent_id",
						"value": [
							"0"
						]
					}
				]
			}
		}
	`)

	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		logs.Errorf("create request fail: %v", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if request.CoreKGToken != "" {
		req.Header.Set("Authorization", "Bearer "+request.CoreKGToken)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logs.Errorf("do request fail: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Errorf("read response body fail: %v", err)
		return nil, err
	}
	fmt.Println(string(body))

	var respData GetKnowledgeFilesResponse
	if err := json.Unmarshal(body, &respData); err != nil {
		logs.Errorf("unmarshal response body fail: %v", err)
		return nil, err
	}
	return &respData, nil
}

func (ck *coreKGKnowledgeSVC) GetKnowledgeFilesUrl(request *GetKnowledgeFilesUrlRequest) (response *GetKnowledgeFilesUrlResponse, err error) {
	corekgUrl, err := coresettings.GetCoreKGUrl()
	if err != nil {
		logs.Errorf("get corekg url fail: %v", err)
		return nil, err
	}
	url := corekgUrl + "/v3/forest.PreviewFileByURL"
	payload := strings.NewReader(`
		{
			"cmd": "/v3/forest.PreviewFileByURL",
			"env": "prod",
			"version": "v1.10.9",
			"request": {
				"file_id": ` + strconv.Itoa(int(request.FileID)) + `
			}
		}
	`)
	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		logs.Errorf("create request fail: %v", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", corekgUrl)
	if request.CoreKGToken != "" {
		req.Header.Set("Authorization", "Bearer "+request.CoreKGToken)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logs.Errorf("do request fail: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Errorf("read response body fail: %v", err)
		return nil, err
	}
	var respData GetKnowledgeFilesUrlResponse
	if err := json.Unmarshal(body, &respData); err != nil {
		logs.Errorf("unmarshal response body fail: %v", err)
		return nil, err
	}
	return &respData, nil
}

func (ck *coreKGKnowledgeSVC) GetKnowledgeSlice(request *GetKnowledgeSliceRequest) (response *GetKnowledgeSliceResponse, err error) {
	corekgUrl, err := coresettings.GetCoreKGUrl()
	if err != nil {
		logs.Errorf("get corekg url fail: %v", err)
		return nil, err
	}
	url := corekgUrl + "/v3/kesearch.ListFileChunk"
	payload := strings.NewReader(`
		{
			"cmd": "/v3/kesearch.ListFileChunk",
			"env": "prod",
			"version": "v1.10.9",
			"request": {
				"file_id": ` + strconv.Itoa(int(request.FileID)) + `,
				"forest_id": ` + strconv.Itoa(int(request.ForestID)) + `
			}
		}
	`)

	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		logs.Errorf("create request fail: %v", err)
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if request.CoreKGToken != "" {
		req.Header.Set("Authorization", "Bearer "+request.CoreKGToken)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logs.Errorf("do request fail: %v", err)
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logs.Errorf("read response body fail: %v", err)
		return nil, err
	}
	var respData GetKnowledgeSliceResponse
	if err := json.Unmarshal(body, &respData); err != nil {
		logs.Errorf("unmarshal response body fail: %v", err)
		return nil, err
	}
	return &respData, nil
}

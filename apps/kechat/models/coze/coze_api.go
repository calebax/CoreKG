package coze

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
)

// GetSpaceAPI 获取coze页面ID
func GetSpaceAPI(ctx *gin.Context, cozeUrl, token string) (spaceID string, code int, err error) {
	pluginData := map[string]interface{}{}
	cozeResult := GetListResponse{}
	jsonData, err := json.Marshal(pluginData)
	if err != nil {
		logs.ErrorContextf(ctx, "marshal json data error, %v", err)
		return "", 0, err
	}

	pluginURL := cozeUrl + "/api/playground_api/space/list"
	req, err := http.NewRequest("POST", pluginURL, bytes.NewBuffer(jsonData))
	if err != nil {
		logs.ErrorContextf(ctx, "create request error, %v", err)
		return "", 0, err
	}

	// 3. 设置必要请求头（保持与登录时一致的浏览器环境）
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logs.ErrorContextf(ctx, "client.Do error, %v", err)
		return "", 0, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		logs.ErrorContextf(ctx, "[GetSpaceAPI] status code error: %v, resp: %v", resp.StatusCode, resp)
		return "", 0, err
	}

	if err = json.NewDecoder(resp.Body).Decode(&cozeResult); err != nil {
		logs.ErrorContextf(ctx, "json.NewDecoder error:  %v", err)
		return "", 0, err
	}
	if len(cozeResult.Data.BotSpaceList) == 0 {
		response, _ := io.ReadAll(resp.Body)
		logs.WarnContextf(ctx, "API: %s \n resp: %v", pluginURL, string(response))
		return "", cozeResult.Code, nil
	}
	return cozeResult.Data.BotSpaceList[0].SpaceID, cozeResult.Code, nil
}

// CreatePlugin 创建coze插件
func CreatePlugin(ctx *gin.Context, key, spaceID, name, cozeUrl, sessionKey string) (pluginID string, err error) {
	corekgUrl, err := settings.GetText("corekg", "corekg_url")
	if err != nil {
		logs.ErrorContextf(ctx, "get coze url err %v", err)
		return
	}
	payload := strings.NewReader(`
{
    "name": "` + name + `",
    "desc": "` + name + `",
    "url": "` + corekgUrl + `/",
    "icon": {
        "uri": "default_icon/plugin_default_icon.png"
    },
	"source_type":"corekg",
    "auth_type": 0,
    "oauth_info": "{}",
    "space_id": "` + spaceID + `",
    "common_params": {
        "1": [],
        "2": [],
        "3": [],
        "4": [
            {
                "name": "User-Agent",
                "value": "Coze/1.0"
            },
            {
                "name": "Authorization",
                "value": "` + key + `"
            }
        ]
    },
    "creation_method": 0,
    "ide_code_runtime": "1",
    "plugin_type": 1
}
`)

	// 2. 创建请求（使用同一个 client，自动携带 Cookie）
	pluginURL := cozeUrl + "/api/plugin_api/register_plugin_meta"
	req, err := http.NewRequest("POST", pluginURL, payload)
	if err != nil {
		logs.ErrorContextf(ctx, "create request error, %v", err)
		return pluginID, err
	}

	// 3. 设置必要请求头（保持与登录时一致的浏览器环境）
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sessionKey)
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logs.ErrorContextf(ctx, "client.Do error, %v", err)
		return pluginID, err
	}
	defer resp.Body.Close()
	var pluginResp CreatePluginResponse
	if err = json.NewDecoder(resp.Body).Decode(&pluginResp); err != nil {
		logs.ErrorContextf(ctx, "json.NewDecoder error:  %v", err)
		return pluginID, err
	}
	if pluginResp.Code != 0 {
		response, _ := io.ReadAll(payload)
		logs.WarnContextf(ctx, "API: %s \n req: %v", pluginURL, payload)
		logs.ErrorContextf(ctx, "API: %s \n resp: %v", pluginURL, string(response))
		return "", errors.New("create coze plugin fail")
	}
	return pluginResp.PluginID, nil
}

// CreateCozeAPI 创建coze工具
func CreateCozeAPI(ctx *gin.Context, pluginID string, cozeUrl, sessionKey string) (string, error) {
	payload := strings.NewReader(`
{
    "plugin_id": "` + pluginID + `",
    "name": "agentAPI",
    "desc": "agentAPI",
    "path": "/pkg1",
    "method": 2,
    "edit_version": 0
}
`)

	// 2. 创建请求（使用同一个 client，自动携带 Cookie）
	pluginURL := cozeUrl + "/api/plugin_api/create_api"
	req, err := http.NewRequest("POST", pluginURL, payload)
	if err != nil {
		logs.ErrorContextf(ctx, "create request error, %v", err)
		return pluginID, err
	}

	// 3. 设置必要请求头（保持与登录时一致的浏览器环境）
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sessionKey)
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logs.ErrorContextf(ctx, "client.Do error, %v", err)
		return pluginID, err
	}
	defer resp.Body.Close()
	var pluginResp CreateCozeAPIResponse
	if err = json.NewDecoder(resp.Body).Decode(&pluginResp); err != nil {
		logs.ErrorContextf(ctx, "json.NewDecoder error:  %v", err)
		return pluginID, err
	}
	if pluginResp.Code != 0 {
		response, _ := io.ReadAll(payload)
		logs.WarnContextf(ctx, "API: %s \n req: %v", pluginURL, payload)
		logs.WarnContextf(ctx, "API: %s \n resp: %v", pluginURL, string(response))
		return "", errors.New("create coze api fail")
	}
	return pluginResp.APIID, nil
}

// UpdateCozeAPI 编辑coze工具
func UpdateCozeAPI(ctx *gin.Context, pluginID, apiID, cozeUrl, sessionKey string, inputNum int) error {
	var inputstr string
	var chatOptions string
	for i := 1; i < inputNum; i++ {
		inputstr += ",{}"
	}
	if inputNum != 0 {
		chatOptions = `
{
            "id": "` + CreateUUID() + `",
            "name": "chat_options",
            "desc": "chat_options",
            "type": 4,
            "location": 3,
            "is_required": true,
            "sub_parameters": [
                {
                    "id": "` + CreateUUID() + `",
                    "name": "input",
                    "desc": "input",
                    "type": 5,
                    "location": 3,
                    "is_required": true,
                    "sub_parameters": [
                        {
                            "id": "` + CreateUUID() + `",
                            "name": "[Array Item]",
                            "desc": "input",
                            "type": 4,
                            "location": 3,
                            "is_required": true,
                            "sub_parameters": [
                                {
                                    "id": "` + CreateUUID() + `",
                                    "name": "name",
                                    "desc": "name",
                                    "type": 1,
                                    "location": 3,
                                    "is_required": true,
                                    "sub_parameters": [],
                                    "global_disable": false,
                                    "local_disable": false,
                                    "deep": 4
                                },
                                {
                                    "id": "` + CreateUUID() + `",
                                    "name": "value",
                                    "desc": "value",
                                    "type": 1,
                                    "location": 3,
                                    "is_required": true,
                                    "sub_parameters": [],
                                    "global_disable": false,
                                    "local_disable": false,
                                    "deep": 4
                                }
                            ],
                            "global_disable": false,
                            "local_disable": false,
                            "deep": 3
                        }
                    ],
                    "global_default": "[{}` + inputstr + `]",
                    "global_disable": false,
                    "local_default": "[{}` + inputstr + `]",
                    "local_disable": false,
                    "deep": 2,
                    "value": "[{}` + inputstr + `]"
                }
            ],
            "global_disable": false,
            "local_disable": false,
            "deep": 1
        },
`
	}
	payload := strings.NewReader(`
{
    "plugin_id": "` + pluginID + `",
    "api_id": "` + apiID + `",
    "path": "v3/chat.Agent/chat/completions",
    "method": 2,                   
        "request_params": [
        ` + chatOptions + `
        {
            "id": "` + CreateUUID() + `",
            "name": "model",
            "desc": "model",
            "type": 1,
            "location": 3,
            "is_required": true,
            "sub_parameters": [],
            "global_default": "",
            "global_disable": false,
            "local_default": "",
            "local_disable": false,
            "deep": 1
        },
        {
            "id": "` + CreateUUID() + `",
            "name": "stream",
            "desc": "stream",
            "type": 6,
            "location": 3,
            "is_required": true,
            "sub_parameters": [],
            "global_default": "",
            "global_disable": false,
            "local_default": "",
            "local_disable": false,
            "deep": 1
        },
        {
            "id": "` + CreateUUID() + `",
            "name": "messages",
            "desc": "messages",
            "type": 5,
            "location": 3,
            "is_required": true,
            "sub_parameters": [
                {
                    "id": "` + CreateUUID() + `",
                    "name": "[Array Item]",
                    "desc": "messages",
                    "type": 4,
                    "location": 3,
                    "is_required": true,
                    "sub_parameters": [
                        {
                            "id": "` + CreateUUID() + `",
                            "name": "role",
                            "desc": "role",
                            "type": 1,
                            "location": 3,
                            "is_required": true,
                            "sub_parameters": [],
                            "global_disable": false,
                            "local_disable": false,
                            "deep": 3
                        },
                        {
                            "id": "` + CreateUUID() + `",
                            "name": "content",
                            "desc": "content",
                            "type": 1,
                            "location": 3,
                            "is_required": true,
                            "sub_parameters": [],
                            "global_disable": false,
                            "local_disable": false,
                            "deep": 3
                        }
                    ],
                    "global_disable": false,
                    "local_disable": false,
                    "deep": 2
                }
            ],
            "global_default": "[{}]",
            "global_disable": false,
            "local_default": "",
            "local_disable": false,
            "deep": 1
        }
    ],
    "response_params": [
        {
            "id": "` + CreateUUID() + `",
            "name": "choices",
            "desc": "choices",
            "type": 5,
            "location": 2,
            "sub_parameters": [
                {
                    "id": "` + CreateUUID() + `",
                    "name": "[Array Item]",
                    "desc": "choices",
                    "type": 4,
                    "sub_parameters": [
                        {
                            "id": "` + CreateUUID() + `",
                            "name": "message",
                            "desc": "message",
                            "type": 4,
                            "sub_parameters": [
                                {
                                    "id": "` + CreateUUID() + `",
                                    "name": "content",
                                    "desc": "content",
                                    "type": 1,
                                    "sub_parameters": [],
                                    "deep": 4
                                },
                                {
                                    "id": "` + CreateUUID() + `",
                                    "name": "role",
                                    "desc": "role",
                                    "type": 1,
                                    "sub_parameters": [],
                                    "deep": 4
                                }
                            ],
                            "deep": 3,
                            "assist_type": null
                        }
                    ],
                    "deep": 2,
                    "assist_type": null
                }
            ],
            "deep": 1,
            "assist_type": null,
            "global_disable": false
        }
    ],
    "edit_version": 0
}
`)

	// 2. 创建请求（使用同一个 client，自动携带 Cookie）
	pluginURL := cozeUrl + "/api/plugin_api/update_api"
	req, err := http.NewRequest("POST", pluginURL, payload)
	if err != nil {
		logs.ErrorContextf(ctx, "create request error, %v", err)
		return err
	}

	// 3. 设置必要请求头（保持与登录时一致的浏览器环境）
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sessionKey)
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logs.ErrorContextf(ctx, "client.Do error, %v", err)
		return err
	}
	defer resp.Body.Close()
	var updateAPIResp UpdateCozeAPIResponse
	if err = json.NewDecoder(resp.Body).Decode(&updateAPIResp); err != nil {
		logs.ErrorContextf(ctx, "json.NewDecoder error:  %v", err)
		return err
	}
	if updateAPIResp.Code != 0 {
		response, _ := io.ReadAll(payload)
		logs.WarnContextf(ctx, "API: %s \n req: %v", pluginURL, payload)
		logs.WarnContextf(ctx, "API: %s \n resp: %v", pluginURL, string(response))
		return errors.New("update coze plugin api failed")
	}
	return nil
}

func CreateUUID() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	const length = 21

	rand.Seed(time.Now().UnixNano())
	randomBytes := make([]byte, length)
	for i := range randomBytes {
		randomBytes[i] = charset[rand.Intn(len(charset))]
	}
	return string(randomBytes)
}

// DebugCozeAPI 试运行coze工具
func DebugCozeAPI(ctx *gin.Context, pluginID, apiID, model, sessionKey string, inputs []chattype.Params, cozeUrl string) error {
	var chatOptionsStr string
	for num, input := range inputs {
		chatOptionsStr += fmt.Sprintf("{\\\"name\\\": \\\"" + input.Name + "\\\",\\\"value\\\": \\\"" + "测试" + "\\\"}")
		if num != len(inputs)-1 {
			chatOptionsStr += ","
		}
	}
	payload := strings.NewReader(`
{
    "plugin_id": "` + pluginID + `",
    "api_id": "` + apiID + `",
    "parameters": "{\"model\":\"` + model + `\",\"stream\":false,\"messages\":[{\"role\": \"user\",\"content\": \"你好\"}],\"chat_options\":{\"input\":[` + chatOptionsStr + `]}}",
    "operation": 1
}
`)

	pluginURL := cozeUrl + "/api/plugin_api/debug_api"
	req, err := http.NewRequest("POST", pluginURL, payload)
	if err != nil {
		logs.ErrorContextf(ctx, "create request error, %v", err)
		return err
	}

	// 3. 设置必要请求头（保持与登录时一致的浏览器环境）
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sessionKey)
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logs.ErrorContextf(ctx, "client.Do error, %v", err)
		return err
	}
	defer resp.Body.Close()
	var debugAPIResp DebugCozeAPIResponse
	if err = json.NewDecoder(resp.Body).Decode(&debugAPIResp); err != nil {
		logs.ErrorContextf(ctx, "json.NewDecoder error:  %v", err)
		return err
	}
	if debugAPIResp.Code != 0 || debugAPIResp.Success != true {
		response, _ := io.ReadAll(payload)
		logs.WarnContextf(ctx, "API: %s \n req: %v", pluginURL, payload)
		logs.WarnContextf(ctx, "API: %s \n resp: %v", pluginURL, string(response))
		return errors.New("debug coze api error")
	}
	return nil
}

// PublishPlugin 发布coze工具
func PublishPlugin(ctx *gin.Context, pluginID, sessionKey string, cozeUrl string) error {

	payload := strings.NewReader(`
{
    "plugin_id": "` + pluginID + `",
    "version_name": "v1.0.0",
    "version_desc": "111"
}
`)

	pluginURL := cozeUrl + "/api/plugin_api/publish_plugin"
	req, err := http.NewRequest("POST", pluginURL, payload)
	if err != nil {
		logs.ErrorContextf(ctx, "create request error, %v", err)
		return err
	}

	// 3. 设置必要请求头（保持与登录时一致的浏览器环境）
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sessionKey)
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logs.ErrorContextf(ctx, "client.Do error, %v", err)
		return err
	}
	defer resp.Body.Close()
	var ppResp publishPluginResponse
	if err = json.NewDecoder(resp.Body).Decode(&ppResp); err != nil {
		logs.ErrorContextf(ctx, "json.NewDecoder error:  %v", err)
		return err
	}
	if ppResp.Code != 0 {
		response, _ := io.ReadAll(payload)
		logs.WarnContextf(ctx, "API: %s \n req: %v", pluginURL, payload)
		logs.WarnContextf(ctx, "API: %s \n resp: %v", pluginURL, string(response))
		return errors.New("publish plugin response error")
	}
	return nil
}

// Deprecated: 不再需要同步
// CreateKnowledgeAPI 创建coze知识库
func CreateKnowledgeAPI(ctx *gin.Context, spaceID, name, forestType, sessionKey string, detailID uint, cozeUrl string) (string, error) {
	payload := strings.NewReader(`
{
    "name": "` + name + `",
    "space_id": "` + spaceID + `",
	"forest_type":"` + forestType + `",
    "icon_uri": "default_icon/text_kn_default_icon.png",
    "format_type": 5,
	"corekg_detail_id": ` + strconv.FormatUint(uint64(detailID), 10) + `
}
`)

	// 2. 创建请求（使用同一个 client，自动携带 Cookie）
	pluginURL := cozeUrl + "/api/knowledge/create"
	req, err := http.NewRequest("POST", pluginURL, payload)
	if err != nil {
		logs.ErrorContextf(ctx, "create request error, %v", err)
		return "", err
	}

	// 3. 设置必要请求头（保持与登录时一致的浏览器环境）
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sessionKey)
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logs.ErrorContextf(ctx, "client.Do error, %v", err)
		return "", err
	}
	defer resp.Body.Close()
	var pluginResp CreateKnowledgeAPIResponse
	if err = json.NewDecoder(resp.Body).Decode(&pluginResp); err != nil {
		logs.ErrorContextf(ctx, "json.NewDecoder error:  %v", err)
		return "", err
	}
	if pluginResp.Code != 0 {
		response, _ := io.ReadAll(payload)
		logs.WarnContextf(ctx, "API: %s \n req: %v", pluginURL, payload)
		logs.WarnContextf(ctx, "API: %s \n resp: %v", pluginURL, string(response))
		return "", errors.New("create knowledge api response error")
	}
	return pluginResp.DatasetID, nil
}

// GetCozeAgent 获取coze插件列表
func GetCozeAgent(ctx *gin.Context) (GetCozeSourceTypeResponse, error) {
	var cozeResp GetCozeSourceTypeResponse

	cozeUrl, err := settings.GetText("corekg", "coze_url")
	if err != nil {
		logs.ErrorContextf(ctx, "get coze url err %v", err)
		return cozeResp, err
	}
	sessionKey := runtime.LoginStatus(ctx).Token
	space, _, err := GetSpaceAPI(ctx, cozeUrl, sessionKey)
	if err != nil {
		logs.ErrorContextf(ctx, "get coze space error, %v", err)
		return cozeResp, err
	}

	if len(space) == 0 {
		logs.WarnContextf(ctx, "get coze space failed")
		return cozeResp, errors.New("coze space is empty")
	}

	payload := strings.NewReader(`
{
    "user_filter": 0,
    "res_type_filter": [
        1
    ],
    "name": "",
    "publish_status_filter": 0,
    "space_id": "` + space + `",
    "size": 0
}
`)

	pluginURL := cozeUrl + "/api/plugin_api/library_resource_list"
	req, err := http.NewRequest("POST", pluginURL, payload)
	if err != nil {
		logs.ErrorContextf(ctx, "create request error, %v", err)
		return cozeResp, err
	}

	// 3. 设置必要请求头（保持与登录时一致的浏览器环境）
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+sessionKey)
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36")

	client := &http.Client{}
	cozeAPIResp, err := client.Do(req)
	if err != nil {
		logs.ErrorContextf(ctx, "client.Do error, %v", err)
		return cozeResp, err
	}
	defer cozeAPIResp.Body.Close()
	if err = json.NewDecoder(cozeAPIResp.Body).Decode(&cozeResp); err != nil {
		logs.ErrorContextf(ctx, "json.NewDecoder error:  %v", err)
		return cozeResp, err
	}
	return cozeResp, nil
}

// Deprecated: 不再需要获取coze数据源
// GetCozeKnowledge 获取coze知识库列表
func GetCozeKnowledge(ctx *gin.Context, resp forest.ForestInfoItemList) (forest.ForestInfoItemList, error) {
	var cozeResp GetCozeSourceTypeResponse
	uin := runtime.Uin(ctx)

	cozeUrl, err := settings.GetText("corekg", "coze_url")
	if err != nil {
		logs.ErrorContextf(ctx, "get coze url err %v", err)
		return resp, err
	}
	sessionKey := runtime.LoginStatus(ctx).Token
	space, _, err := GetSpaceAPI(ctx, cozeUrl, sessionKey)
	if err != nil {
		logs.ErrorContextf(ctx, "get coze space error, %v", err)
		return resp, err
	}
	if space == "" {
		logs.WarnContextf(ctx, "get coze space failed")
		return resp, errors.New("coze space is empty")
	}
	payload := strings.NewReader(`
{
    "user_filter": 0,
    "res_type_filter": [
        4, -1
    ],
    "name": "",
    "publish_status_filter": 0,
    "space_id": "` + space + `",
    "size": 0
}
`)

	pluginURL := cozeUrl + "/api/plugin_api/library_resource_list"
	req, err := http.NewRequest("POST", pluginURL, payload)
	if err != nil {
		logs.ErrorContextf(ctx, "create request error, %v", err)
		return resp, err
	}

	// 3. 设置必要请求头（保持与登录时一致的浏览器环境）
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36")
	req.Header.Set("Authorization", "Bearer "+sessionKey)

	client := &http.Client{}
	cozeAPIResp, err := client.Do(req)
	if err != nil {
		logs.ErrorContextf(ctx, "client.Do error, %v", err)
		return resp, err
	}
	defer cozeAPIResp.Body.Close()
	if cozeAPIResp.StatusCode != http.StatusOK {
		logs.ErrorContextf(ctx, "cozeAPIResp.StatusCode error, %d", cozeAPIResp.StatusCode)
		return resp, fmt.Errorf("coze knowledge api response error: %v,resp :%v", cozeAPIResp.StatusCode, cozeAPIResp)
	}

	if err = json.NewDecoder(cozeAPIResp.Body).Decode(&cozeResp); err != nil {
		logs.ErrorContextf(ctx, "json.NewDecoder error:  %v", err)
		return resp, err
	}
	items, err := chattype.GetCozeMappingByID(ctx, uin, chattype.ChatTypeForest)
	if err != nil {
		logs.ErrorContextf(ctx, "GetCozeMappingByID error, %v", err)
		return resp, err
	}
	corekgMap := make(map[uint]int)
	for i, s := range resp.Data {
		corekgMap[s.ID] = i
		resp.Data[i].IsSynced = false
	}
	mappingMap := make(map[string]uint)
	for _, mapping := range items {
		mappingMap[mapping.CozeID] = mapping.CoreKGID
	}
	for _, s := range cozeResp.ResourceList {
		if cozeId, ok := mappingMap[s.ResID]; ok {
			if v, ok := corekgMap[cozeId]; ok {
				resp.Data[v].IsSynced = true
			}
		}
	}
	return resp, nil
}

// DeleteCozePluginAPI 删除coze插件
func DeleteCozePluginAPI(ctx *gin.Context, pluginID, token, cozeUrl string) error {

	payload := strings.NewReader(`
{
    "plugin_id": "` + pluginID + `"
}
`)

	pluginURL := cozeUrl + "/api/plugin_api/del_plugin"
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
	var wkResp CreateCozeWorkflowAPIResponse
	if err = json.NewDecoder(body).Decode(&wkResp); err != nil {
		logs.ErrorContextf(ctx, "json.NewDecoder error:  %s", err)
		return err
	}
	if wkResp.Code != 0 {
		response, _ := io.ReadAll(body)
		logs.WarnContextf(ctx, "API: %s \n req: %v", pluginURL, payload)
		logs.WarnContextf(ctx, "API: %s \n resp: %v", pluginURL, string(response))
		return errors.New("deleted coze api failed")
	}
	return nil
}

// DeleteCozeKnowledgeAPI 删除coze知识库
func DeleteCozeKnowledgeAPI(ctx *gin.Context, pluginID, token, cozeUrl string) error {

	payload := strings.NewReader(`
{
    "dataset_id": "` + pluginID + `"
}
`)

	pluginURL := cozeUrl + "/api/knowledge/delete"
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
	var wkResp CreateCozeWorkflowAPIResponse
	if err = json.NewDecoder(body).Decode(&wkResp); err != nil {
		logs.ErrorContextf(ctx, "json.NewDecoder error:  %s", err)
		return err
	}
	if wkResp.Code != 0 {
		response, _ := io.ReadAll(body)
		logs.WarnContextf(ctx, "API: %s \n req: %v", pluginURL, payload)
		logs.ErrorContextf(ctx, "API: %s \n resp: %v", pluginURL, string(response))
		return errors.New("deleted coze knowledge api failed")
	}
	return nil
}

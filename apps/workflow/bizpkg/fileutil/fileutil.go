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
package fileutil

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"

	"github.com/ygpkg/yg-go/logs"
)

func GetWorkingDirectory() string {
	root, err := os.Getwd()
	if err != nil {
		logs.Warnf("[InitConfig] Failed to get current working directory: %v", err)
		root = os.Getenv("PWD")
	}
	return root
}

// GetAppRoot 返回 workflow 应用根目录（apps/workflow 的绝对路径）。
//
// workflow 的 model_meta.json、prompt 模板、workflow 运行配置等后端资源统一位于
// apps/workflow/conf 下。这里从进程工作目录（聚合进 corekg 时为仓库根、独立运行时为
// apps/workflow）向上查找含 go.mod 的仓库根，再拼上固定的相对段 apps/workflow，
// 从而在两种运行形态下都能正确定位，避免依赖业务代码随机调用 os.Getwd() 猜测 cwd。
// 调用方应基于该根拼接 "conf/..."。
func GetAppRoot() string {
	dir, err := os.Getwd()
	if err != nil {
		dir = os.Getenv("PWD")
	}

	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return filepath.Join(dir, "apps", "workflow")
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	// 兜底：找不到仓库根时回到进程工作目录（独立运行时 cwd=apps/workflow）。
	return GetWorkingDirectory()
}

func ReadJinja2PromptTemplate(jsonFilePath string) (prompt.ChatTemplate, error) {
	b, err := os.ReadFile(jsonFilePath)
	if err != nil {
		return nil, err
	}
	var m2qMessages []*schema.Message
	if err = json.Unmarshal(b, &m2qMessages); err != nil {
		return nil, err
	}
	tpl := make([]schema.MessagesTemplate, len(m2qMessages))
	for i := range m2qMessages {
		tpl[i] = m2qMessages[i]
	}
	return prompt.FromMessages(schema.Jinja2, tpl...), nil
}

package agent

import (
	"context"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/fs"
	"github.com/insmtx/corekg/apps/kecore/services/graphragsearch/search"
	"github.com/insmtx/corekg/apps/ketask/models/ragtypes"
	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
	"github.com/ygpkg/yg-go/logs"
)

const (
	// FilterPrompt 筛选
	FilterPrompt = `你是一个只负责返回“文件ID列表”的检索结果筛选器，而不是对话助手。
根据检索到的文件摘要筛选出可能相关的文件
【数据说明】
系统已检索到若干文件，信息如下：
------BEGIN------
informations: {informations}
------END------

每个文件对象包含：
- analysis：文件摘要内容
- file_id：文件唯一ID

【用户问题】
{query}

【唯一任务】
根据“用户问题”，从 informations 中筛选出最相关的文件。

【强制输出规则（必须严格遵守）】
1. 输出内容只能是【文件ID】，以英文逗号分隔，例如：
   1,2,1001
2. 不允许输出任何其他字符，包括但不限于：
   - 中文或英文说明
   - 空格、换行
   - 标点（除英文逗号外）
   - 代码块、Markdown
   - 注释、前缀、后缀
3. 返回的 file_id 必须真实存在于 informations 中，严禁编造。
4. 最多返回 10 个 file_id。
5. 如果没有任何相关文件，返回空字符串（什么都不输出）。
6. 任何不符合以上规则的输出，均视为任务失败。

【重要】
你不是聊天模型，不需要解释、不需要总结、不需要推理展示。
你唯一允许输出的内容：符合规则的文件ID列表。

`
)

type Analysis struct {
	FileID   uint   `json:"file_id"`
	Analysis string `json:"analysis"`
	// Content  string `json:"content"`
}

type FileData struct {
	FileID   uint   `json:"file_id"`
	Analysis string `json:"analysis"`
	Content  string `json:"content"`
}

func NewFilterAgent(ctx context.Context, chatModel model.ToolCallingChatModel, files []search.DirectoryInfo) (adk.Agent, []*FileData, error) {
	con := []*FileData{}
	an := []*Analysis{}
	fMap := map[uint]search.DirectoryInfo{}
	for _, v := range files {
		if _, ok := fMap[v.FileID]; ok {
			continue
		}
		fMap[v.FileID] = v
		f, err := forest.GetForestFileByID(v.FileID)
		if err != nil {
			logs.ErrorContextf(ctx, "GetAnalysis GetForestFileByID err: %v", err)
			continue
		}
		w := essearch.NewPureWrapper(ctx, "ke_0", []uint{f.ForestID}, []uint{f.ID}, nil)
		desc, err := w.GetFileDesc()
		if err != nil {
			logs.ErrorContextf(ctx, "GetAnalysis GetAnalysis err: %v,desc[%v]", err, desc)
			return nil, nil, err
		}
		if desc == nil {
			desc = &ragtypes.FileDescription{
				Description: "",
			}
		}
		content, err := fs.GetFileContent(f)
		if err != nil {
			logs.ErrorContextf(ctx, "GetAnalysis GetFileContent err: %v", err)
			return nil, nil, err
		}
		con = append(con, &FileData{
			Content:  string(content),
			Analysis: desc.Description,
			FileID:   f.ID,
		})
		an = append(an, &Analysis{
			FileID:   f.ID,
			Analysis: desc.Description,
		})
	}

	ag, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "SearchFilterAnalyst",
		Description: "搜索到的内容进行筛选",
		Model:       chatModel,
		Instruction: `你是一个专业的内容筛选助手，能够根据用户提供的内容和用户的问题筛选出与用户问题最相关的内容。`,
		GenModelInput: func(ctx context.Context, instruction string, input *adk.AgentInput) ([]adk.Message, error) {
			logs.InfoContextf(ctx, "NewFilterAgent GenModelInput : %s", logs.JSON(input))
			ct := prompt.FromMessages(schema.FString,
				schema.SystemMessage(instruction),
				schema.UserMessage(FilterPrompt),
			)

			msgs, err := ct.Format(ctx, map[string]any{
				"informations": logs.JSON(an),
				"query":        input.Messages[0].Content,
			})
			if err != nil {
				logs.ErrorContextf(ctx, "NewFilterAgent GenModelInput ct.Format error: %v", err)
				return nil, err
			}
			return msgs, nil
		},
	})
	if err != nil {
		logs.ErrorContextf(ctx, "NewFilterAgent NewChatModelAgent error: %v", err)
		return nil, nil, err
	}
	return ag, con, nil
}

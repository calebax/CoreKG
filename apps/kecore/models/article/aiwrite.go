package article

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/models/chatagent"
	"github.com/insmtx/corekg/apps/kechat/models/chatclient"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kechat/models/llmchat"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kesearch/pkg/ai/tools"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/apis/sseclient"
	"github.com/ygpkg/yg-go/dbtools/redispool"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
	"github.com/ygpkg/yg-go/types"
)

type AIWrapper struct {
	Ctx       *gin.Context
	SysAgent  string
	Question  string
	APIKey    string
	RequestID string
	ForestIDs types.UintArray
	Cmd       foresttype.CmdString
}

func NewAIWriteWrapper(ctx *gin.Context, cmd foresttype.CmdString, question string, apiKey string, forestIDs types.UintArray) (*AIWrapper, error) {
	if len(question) <= 0 {
		logs.WarnContextf(ctx, "NewAIWrapper:try to create a new AIWrapper with empty question")
		return nil, fmt.Errorf("question shouldn't be empty")
	}
	if apiKey == "" {
		sysAPIKey, err := settings.GetText(global.SettingGroupKnowledge, global.SettingKeySystemLlmAPIKey)
		if err != nil {
			return nil, fmt.Errorf("failed to get system api key: %v", err)
		}
		apiKey = sysAPIKey
	}
	w := &AIWrapper{
		Ctx:       ctx,
		Question:  question,
		APIKey:    apiKey,
		RequestID: runtime.RequestID(ctx),
		ForestIDs: forestIDs,
		Cmd:       cmd,
	}
	switch cmd {
	case foresttype.CmdAbbreviation:
		w.SysAgent = chatagent.GetAgentI18nName(ctx, runtime.GetLanguage(ctx), global.ChatAgentAIWriteAbbreviation)
	case foresttype.CmdExpansion:
		w.SysAgent = chatagent.GetAgentI18nName(ctx, runtime.GetLanguage(ctx), global.ChatAgentAIWriteExpansion)
	case foresttype.CmdEmbellishment:
		w.SysAgent = chatagent.GetAgentI18nName(ctx, runtime.GetLanguage(ctx), global.ChatAgentAIWriteEmbellishment)
	case foresttype.CmdProofreading:
		w.SysAgent = chatagent.GetAgentI18nName(ctx, runtime.GetLanguage(ctx), global.ChatAgentAIWriteProofreading)
	case foresttype.CmdContinuation:
		w.SysAgent = chatagent.GetAgentI18nName(ctx, runtime.GetLanguage(ctx), global.ChatAgentAIWriteContinuation)
	default:
		w.SysAgent = chatagent.GetAgentI18nName(ctx, runtime.GetLanguage(ctx), global.ChatAgentAIWriteCustom)
	}
	logs.DebugContextf(ctx, "set AiWrite sys_agent =  %v", w.SysAgent)
	return w, nil
}

// DoCmd will do cmd for different chat agent
func (w *AIWrapper) DoCmd() error {
	frsIDs := w.ForestIDs.Slice()
	var forestSearchResult string = ""
	var hasSearchResult bool = false
	if len(frsIDs) > 0 {
		// TODO: es 索引改为动态传入
		sharedRefs := &tools.SharedReferences{
			Refs: make([]*chattype.QueryReference, 0),
		}
		forestSearchTool, err := tools.NewForestSearchTool(w.Ctx, &tools.ForestSearchToolConfig{
			Ctx:              w.Ctx,
			EsIndex:          "ke_0",
			ForestIDs:        frsIDs,
			FileIDs:          make([]uint, 0),
			ReferencesResult: sharedRefs,
		})
		if err != nil {
			logs.ErrorContextf(w.Ctx, "[forestChatMode] ForestChat NewForestSearchTool error: %v", err)
		} else {
			req := &tools.SearchRequest{
				Question:       w.Question,
				SearchStrategy: "common_questions",
			}
			reqBody, _ := json.Marshal(req)
			searchResult, err := forestSearchTool.InvokableRun(w.Ctx, string(reqBody))
			if err != nil {
				logs.ErrorContextf(w.Ctx, "[forestChatMode] ForestChat InvokableRun error: %v", err)
			}
			forestSearchResult = searchResult

			if err == nil && tools.HasNonEmptyResult(searchResult) {
				hasSearchResult = true
			}
		}
	}

	req := &chattype.ChatRequestBody{
		Stream:      true,
		Model:       w.SysAgent,
		LLMModelID:  0,
		ChatOptions: chattype.ChatOptions{},
	}

	if w.SysAgent != chatagent.GetAgentI18nName(w.Ctx, runtime.GetLanguage(w.Ctx), global.ChatAgentAIWriteCustom) {
		req.ChatOptions = chattype.ChatOptions{
			Input: []chattype.Input{
				//original text
				{Name: "input1", Value: w.Question},
				//forest search result
				{Name: "input2", Value: forestSearchResult},
				{Name: "input3", Value: func() string {
					if hasSearchResult {
						return ContinuationReferencePrompt
					}
					return ""
				}()},
			},
		}
	} else {
		req.ChatOptions = chattype.ChatOptions{
			Input: []chattype.Input{
				//custom cmd
				{Name: "input1", Value: string(w.Cmd)},
				//original text
				{Name: "input2", Value: w.Question},
			},
		}
	}

	logs.DebugContextf(w.Ctx, "req: %v", req)

	wrp, err := chatclient.NewInternalChat(w.Ctx, w.RequestID, w.APIKey, 0, req)
	if err != nil {
		logs.ErrorContextf(w.Ctx, "failed to create internal chat: %v", err)
		return err
	}
	sseClient := sseclient.New(sseclient.WithRedisClient(redispool.Redis()), sseclient.WithExpiration(time.Minute*5))
	defer sseClient.Close(w.Ctx, w.RequestID)
	sseClient.SetHeaders(w.Ctx.Writer)
	var buf strings.Builder
	defer func() {
		res := buf.String()
		logs.DebugContextf(w.Ctx, "Captch AIWrite result: %v", res[:min(100, len(res))])
		dbutil.Knownow().Save(&foresttype.KeArticleHistory{
			Cmd:       w.Cmd,
			Content:   w.Question,
			Result:    res,
			CompanyID: runtime.CompanyID(w.Ctx),
			Uin:       runtime.Uin(w.Ctx),
		})
	}()
	_, err = wrp.AgentChatInternal(func(chunk *chattype.ChatStreamResponseBody) error {
		writeResult := llmchat.WriteResult{
			ReasoningContent: chunk.Choices[0].Delta.ReasoningContent,
			Content:          chunk.Choices[0].Delta.Content,
		}
		buf.WriteString(chunk.Choices[0].Delta.Content)
		if stoped, err := sseClient.WriteMessage(w.Ctx, w.Ctx.Writer, w.RequestID, writeResult.String()); err != nil {
			logs.ErrorContextf(w.Ctx, "AIWrapper DoCmd write response error: %s", err)
			return err
		} else if stoped {
			logs.InfoContextf(w.Ctx, "AIWrapper DoCmd stream Stoped by client")
			return nil
		}
		return nil
	})
	if err != nil {
		logs.ErrorContextf(w.Ctx, "DoCmd failed err %v", err)
		return err
	}
	return nil
}

const (
	ContinuationReferencePrompt = `
#### 参考来源素材使用规则（如提供 **<Reference>**）
 - <Reference> 内容来自**专业知识库或已知材料**：
  - 默认仅用于**隐性风格、术语习惯、论证方式**的借鉴；
  - 在需要事实性支撑时，可作为**唯一合法的信息来源之一**。
- 在不涉及事实性陈述时，可以从中吸收专业术语使用习惯、论证逻辑或行文结构特征，而不显性引用。
- **若提供了 <Reference>，且原文内容允许事实性或论证性延展，续写时应至少引入一处与原文逻辑一致的事实性补充、具体论据或明确结论，并按下述“引用适用规则”提供对应引用。**
- **严禁**直接复制、改写或大段引用参考素材原文；即便在需要引用时，也只能用于事实依据，不得进行内容性复写。

---

### ❗引用生成的强制约束（非常重要）
  - **绝对禁止**生成任何 **{Reference ...}** 结构或类似引用标记
  - **引用标签只能且必须基于 **<Reference>** 中真实存在的信息生成**：
  - ❌ 严禁编造、猜测、推断或虚构任何引用标签  
  - ❌ 严禁引用未在 **<Reference>** 中出现的文件、chunk 或信息  
- 如果 **<Reference>** 中**不存在**可支撑某一事实的内容：
  - 则**不得**引入该事实  
  - 或必须改写为**不需要引用的非事实性表述**
- **不得为了满足“看起来专业”而强行添加引用**

---

### 何时必须引用

仅在出现以下内容时允许且必须引用：

- 明确的事实性陈述  
- 具体数据、时间、比例、结论  
- 非通用、非常识性的专业判断

### 何时不需要引用：
- 广泛接受的常识、通用解释、背景性描述无需引用。
- 原文本身已隐含且无需外部支撑的内容

---

### 引用格式规范（技术要求）

**统一使用以下文本块格式：**
` + "```\n{Reference §fileID[chunkSequence1, chunkSequence2, ...]}\n```" + `
#### 示例说明：
- **单一来源：**
` + "```\n{Reference §1234[16, 35, 108]}\n```" + `
- **多来源合并引用：**
` + "```\n{Reference §1234[16, 18], §4567[24]}\n```" + `

#### 引用粒度要求：
- 同一句话中，如多个信息来自同一来源，仅在句末引用一次。
- 避免为相邻句子或同一事实拆分多次引用，防止碎片化。

#### 严禁行为（再次强调）：
- ❌ 严禁编造、猜测或虚构引用标签
- ❌ 引用不存在的 fileID 或 chunk 序列
- ❌ 严禁为通用知识、常识性内容添加引用

---
	`
)

package qachat

import (
	"context"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/pkgs/testutils"
	"github.com/ygpkg/yg-go/logs"
)

func TestChat(t *testing.T) {
	testutils.Initialize(testutils.AppNameKechat)
	defer testutils.Close()
	ctx := &gin.Context{}
	req := &chattype.ChatRequestBody{
		Stream:     true,
		Model:      "CJKGBXd",
		LLMModelID: 1,
		ChatOptions: chattype.ChatOptions{
			Input: []chattype.Input{
				{Name: "input1", Value: "qs.Question"},
				{Name: "input2", Value: "searchstr"},
				{Name: "input3", Value: "history"},
				{Name: "input4", Value: "history"},
			},
		},
	}
	w, err := NewInternalChat(ctx, "123", "xxx", 1, req)
	if err != nil {
		t.Fatalf("failed to create internal chat: %v", err)
	}
	res, err := w.AgentChatInternal(func(chunk *chattype.ChatStreamResponseBody) error {
		for _, v := range chunk.Choices {
			print(v.Delta.ReasoningContent)
			print(v.Delta.Content)
		}
		return nil
	})
	if err != nil {
		logs.WarnContextf(ctx, "err: %v", err)
	}
	logs.InfoContextf(ctx, "res: %+v", res)
}

func TestParseReferences(t *testing.T) {
	a, _ := parseReferences(a)
	logs.InfoContextf(context.Background(), "a: %+v", a)
}

var a = `根据检索到的文档信息，以下是关于室外POE供电AP的推荐及配置建议：

### 推荐型号
1. **UAP672X**  
   - **产品特点**：IP67防护等级，6.5KV防浪涌  
   - **适用场景**：室外环境  
   {Reference §1429[47]}

### 配置建议
1. **POE交换机选择**  
   - 需搭配支持POE供电的交换机，例如：  
     - **US218-HP**（整机最大输出功率240W）  
     - **US206-P**（4口POE）或 **US210-P**（8口POE），根据接入端数量选择  
     {Reference §1429[47], §1450[77,79,80,81]}  

2. **部署参数**  
   - **覆盖面积**：60平米/AP（吸顶AP参考值，室外需根据实际环境调整）  
   - **终端数量**：建议不超过70个终端/AP  
   {Reference §1450[77,79,80,81]}  

3. **安装注意事项**  
   - 确保交换机供电功率满足AP需求（如UAP672X需符合IP67防护等级对应的供电稳定性）  
   - 若需多AP部署，建议通过UR7103/UR7208网关管理带机量  
   {Reference §1450[77,79,80,81]}  

若需更详细的配置步骤（如具体交换机端口设置或AP固件配置），当前检索信息中未提供明确指导，建议补充相关技术文档以便进一步检索。  

> 注：以上推荐均基于文档中的设备清单及参数，实际部署需结合现场环境测试。
`

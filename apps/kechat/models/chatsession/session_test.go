package chatsession

import (
	"context"
	"testing"

	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
)

func TestGetLLmSessionName(t *testing.T) {
	dbtools.InitMultiDBConn(map[string]string{
		"core": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=True&loc=Local",
		"chat": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	})
	str, _ := GetLLmSessionName(context.Background(), "我现在需要一个室外可POE供电的AP，请给我推荐下型号，并告诉我怎么配置", `根据检索到的文档信息，以下是关于室外POE供电AP的推荐及配置建议：

### 推荐型号
1. **UAP672X**  
   - **产品特点**：IP67防护等级，6.5KV防浪涌  
   - **适用场景**：室外环境  
   [Reference §1429[47]]

### 配置建议
1. **POE交换机选择**  
   - 需搭配支持POE供电的交换机，例如：  
     - **US218-HP**（整机最大输出功率240W）  
     - **US206-P**（4口POE）或 **US210-P**（8口POE），根据接入端数量选择  
     [Reference §1429[47], §1450[77,79,80,81]]  

2. **部署参数**  
   - **覆盖面积**：60平米/AP（吸顶AP参考值，室外需根据实际环境调整）  
   - **终端数量**：建议不超过70个终端/AP  
   [Reference §1450[77,79,80,81]]  

3. **安装注意事项**  
   - 确保交换机供电功率满足AP需求（如UAP672X需符合IP67防护等级对应的供电稳定性）  
   - 若需多AP部署，建议通过UR7103/UR7208网关管理带机量  
   [Reference §1450[77,79,80,81]]  

若需更详细的配置步骤（如具体交换机端口设置或AP固件配置），当前检索信息中未提供明确指导，建议补充相关技术文档以便进一步检索。  

> 注：以上推荐均基于文档中的设备清单及参数，实际部署需结合现场环境测试。
`)
	println(str)
}

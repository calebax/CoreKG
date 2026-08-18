package coze

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/perm"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/random"
)

func GetChatAgentByID(ctx context.Context, id uint) (*chattype.ChatAgent, error) {
	var agent chattype.ChatAgent
	if err := dbutil.Chat().WithContext(ctx).Where("id = ?", id).Where("deleted_at IS NULL").First(&agent).Error; err != nil {
		logs.ErrorContextf(ctx, "GetChatAgentByID err: %v", err)
		return nil, err
	}
	return &agent, nil
}

// GenerateSecretKey 生成随机密钥
func GenerateSecretKey() string {
	// 固定前缀
	prefix := "yg-"

	// 随机生成16字节的标识符
	randomBytes := make([]byte, 16)
	_, err := rand.Read(randomBytes)
	if err != nil {
		// panic(err)
		// logs.Errorf("[GenerateSecretKey] err: %s", err.Error())
	}

	// 转换为十六进制字符串
	return prefix + hex.EncodeToString(randomBytes)
}

// CreateAPIKey 创建轻应用APIKey
func CreateAPIKey(ctx *gin.Context, agentID, uin, companyID uint) (string, error) {
	ag, err := GetChatAgentByID(ctx, agentID)
	if err != nil {
		logs.ErrorContextf(ctx, "chat_agent.GetChatAgentByID(%d) failed: %v", agentID, err)
		return "", err
	}

	// 检查当前用户是否是管理员或机器人所有者
	if !perm.HasManageAct(ctx, uin, ag.ID, foresttype.ResourceTypeAgent) {
		logs.WarnContextf(ctx, "uin[%v] desire to update resource[%v]_id[%v] but isn't manager", uin, foresttype.ResourceTypeAgent, ag.ID)
		runtime.BadRequest(ctx, "无权限更新此资源")
		return "", err
	}
	name := fmt.Sprintf("%v_%v", ag.ShowName, random.String(6))
	key := &accounttype.APIKey{
		CompanyID:    companyID,
		Uin:          uin,
		Name:         name,
		APIKey:       GenerateSecretKey(),
		Purpose:      fmt.Sprintf("agent-%v-api", ag.ShowName),
		ResourceType: accounttype.ResourceTypeAgent,
		ResourceID:   ag.ID,
		Status:       accounttype.AccessKeyStatusNormal,
	}

	if err = dbutil.Account().Create(&key).Error; err != nil {
		logs.ErrorContextf(ctx, "create apikey failed: %v", err)
		runtime.InternalError(ctx, "创建apikey失败")
		return "", err
	}
	return key.APIKey, nil
}

func DeleteCozeMappingByCozeID(ctx context.Context, id string, chatType chattype.ChatType) error {
	return chattype.DeleteCozeMappingByCozeID(ctx, id, chatType)
}

func GetAgentMapping(ctx context.Context, cozeID string, chatType chattype.ChatType) (chattype.ChatCozeMapping, error) {
	agentMap, err := chattype.GetCozeMappingByCozeID(ctx, cozeID, chatType)
	if err != nil {
		logs.ErrorContextf(ctx, "GetCozeMappingByCozeID error, %s", err.Error())
		return agentMap, err
	}
	return agentMap, nil
}

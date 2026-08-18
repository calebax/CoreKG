package chattype

import (
	"context"
	"fmt"

	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

type MigrateResourceType string

const (
	MigrateResourceTypeChatQuesion    = "chat_question"
	MigrateResourceTypeForestSession  = "forest_session"
	MigrateResourceTypeForestQuestion = "forest_question"
)

// ChatModel Chat模型表
type ChatMigrate struct {
	gorm.Model

	ResourceType MigrateResourceType `gorm:"column:resource_type;type:varchar(64);not null;comment:'资源类型'" json:"resource_type"`
	ResourceID   uint                `gorm:"column:resource_id;type:bigint;not null;default:0;comment:'资源ID'" json:"resource_id"`
	// 目标ID
	TargetID    uint   `gorm:"column:target_id;type:bigint;not null;default:0;comment:'目标ID'" json:"target_id"`
	TargetIDStr string `gorm:"column:target_id_str;type:varchar(64);not null;default:'';comment:'目标ID'" json:"target_id_str"`
}

// TableName 表名
func (ChatMigrate) TableName() string {
	return TableNameChatMigrate
}

// ExistChatMigrate 判断当前记录是否迁移过了
func ExistChatMigrate(ctx context.Context, resource_id uint, resource_type string) error {
	var chatMigrate ChatMigrate
	err := dbutil.Chat().Model(&ChatMigrate{}).Where("resource_id = ? and resource_type = ?", resource_id, resource_type).First(&chatMigrate).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil
		}
		logs.ErrorContextf(ctx, "get chat migrate error: %v", err)
		return err
	}
	return fmt.Errorf("chat migrate already exists")
}

func GetChatMigrate(ctx context.Context, resource_id uint, resource_type string) (*ChatMigrate, error) {
	var chatMigrate ChatMigrate
	err := dbutil.Chat().WithContext(ctx).Model(&ChatMigrate{}).Where("resource_id = ? and resource_type = ?", resource_id, resource_type).First(&chatMigrate).Error
	if err != nil {
		logs.ErrorContextf(ctx, "get chat migrate error: %v", err)
		return nil, err
	}
	return &chatMigrate, nil
}

// CreateChatMigrate 创建聊天记录迁移
func CreateChatMigrate(ctx context.Context, chatMigrate *ChatMigrate) error {
	return dbutil.Chat().Model(&ChatMigrate{}).Create(chatMigrate).Error
}

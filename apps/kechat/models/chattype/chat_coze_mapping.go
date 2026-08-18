package chattype

import (
	"context"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

type ChatType string

const (
	// ChatTypeAgentApp 轻应用
	ChatTypeAgentApp ChatType = "agent_app"
	// ChatTypeForest 知识库
	ChatTypeForest ChatType = "forest"
	// ChatTypeWorkflow 工作流
	ChatTypeWorkflow ChatType = "workflow"
)

// ChatCozeMapping corekg-coze映射表
type ChatCozeMapping struct {
	gorm.Model
	// Uin 创建人ID
	Uin      uint     `gorm:"column:uin;type:int;not null;default:0;index:uin;comment:创建人ID" json:"uin"`
	Type     ChatType `gorm:"column:type;type:varchar(32);comment:类型" json:"type"`
	CoreKGID uint     `gorm:"column:corekg_id;type:int;not null;default:0;index:uin;comment:corekg侧ID" json:"corekg_id"`
	CozeID   string   `gorm:"column:coze_id;type:int;not null;default:0;index:uin;comment:coze侧ID" json:"coze_id"`
}

// TableName 表名
func (ChatCozeMapping) TableName() string {
	return TableNameChatCozeMapping
}

// CreateCozeMapping 创建映射关系
func CreateCozeMapping(ctx context.Context, mappint *ChatCozeMapping) error {
	return dbutil.Chat().Model(&ChatCozeMapping{}).Create(mappint).Error
}

// GetCozeMappingByID 根据uin获取映射关系
func GetCozeMappingByID(ctx context.Context, uin uint, chatType ChatType) ([]ChatCozeMapping, error) {
	var items []ChatCozeMapping
	err := dbutil.Chat().Model(&ChatCozeMapping{}).
		Where("uin = ?", uin).
		Where("type = ?", chatType).
		Find(&items).Error
	if err != nil {
		logs.ErrorContext(ctx, "GetCozeMappingByID error, %s", err.Error())
	}
	return items, err
}

// GetCozeMappingByCoreKGID 根据corekgID获取映射关系
func GetCozeMappingByCoreKGID(ctx context.Context, corekgID uint) ([]ChatCozeMapping, error) {
	var items []ChatCozeMapping
	err := dbutil.Chat().Model(&ChatCozeMapping{}).
		Where("corekg_id = ?", corekgID).
		Find(&items).Error
	if err != nil {
		logs.ErrorContext(ctx, "GetCozeMappingByID error, %s", err.Error())
	}
	return items, err
}

// DeleteCozeMappingByCozeID 根据coze组件ID删除映射关系
func DeleteCozeMappingByCozeID(ctx context.Context, id string, chatType ChatType) error {
	err := dbutil.Chat().Model(&ChatCozeMapping{}).
		Where("coze_id = ?", id).
		Where("type = ?", chatType).
		Unscoped().
		Delete(&ChatCozeMapping{}).Error
	if err != nil {
		logs.ErrorContext(ctx, "DeleteCozeMappingByCozeID error, %s", err.Error())
	}
	return err
}

// DeleteCozeMappingByCorekgID 根据corekgID删除映射关系
func DeleteCozeMappingByCorekgID(ctx context.Context, id uint) error {
	err := dbutil.Chat().Model(&ChatCozeMapping{}).
		Where("corekg_id = ?", id).
		Unscoped().
		Delete(&ChatCozeMapping{}).Error
	if err != nil {
		logs.ErrorContext(ctx, "DeleteCozeMappingByCorekgID error, %s", err.Error())
	}
	return err
}

// GetCozeMappingByCozeID 根据cozeID获取映射关系
func GetCozeMappingByCozeID(ctx context.Context, cozeID string, chatType ChatType) (ChatCozeMapping, error) {
	var item ChatCozeMapping
	err := dbutil.Chat().Model(&ChatCozeMapping{}).
		Where("coze_id = ?", cozeID).
		Where("type = ?", chatType).
		Find(&item).Error
	if err != nil {
		logs.ErrorContext(ctx, "GetCozeMappingByCozeID error, %s", err.Error())
	}
	return item, err
}

// GetCozeMappingByAgentID 根据agentID获取映射关系
func GetCozeMappingByAgentID(ctx context.Context, cozeID uint, chatType ChatType) (ChatCozeMapping, error) {
	var item ChatCozeMapping
	err := dbutil.Chat().Model(&ChatCozeMapping{}).
		Where("corekg_id = ?", cozeID).
		Where("type = ?", chatType).
		Find(&item).Error
	if err != nil {
		logs.ErrorContext(ctx, "GetCozeMappingByCozeID error, %s", err.Error())
	}
	return item, err
}

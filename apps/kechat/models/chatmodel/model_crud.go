package chatmodel

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/insmtx/corekg/apps/kechat/models/chat"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
)

// GetModelByID 获取模型数据chattype.ChatModel
func GetModelByID(ctx context.Context, id uint) (*chattype.ChatModel, error) {
	model := &chattype.ChatModel{}
	err := dbutil.Chat().Where("id = ?", id).First(model).Error
	if err != nil {
		logs.ErrorContextf(ctx, "GetModelByID query from db error: %v", err)
		return nil, err
	}
	return model, nil
}

type QueryModelListResponse struct {
	apiobj.QueryResponse
	Data []chattype.ChatModelDTO
}

// QueryModelList 查询模型列表
func QueryModelList(ctx context.Context, opt apiobj.PageQuery, modelList *QueryModelListResponse) error {
	query := dbutil.Chat().WithContext(ctx).Table(chattype.TableNameChatModel+" model").
		Where("model.public_type = ?", chattype.PublecTypeSystem).
		Where("model.deleted_at is NULL")

	for _, filter := range opt.Filters {
		switch filter.Field {
		case "name":
			query = query.Where("model.show_name = ?", filter.Value[0])
		default:
			logs.ErrorContextf(ctx, "[chat][QueryModelList] invalid filter field: %s", filter.Field)
			return fmt.Errorf("invalid filter field: %s", filter.Field)
		}
	}

	if err := query.Count(&modelList.Total).Error; err != nil {
		logs.ErrorContextf(ctx, "QueryModelList query count error: %v", err)
		return err
	}
	if modelList.Total == 0 {
		return nil
	}
	if len(opt.OrderBy) > 0 {
		query = query.Order(strings.Join(opt.OrderBy, ","))
	}

	query = query.Offset(opt.Offset)
	if !opt.ListAll && opt.Limit > 0 {
		query = query.Limit(opt.Limit)
	}

	err := query.Find(&modelList.Data).Error
	if err != nil {
		logs.ErrorContextf(ctx, "QueryModelList query from db error: %v", err)
		return err
	}

	// 提取所有 ChatModel 的 UserID
	var uins []uint
	for _, cm := range modelList.Data {
		uins = append(uins, cm.Uin)
	}
	users := []*UIN{}
	err = dbutil.Account().WithContext(ctx).Table("user_identification").
		Where("id IN ?", uins).
		Find(&users).Error
	if err != nil {
		logs.ErrorContextf(ctx, "QueryModelList query users error: %v", err)
		return err
	}

	for i, cm := range modelList.Data {
		for _, u := range users {
			if cm.Uin == u.ID {
				modelList.Data[i].UserName = u.Name
				break
			}
		}
	}

	// 查询最近使用的模型
	recentUsedModelEntityList, err := chat.NewChatRecentUsedModelDao().GetListByCond(ctx, &chat.ChatRecentUsedModelCond{
		Uin: opt.Uin,
	})
	modelUsedMap := make(map[uint]chattype.ChatRecentUsedModel, len(recentUsedModelEntityList))
	if err != nil {
		logs.ErrorContextf(ctx, "QueryModelList query recent used models error: %v", err)
	} else {
		for _, v := range recentUsedModelEntityList {
			modelUsedMap[v.ModelID] = v
		}
	}
	for i, cm := range modelList.Data {
		if usedModel, ok := modelUsedMap[cm.ID]; ok {
			modelList.Data[i].IsLastUsed = true
			modelList.Data[i].UsageCount = usedModel.UsageCount
			if usedModel.LastUsedAt != nil {
				modelList.Data[i].LastUsedAt = usedModel.LastUsedAt
			}
		}
	}

	return nil
}

type UIN struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// CreateModel 创建模型
func CreateModel(ctx context.Context, model *chattype.ChatModel) error {
	if err := dbutil.Chat().Create(model).Error; err != nil {
		logs.ErrorContextf(ctx, "CreateModel create model error: %v", err)
		return err
	}
	return nil
}

var (
	// 1. [a-zA-Z0-9]       : 首字符依然限制为字母或数字（避免以符号开头）
	// 2. [a-zA-Z0-9._/:-]* : 后续字符允许 冒号(:) 点(.) 下划线(_) 斜杠(/) 横杠(-)
	modelNameRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9._/:-]*$`)
)

func ValidModel(modelName string) bool {
	if len(modelName) == 0 {
		return false
	}
	return modelNameRegex.MatchString(modelName)
}

// GetProviderHeadURL 获取模型提供商的头像
func GetProviderHeadURL(provider string) string {
	switch provider {
	case "aliyun":
		return "https://prod-roc-1251908240.cos.ap-beijing.myqcloud.com/yg/chat/qwen.png"
	case "deepseek":
		return "https://prod-roc-1251908240.cos.ap-beijing.myqcloud.com/yg/chat/deepseek.jpg"
	default:
		return ""
	}
}

// DeleteModel 删除模型
func DeleteModel(ctx context.Context, id uint) error {
	if err := dbutil.Chat().
		Where("id = ?", id).Delete(&chattype.ChatModel{}).Error; err != nil {
		logs.ErrorContextf(ctx, "DeleteModel delete model error: %v", err)
		return err
	}
	return nil
}

// UpdateModel 更新模型
func UpdateModel(ctx context.Context, model *chattype.ChatModel) error {
	if err := dbutil.Chat().Save(model).Error; err != nil {
		logs.ErrorContextf(ctx, "UpdateModel update model error: %v", err)
		return err
	}
	return nil
}

// GetModelNameByID 根据模型ID查询对应的模型名称
func GetModelNameByIDs(ctx context.Context, ids []uint) ([]LLmModel, error) {
	// 处理空ids情况
	if len(ids) == 0 {
		logs.WarnContextf(ctx, "[GetModelNameByID] ids is null: %v", ids)
		return nil, nil
	}

	var model []LLmModel
	if err := dbutil.Chat().Table(chattype.TableNameChatModel).
		Select("id,model_name").
		Where("id IN (?)", ids).
		Find(&model).Error; err != nil {
		logs.ErrorContextf(ctx, "[chat] [GetModelNameByID] failed to find llm_model data for ids %v, error: %v", ids, err)
		return nil, fmt.Errorf("failed to find model names: %w", err)
	}

	return model, nil
}

type LLmModel struct {
	ID          uint   `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

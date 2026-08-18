package chatsession

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
)

// CreateSession 创建session
func CreateSession(ctx context.Context, sess *chattype.ChatSession) error {
	if sess.Name == "" {
		sess.Name = chattype.DefaultSessionName
	}
	err := dbutil.Chat().WithContext(ctx).Create(sess).Error
	if err != nil {
		logs.ErrorContextf(ctx, "CreateChatSeesion error: %v", err)
		return err
	}
	return nil
}

// GetChatSession 获取会话
func GetChatSession(ctx context.Context, uin, id uint) (*chattype.ChatSession, error) {
	sess := &chattype.ChatSession{}
	err := dbutil.Chat().WithContext(ctx).Where("uin = ? and id = ?", uin, id).First(sess).Error
	if err != nil {
		logs.ErrorContextf(ctx, "GetChatSession error: %v", err)
		return nil, err
	}
	return sess, nil
}

// QueryListChatSessions 查询会话列表
func QueryListChatSessions(ctx context.Context, opt apiobj.PageQuery, projectID int, sessionList *QueryListChatSessionsResponse) error {
	// 初始化查询
	query := dbutil.Chat().WithContext(ctx).Table(chattype.TableNameChatSessions + " as session").
		// Select("chat_sessions.*").
		// Joins("LEFT JOIN chat_agent ON chat_sessions.base_agent_id = chat_agent.id AND chat_agent.deleted_at IS NULL").
		// Where("session.uin = ?", opt.Uin).
		Where("session.deleted_at IS NULL")
	// Where("chat_agent.deleted_at IS NULL")
	if projectID == 0 {
		//if projectID == 0 : fetch all session
		query = query.Where("session.uin = ?", opt.Uin)
	} else if projectID > 0 {
		//if projectID > 0 : fetch session in project
		query = query.
			Where("session.subject_id = ?", projectID).
			Where("session.company_id = ?", opt.CompanyID)
	} else {
		//if projectID < 0 : fetch all free session
		query = query.
			Where("session.uin = ?", opt.Uin).
			Where("session.company_id = ?", opt.CompanyID).
			Where("session.subject_id = ?", 0).
			Where("session.resource_type NOT IN (?)", []chattype.ResourceType{chattype.ResourceTypeAgent})
	}

	// 添加过滤器
	for _, filter := range opt.Filters {
		switch filter.Field {
		case "name":
			query = query.Where("session.name = ?", filter.Value[0])
		case "resource_id":
			query = query.Where("session.base_agent_id = ?", filter.Value[0])
		case "resource_type":
			query = query.Where("session.resource_type IN (?)", filter.Value)
		default:
			logs.ErrorContextf(ctx, "[chat][QueryListChatSessions] invalid filter field: %s", filter.Field)
			return fmt.Errorf("invalid filter field: %s", filter.Field)
		}
	}

	// 查询总记录数
	if err := query.Count(&sessionList.Total).Error; err != nil {
		logs.ErrorContextf(ctx, "QueryListChatSessions error: %v", err)
		return err
	}
	if sessionList.Total == 0 {
		return nil
	}

	// 排序逻辑：置顶优先，按更新时间倒序
	query = query.Order("session.is_top DESC").Order("session.updated_at DESC,session.created_at DESC")

	if len(opt.OrderBy) > 0 {
		query = query.Order(strings.Join(opt.OrderBy, ","))
	}

	// 分页逻辑
	query = query.Offset(opt.Offset)
	if !opt.ListAll && opt.Limit > 0 {
		query = query.Limit(opt.Limit)
	}

	// 查询数据
	err := query.Find(&sessionList.Data).Error
	if err != nil {
		logs.ErrorContextf(ctx, "QueryListChatSessions error: %v", err)
		return err
	}

	return nil
}

type QueryListChatSessionsResponse struct {
	apiobj.QueryResponse
	Data []*chattype.ChatSession
}

// UpdateChatSession 更新会话
func UpdateChatSession(ctx context.Context, sess *chattype.ChatSession) error {
	err := dbutil.Chat().WithContext(ctx).Save(sess).Error
	if err != nil {
		logs.ErrorContextf(ctx, "UpdateChatSession error: %v", err)
		return err
	}
	return nil
}

// UpdateSessionName 取前10个字修改会话名称 暂时废弃
func UpdateSessionName(ctx context.Context, sess *chattype.ChatSession, question string) {
	if question == "" {
		return
	}
	// 会话命名
	if sess.Name == chattype.DefaultSessionName {
		runes := []rune(question)
		sess.Name = question
		if len(runes) > 10 {
			sess.Name = string(runes[:10])
		}
		logs.InfoContextf(ctx, "[UpdateSessionName] session.Name: %v", sess.Name)
	}
	UpdateChatSession(ctx, sess)
}

// UpdateSessionNameWithLLM 根据问答结果由模型总结会话名称
func UpdateSessionNameWithLLM(ctx context.Context, sess *chattype.ChatSession, question, answer string) {
	defer func() {
		sess.UpdatedAt = time.Now()
		if err := UpdateChatSession(ctx, sess); err != nil {
			logs.ErrorContextf(ctx, "UpdateSessionNameWithLLM error: %v", err)
		}
	}()
	if question == "" {
		return
	}
	// 会话命名
	if sess.Name == chattype.DefaultSessionName {
		name, err := GetLLmSessionName(ctx, question, answer)
		if err != nil {
			logs.ErrorContextf(ctx, "[UpdateSessionNameWithLLM] session.Name: %s,err:%v", name, err)
			return
		}
		sess.Name = name
		if len([]rune(sess.Name)) > 10 {
			sess.Name = string([]rune(sess.Name)[:10])
		}
		logs.InfoContextf(ctx, "[UpdateSessionNameWithLLM] session.Name: %v", sess.Name)
	}
}

// DeleteSession 删除会话
func DeleteSession(ctx context.Context, id uint) error {
	err := dbutil.Chat().WithContext(ctx).Where("id = ?", id).Delete(&chattype.ChatSession{}).Error
	if err != nil {
		logs.ErrorContextf(ctx, "DeleteSession error: %v", err)
		return err
	}
	return nil
}

// IsExistSetTopChatSession 判断会话是否置顶
func IsExistSetTopChatSession(ctx context.Context, uin, id uint) (bool, error) {
	var count int64
	err := dbutil.Chat().WithContext(ctx).Model(&chattype.ChatSession{}).
		Where("uin = ? AND id = ? AND is_top = 1", uin, id).
		Count(&count).Error
	if err != nil {
		logs.ErrorContextf(ctx, "IsExistSetTopChatSession err: %v", err)
		return false, err
	}
	return count > 0, nil
}

// CancelSetTopChatSession 取消置顶会话
func CancelSetTopChatSession(ctx context.Context, uin, id uint) error {
	var sess chattype.ChatSession
	err := dbutil.Chat().WithContext(ctx).Unscoped().
		Where("uin = ? AND id = ?", uin, id).
		First(&sess).Error
	if err != nil {
		logs.ErrorContextf(ctx, "CancelSetTopChatSession err: %v", err)
		return err
	}
	sess.IsTop = -1
	err = dbutil.Chat().WithContext(ctx).Save(&sess).Error
	if err != nil {
		logs.ErrorContextf(ctx, "CancelSetTopChatSession err: %v", err)
		return err
	}
	return nil
}

// CreateSetTopChatSession 创建置顶会话
func CreateSetTopChatSession(ctx context.Context, uin, id uint) error {
	var sess chattype.ChatSession
	err := dbutil.Chat().WithContext(ctx).Unscoped().
		Where("uin = ? AND id = ?", uin, id).
		First(&sess).Error
	if err != nil {
		logs.ErrorContextf(ctx, "CreateSetTopChatSession err: %v", err)
		return err
	}
	sess.IsTop = 1
	err = dbutil.Chat().WithContext(ctx).Save(&sess).Error
	if err != nil {
		logs.ErrorContextf(ctx, "CreateSetTopChatSession err: %v", err)
		return err
	}
	return nil
}

func GetExternalSessionByID(ctx context.Context, externalID string, versionID uint) (*chattype.ChatSession, error) {
	var res *chattype.ChatSession
	if err := dbutil.Chat().WithContext(ctx).
		Where("external_id = ?", externalID).
		Where("agent_version = ?", versionID).
		First(&res).Error; err != nil {
		return nil, err
	}
	return res, nil
}

// GetFileLastSession 获取最近一次会话
func GetFileLastSession(ctx context.Context, uin, file_id uint) (*chattype.ChatSession, error) {
	out := &chattype.ChatSession{}
	if err := dbutil.Chat().WithContext(ctx).Where("uin = ?", uin).
		Where("file_id = ?", file_id).
		Where("resource_type = ?", chattype.ResourceTypeFile).
		Where("deleted_at is null").
		Order("updated_at desc").
		First(out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

type SessionInfo struct {
	chattype.ChatSession
	ForestNames     []SourceInfo `json:"forest_names"`
	FileNames       []SourceInfo `json:"file_names"`
	ExcelNames      []SourceInfo `json:"excel_names"`       // excel 列表
	ExcelSheetNames []SourceInfo `json:"excel_sheet_names"` // excel sheet 列表
	// DBNames 数据库名称列表
	DBNames []SourceInfo `json:"db_names"`
	// DBTableNames 数据库表名称列表
	DBTableNames []SourceInfo `json:"db_table_names"`
}

type SourceInfo struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// GetSessionInfo will return all related name-like info about this session
func GetSessionInfo(ctx *gin.Context, id uint) (*SessionInfo, error) {
	var session *chattype.ChatSession
	if err := dbutil.Chat().First(&session, id).Error; err != nil {
		return nil, err
	}

	si := &SessionInfo{
		ChatSession: *session,
	}
	if si.ModelID == 0 {
		si.ModelID = 1
	}

	var result []SourceInfo
	switch session.ResourceType {
	case chattype.ResourceTypeForest:
		if len(session.ForestIDList) > 0 {
			if err := dbutil.Knownow().Table(foresttype.TableNameKnownowForest).
				Select("id,name").
				Where("id in ?", session.ForestIDList.Slice()).
				Find(&result).
				Error; err != nil {
				return nil, err
			}
			si.ForestNames = result
		}
	case chattype.ResourceTypeFileList, chattype.ResourceTypeDirList:
		if len(session.FileIDList) > 0 {
			if err := dbutil.Knownow().Table(foresttype.TableNameKnownowForestFile).
				Select("id,name").
				Where("id in ?", session.FileIDList.Slice()).
				Find(&result).
				Error; err != nil {
				return nil, err
			}
			si.FileNames = result
		}
	case chattype.ResourceTypeFile:
		if session.FileID > 0 {
			if err := dbutil.Knownow().Table(foresttype.TableNameKnownowForestFile).
				Select("id,name").
				Where("id = ?", session.FileID).
				Find(&result).
				Error; err != nil {
				return nil, err
			}
			si.FileNames = result
		}
	case chattype.ResourceTypeExcelList, chattype.ResourceTypeReactExcelList:
		if len(session.ExcelIDList) > 0 {
			forestFileEntityList, err := forest.NewForestFileDao().GetListByCond(ctx, &forest.ForestFileCond{
				IDs: session.ExcelIDList.Slice(),
			})
			if err != nil {
				return nil, err
			}
			excelNames := make([]SourceInfo, 0, len(forestFileEntityList))
			for _, v := range forestFileEntityList {
				excelNames = append(excelNames, SourceInfo{
					ID:   v.ID,
					Name: v.Name,
				})
			}
			si.ExcelNames = excelNames
		}
	case chattype.ResourceTypeExcelSheetList:
		if len(session.ExcelSheetIDList) > 0 {
			excelSheetEntityList, err := forest.NewForestExcelSheetDao().GetListByCond(ctx, &forest.ForestExcelSheetCond{
				IDS: session.ExcelSheetIDList.Slice(),
			})
			if err != nil {
				return nil, err
			}
			sheetNames := make([]SourceInfo, 0, len(excelSheetEntityList))
			for _, v := range excelSheetEntityList {
				sheetNames = append(sheetNames, SourceInfo{
					ID:   v.ID,
					Name: v.SheetName,
				})
			}
			si.ExcelSheetNames = sheetNames

		}
	case chattype.ResourceTypeDBList:
		if len(session.DBList) > 0 {
			dbNames := make([]SourceInfo, 0, len(session.DBList))
			for _, v := range session.DBList.Slice() {
				dbNames = append(dbNames, SourceInfo{
					Name: v,
				})
			}
			si.DBNames = dbNames
		}
	case chattype.ResourceTypeDBTableList:
		if len(session.DBTableList) > 0 {
			dbTableNames := make([]SourceInfo, 0, len(session.DBTableList))
			for _, v := range session.DBTableList.Slice() {
				dbTableNames = append(dbTableNames, SourceInfo{
					Name: v,
				})
			}
			si.DBTableNames = dbTableNames
		}
	case chattype.ResourceTypeModel, chattype.ResourceTypeAgent:
	case chattype.ResourceTypeExternalData:
	case chattype.ResourceTypeGraphSearch:
	default:
		return nil, fmt.Errorf("unknown session type: %v", session.ResourceType)
	}
	return si, nil
}

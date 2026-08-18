package keqa

import (
	"context"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
)

// QueryQASessionListResponse 表单类型响应
type QueryQASessionListResponse struct {
	apiobj.QueryResponse

	Data []*foresttype.KnownowQASession
}

// QueryQASessionList 查询知识问答会话列表
func QueryQASessionList(ctx context.Context, opt *apiobj.PageQuery, resp *QueryQASessionListResponse) error {
	query := dbutil.Knownow().WithContext(ctx).Table(foresttype.TableNameKnownowQASession+" qs").
		Where("qs.deleted_at is null").
		Where("qs.name != ?", foresttype.DefaultSessionName)

	if opt.Uin != 0 {
		query = query.Where("uin=?", opt.Uin)
	}

	for _, filter := range opt.Filters {
		switch filter.Field {
		case "type":
			query = query.Where("qs.type in (?) ", filter.Value)
		case "name":
			query = query.Where("qs.name like ?", "%"+filter.Value[0]+"%")
		default:
			logs.WarnContextf(ctx, "[hscustomer][QueryCustomerList] invalid filter field: %s", filter.Field)
			return fmt.Errorf("invalid filter field: %s", filter.Field)
		}
	}

	if !opt.BeginTime.IsZero() {
		query = query.Where("qs.created_at >= ?", opt.BeginTime)
	}
	if !opt.EndTime.IsZero() {
		query = query.Where("qs.created_at <= ?", opt.EndTime)
	}

	if err := query.Count(&resp.Total).Error; err != nil {
		return err
	}
	if resp.Total == 0 {
		return nil
	}

	if len(opt.OrderBy) > 0 {
		query = query.Order(strings.Join(opt.OrderBy, ","))
	} else {
		query = query.Order("qs.created_at")
	}

	query = query.Offset(opt.Offset)
	if !opt.ListAll && opt.Limit > 0 {
		query = query.Limit(opt.Limit)
	}
	err := query.Find(&resp.Data).Error
	if err != nil {
		return err
	}
	return nil
}

// CreateForestSession 创建本地知识森林会话
func CreateForestSession(session *foresttype.KnownowQASession) (*foresttype.KnownowQASession, error) {
	err := dbutil.Knownow().Create(session).Error
	return session, err
}

// GetCustomerSession 获取session
func GetCustomerSession(uin, sessionID uint) (*foresttype.KnownowQASession, error) {
	out := &foresttype.KnownowQASession{}
	if err := dbutil.Knownow().Where("uin = ?", uin).
		First(out, sessionID).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// GetFileLastSession 获取最近一次会话
func GetFileLastSession(uin, file_id uint) (*foresttype.KnownowQASession, error) {
	out := &foresttype.KnownowQASession{}
	if err := dbutil.Knownow().Where("uin = ?", uin).
		Where("file_id = ?", file_id).
		Where("type = ?", foresttype.KnownowQASessionTypeFile).
		Where("deleted_at is null").
		Order("updated_at desc").
		First(out).Error; err != nil {
		return nil, err
	}
	return out, nil
}

// ModifySession 修改会话
func ModifySession(ctx context.Context, session *foresttype.KnownowQASession) error {
	result := dbutil.Knownow().
		Table(foresttype.TableNameKnownowQASession).
		Where("id = ?", session.ID).
		Updates(map[string]interface{}{
			"name": session.Name,
		})
	if result.Error != nil {
		logs.ErrorContextf(ctx, "[knownow-forest][ModifySession] failed to modify session: %v", result.Error)
		return result.Error
	}
	return nil
}

// DeleteSession 删除会话
func DeleteSession(ctx context.Context, id uint) error {
	// 开启事务
	tx := dbutil.Knownow().Begin()
	if tx.Error != nil {
		return tx.Error
	}

	// 删除会话
	if err := tx.Table(foresttype.TableNameKnownowQASession).
		Where("id = ?", id).
		Delete(&foresttype.KnownowQASession{}).Error; err != nil {
		tx.Rollback()
		logs.ErrorContextf(ctx, "[knownow-forest][DeleteSession] failed to delete session with id %d: %v", id, err)
		return err
	}

	// 删除会话下的问题
	if err := tx.Table(foresttype.TableNameKnownowForestQA).
		Where("session_id = ?", id).
		Delete(&foresttype.KnownowForestQA{}).Error; err != nil {
		tx.Rollback()
		logs.ErrorContextf(ctx, "[knownow-forest][DeleteSession] failed to delete questions for session with id %d: %v", id, err)
		return err
	}

	// 提交事务
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		logs.ErrorContextf(ctx, "[knownow-forest][DeleteSession] failed to commit transaction: %v", err)
		return err
	}

	return nil
}

type SessionInfo struct {
	foresttype.KnownowQASession
	ForestNames     []SourceInfo `json:"forest_names"`
	FileNames       []SourceInfo `json:"file_names"`
	ExcelNames      []SourceInfo `json:"excel_names"`       // excel 列表
	ExcelSheetNames []SourceInfo `json:"excel_sheet_names"` // excel sheet 列表
	LLMModelName    string       `gorm:"column:show_name" json:"llm_model_name"`
}

type SourceInfo struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// GetSessionInfo will return all related name-like info about this session
func GetSessionInfo(ctx *gin.Context, id uint) (*SessionInfo, error) {
	var session *foresttype.KnownowQASession
	if err := dbutil.Knownow().First(&session, id).Error; err != nil {
		return nil, err
	}

	si := &SessionInfo{
		KnownowQASession: *session,
	}
	if si.LLMModelID == 0 {
		si.LLMModelID = 1
	}

	if err := dbutil.Chat().Table(chattype.TableNameChatModel).
		Select("show_name").
		Find(&si.LLMModelName, "id = ?", si.LLMModelID).
		Error; err != nil {
		return nil, err
	}

	var result []SourceInfo
	switch session.Type {
	case foresttype.KnownowQASessionTypeForest:
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
	case foresttype.KnownowQASessionTypeFileList, foresttype.KnownowQASessionTypeDirList:
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
	case foresttype.KnownowQASessionTypeFile:
		if session.FileID > 0 {
			if err := dbutil.Knownow().Table(foresttype.TableNameKnownowForestFile).
				Select("id,name").
				Where("id = ?", session.FileID).
				Find(&result).
				Error; err != nil {
				return nil, err
			}
		}
	case foresttype.KnownowQASessionTypeExcelList:
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
	case foresttype.KnownowQASessionTypeExcelSheetList:
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
	default:
		return nil, fmt.Errorf("unknown session type: %v", session.Type)
	}
	return si, nil
}

package chat

import (
	"context"
	"fmt"
	"time"

	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

type ChatSessionsCond struct {
	BaseCond
	Filters         []apiobj.Filter
	ID              uint
	FileID          uint
	SubjectID       uint
	SubjectIDGTZero bool
	FileIDGTZero    bool
	SubjectIDs      []uint
	AgentID         uint
}

type ChatSessionsDao struct {
	BaseModel
}

func NewChatSessionsDao() *ChatSessionsDao {
	return &ChatSessionsDao{}
}

func (dao *ChatSessionsDao) TableName() string {
	return chattype.TableNameChatSessions
}

func (dao *ChatSessionsDao) WithTx(db *gorm.DB) *ChatSessionsDao {
	return &ChatSessionsDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *ChatSessionsDao) Insert(ctx context.Context, entity *chattype.ChatSession) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[ChatSessionsDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *ChatSessionsDao) BatchInsert(ctx context.Context, entityList chattype.ChatSessionList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[ChatSessionsDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return fmt.Errorf("[ChatSessionsDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *ChatSessionsDao) UpdateByID(ctx context.Context, id uint, entity *chattype.ChatSession) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[ChatSessionsDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *ChatSessionsDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[ChatSessionsDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *ChatSessionsDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_time": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[ChatSessionsDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *ChatSessionsDao) GetByID(ctx context.Context, id uint) (*chattype.ChatSession, error) {
	var entity chattype.ChatSession
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[ChatSessionsDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *ChatSessionsDao) GetByCond(ctx context.Context, cond *ChatSessionsCond) (*chattype.ChatSession, error) {
	var entity chattype.ChatSession
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[ChatSessionsDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *ChatSessionsDao) GetListByCond(ctx context.Context, cond *ChatSessionsCond) (chattype.ChatSessionList, error) {
	var entityList chattype.ChatSessionList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[ChatSessionsDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

func (dao *ChatSessionsDao) GetPageListByCond(ctx context.Context, cond *ChatSessionsCond) (chattype.ChatSessionList, int64, error) {
	db := dao.DB(ctx).Model(&chattype.ChatSession{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[ChatSessionsDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList chattype.ChatSessionList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[ChatSessionsDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *ChatSessionsDao) CountByCond(ctx context.Context, cond *ChatSessionsCond) (int64, error) {
	db := dao.DB(ctx).Model(&chattype.ChatSession{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[ChatSessionsDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

func (dao *ChatSessionsDao) BuildCondition(db *gorm.DB, cond *ChatSessionsCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}
	if cond.FileID > 0 {
		db = db.Where(fmt.Sprintf("%s.file_id = ?", dao.TableName()), cond.FileID)
	}
	if cond.SubjectID > 0 {
		db = db.Where(fmt.Sprintf("%s.subject_id = ?", dao.TableName()), cond.SubjectID)
	}
	if cond.SubjectIDGTZero {
		db = db.Where(fmt.Sprintf("%s.subject_id > ?", dao.TableName()), 0)
	}
	if cond.FileIDGTZero {
		db = db.Where(fmt.Sprintf("%s.file_id > ?", dao.TableName()), 0)
	}

	if len(cond.SubjectIDs) > 0 {
		db = db.Where(fmt.Sprintf("%s.subject_id IN ?", dao.TableName()), cond.SubjectIDs)
	}
	if cond.AgentID > 0 {
		db = db.Where(fmt.Sprintf("%s.agent_id = ?", dao.TableName()), cond.AgentID)
	}
}

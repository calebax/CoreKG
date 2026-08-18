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

type ChatChartCanvasCond struct {
	BaseCond
	Filters     []apiobj.Filter
	ID          uint
	SubjectType chattype.SessionSubjectType
	SubjectID   uint
}

type ChatChartCanvasDao struct {
	BaseModel
}

func NewChatChartCanvasDao() *ChatChartCanvasDao {
	return &ChatChartCanvasDao{}
}

func (dao *ChatChartCanvasDao) TableName() string {
	return chattype.TableNameChatChartCanvas
}

func (dao *ChatChartCanvasDao) WithTx(db *gorm.DB) *ChatChartCanvasDao {
	return &ChatChartCanvasDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *ChatChartCanvasDao) Insert(ctx context.Context, entity *chattype.ChatChartCanvas) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[ChatChartCanvasDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *ChatChartCanvasDao) BatchInsert(ctx context.Context, entityList chattype.ChatChartCanvasList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[ChatChartCanvasDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return fmt.Errorf("[ChatChartCanvasDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *ChatChartCanvasDao) UpdateByID(ctx context.Context, id uint, entity *chattype.ChatChartCanvas) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[ChatChartCanvasDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *ChatChartCanvasDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[ChatChartCanvasDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *ChatChartCanvasDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_time": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[ChatChartCanvasDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *ChatChartCanvasDao) GetByID(ctx context.Context, id uint) (*chattype.ChatChartCanvas, error) {
	var entity chattype.ChatChartCanvas
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[ChatChartCanvasDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *ChatChartCanvasDao) GetByCond(ctx context.Context, cond *ChatChartCanvasCond) (*chattype.ChatChartCanvas, error) {
	var entity chattype.ChatChartCanvas
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[ChatChartCanvasDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *ChatChartCanvasDao) GetListByCond(ctx context.Context, cond *ChatChartCanvasCond) (chattype.ChatChartCanvasList, error) {
	var entityList chattype.ChatChartCanvasList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[ChatChartCanvasDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

func (dao *ChatChartCanvasDao) GetPageListByCond(ctx context.Context, cond *ChatChartCanvasCond) (chattype.ChatChartCanvasList, int64, error) {
	db := dao.DB(ctx).Model(&chattype.ChatChartCanvas{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[ChatChartCanvasDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList chattype.ChatChartCanvasList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[ChatChartCanvasDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *ChatChartCanvasDao) CountByCond(ctx context.Context, cond *ChatChartCanvasCond) (int64, error) {
	db := dao.DB(ctx).Model(&chattype.ChatChartCanvas{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[ChatChartCanvasDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

func (dao *ChatChartCanvasDao) BuildCondition(db *gorm.DB, cond *ChatChartCanvasCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}
	if cond.SubjectType != "" {
		query := fmt.Sprintf("%s.subject_type = ?", dao.TableName())
		db.Where(query, cond.SubjectType)
	}
	if cond.SubjectID > 0 {
		query := fmt.Sprintf("%s.subject_id = ?", dao.TableName())
		db.Where(query, cond.SubjectID)
	}
}

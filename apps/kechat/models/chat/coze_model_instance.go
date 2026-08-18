package chat

import (
	"context"
	"fmt"
	"time"

	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

type ModelInstanceCond struct {
	BaseCond
	Filters []apiobj.Filter
	ID      uint
}

type CozeModelInstanceDao struct {
	BaseModel
}

func NewCozeModelInstanceDao() *CozeModelInstanceDao {
	return &CozeModelInstanceDao{
		BaseModel: BaseModel{DBClient: dbutil.Coze()},
	}
}

func (dao *CozeModelInstanceDao) TableName() string {
	return chattype.TableNameModelInstance
}

func (dao *CozeModelInstanceDao) WithTx(db *gorm.DB) *CozeModelInstanceDao {
	return &CozeModelInstanceDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

// DB returns Coze DB by default to avoid writing model_instance into the chat database.
func (dao *CozeModelInstanceDao) DB(ctx context.Context) *gorm.DB {
	if dao.DBClient != nil {
		return dao.DBClient.WithContext(ctx)
	}
	return dbutil.Coze().WithContext(ctx)
}

func (dao *CozeModelInstanceDao) Insert(ctx context.Context, entity *chattype.CozeModelInstance) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[ModelInstanceDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *CozeModelInstanceDao) BatchInsert(ctx context.Context, entityList chattype.ModelInstanceList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[ModelInstanceDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return fmt.Errorf("[ModelInstanceDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *CozeModelInstanceDao) UpdateByID(ctx context.Context, id uint, entity *chattype.CozeModelInstance) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[ModelInstanceDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *CozeModelInstanceDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[ModelInstanceDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *CozeModelInstanceDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[ModelInstanceDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *CozeModelInstanceDao) GetByID(ctx context.Context, id uint) (*chattype.CozeModelInstance, error) {
	var entity chattype.CozeModelInstance
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[ModelInstanceDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *CozeModelInstanceDao) GetByCond(ctx context.Context, cond *ModelInstanceCond) (*chattype.CozeModelInstance, error) {
	var entity chattype.CozeModelInstance
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[ModelInstanceDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *CozeModelInstanceDao) GetListByCond(ctx context.Context, cond *ModelInstanceCond) (chattype.ModelInstanceList, error) {
	var entityList chattype.ModelInstanceList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[ModelInstanceDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

func (dao *CozeModelInstanceDao) GetPageListByCond(ctx context.Context, cond *ModelInstanceCond) (chattype.ModelInstanceList, int64, error) {
	db := dao.DB(ctx).Model(&chattype.CozeModelInstance{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[ModelInstanceDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList chattype.ModelInstanceList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[ModelInstanceDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *CozeModelInstanceDao) CountByCond(ctx context.Context, cond *ModelInstanceCond) (int64, error) {
	db := dao.DB(ctx).Model(&chattype.CozeModelInstance{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[ModelInstanceDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

func (dao *CozeModelInstanceDao) BuildCondition(db *gorm.DB, cond *ModelInstanceCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}
}

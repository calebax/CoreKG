package forest

import (
	"context"
	"fmt"
	"time"

	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

type KeUinMessageCond struct {
	BaseCond
	Filters []apiobj.Filter
	ID      uint
}

type KeUinMessageDao struct {
	BaseModel
}

func NewKeUinMessageDao() *KeUinMessageDao {
	return &KeUinMessageDao{}
}

func (dao *KeUinMessageDao) TableName() string {
	return foresttype.TableNameKeUinMessage
}

func (dao *KeUinMessageDao) WithTx(db *gorm.DB) *KeUinMessageDao {
	return &KeUinMessageDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *KeUinMessageDao) Insert(ctx context.Context, entity *foresttype.KeUinMessage) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[KeUinMessageDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *KeUinMessageDao) BatchInsert(ctx context.Context, entityList foresttype.KeUinMessageList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[KeUinMessageDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return fmt.Errorf("[KeUinMessageDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *KeUinMessageDao) UpdateByID(ctx context.Context, id uint, entity *foresttype.KeUinMessage) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[KeUinMessageDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *KeUinMessageDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[KeUinMessageDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *KeUinMessageDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[KeUinMessageDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *KeUinMessageDao) DeleteByIDs(ctx context.Context, ids []uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("id in ?", ids).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[KeUinMessageDao] Delete fail, ids:%v, err: %v", ids, err)
	}
	return nil
}

func (dao *KeUinMessageDao) DeleteByUin(ctx context.Context, uin uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("uin = ?", uin).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[KeUinMessageDao] Delete fail, uin:%d, err: %v", uin, err)
	}
	return nil
}

func (dao *KeUinMessageDao) GetByID(ctx context.Context, id uint) (*foresttype.KeUinMessage, error) {
	var entity foresttype.KeUinMessage
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[KeUinMessageDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *KeUinMessageDao) GetByCond(ctx context.Context, cond *KeUinMessageCond) (*foresttype.KeUinMessage, error) {
	var entity foresttype.KeUinMessage
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[KeUinMessageDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *KeUinMessageDao) GetListByCond(ctx context.Context, cond *KeUinMessageCond) (foresttype.KeUinMessageList, error) {
	var entityList foresttype.KeUinMessageList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[KeUinMessageDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

func (dao *KeUinMessageDao) GetPageListByCond(ctx context.Context, cond *KeUinMessageCond) (foresttype.KeUinMessageList, int64, error) {
	db := dao.DB(ctx).Model(&foresttype.KeUinMessage{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[KeUinMessageDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList foresttype.KeUinMessageList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[KeUinMessageDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *KeUinMessageDao) CountByCond(ctx context.Context, cond *KeUinMessageCond) (int64, error) {
	db := dao.DB(ctx).Model(&foresttype.KeUinMessage{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	if cond.Filters != nil {
		for _, filter := range cond.Filters {
			switch filter.Field {
			case "read_status":
				db.Where("read_status IN (?)", filter.Value)
			default:
				logs.WarnContextf(ctx, "[KeUinMessageDao] CountByCond unknown filter field: %s", filter.Field)
			}
		}
	}
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[KeUinMessageDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

func (dao *KeUinMessageDao) BuildCondition(db *gorm.DB, cond *KeUinMessageCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}
}

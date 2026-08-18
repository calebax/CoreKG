package account

import (
	"context"
	"fmt"
	"time"

	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

type ApiKeyCond struct {
	BaseCond
	Filters []apiobj.Filter
	ID      uint
}

type APIKeyDao struct {
	BaseModel
}

func NewApiKeyDao() *APIKeyDao {
	return &APIKeyDao{}
}

func (dao *APIKeyDao) TableName() string {
	return accounttype.TableNameAPIKey
}

func (dao *APIKeyDao) WithTx(db *gorm.DB) *APIKeyDao {
	return &APIKeyDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *APIKeyDao) Insert(ctx context.Context, entity *accounttype.APIKey) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[APIKeyDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *APIKeyDao) BatchInsert(ctx context.Context, entityList accounttype.APIKeyList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[APIKeyDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return fmt.Errorf("[APIKeyDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *APIKeyDao) UpdateByID(ctx context.Context, id uint, entity *accounttype.APIKey) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[APIKeyDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *APIKeyDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[APIKeyDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *APIKeyDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[APIKeyDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *APIKeyDao) GetByID(ctx context.Context, id uint) (*accounttype.APIKey, error) {
	var entity accounttype.APIKey
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[APIKeyDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *APIKeyDao) GetByCond(ctx context.Context, cond *ApiKeyCond) (*accounttype.APIKey, error) {
	var entity accounttype.APIKey
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[APIKeyDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *APIKeyDao) GetListByCond(ctx context.Context, cond *ApiKeyCond) (accounttype.APIKeyList, error) {
	var entityList accounttype.APIKeyList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[APIKeyDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

func (dao *APIKeyDao) GetPageListByCond(ctx context.Context, cond *ApiKeyCond) (accounttype.APIKeyList, int64, error) {
	db := dao.DB(ctx).Model(&accounttype.APIKey{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[APIKeyDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList accounttype.APIKeyList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[APIKeyDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *APIKeyDao) CountByCond(ctx context.Context, cond *ApiKeyCond) (int64, error) {
	db := dao.DB(ctx).Model(&accounttype.APIKey{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[APIKeyDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

func (dao *APIKeyDao) BuildCondition(db *gorm.DB, cond *ApiKeyCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}
}

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

type UserCond struct {
	BaseCond
	Filters []apiobj.Filter
	ID      uint
}

type UserDao struct {
	BaseModel
}

func NewUserDao() *UserDao {
	return &UserDao{}
}

func (dao *UserDao) TableName() string {
	return accounttype.TableNameUser
}

func (dao *UserDao) WithTx(db *gorm.DB) *UserDao {
	return &UserDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *UserDao) Insert(ctx context.Context, entity *accounttype.User) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[UserDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *UserDao) BatchInsert(ctx context.Context, entityList accounttype.UserList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[UserDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return fmt.Errorf("[UserDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *UserDao) UpdateByID(ctx context.Context, id uint, entity *accounttype.User) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[UserDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *UserDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[UserDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *UserDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[UserDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *UserDao) GetByID(ctx context.Context, id uint) (*accounttype.User, error) {
	var entity accounttype.User
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[UserDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *UserDao) GetByCond(ctx context.Context, cond *UserCond) (*accounttype.User, error) {
	var entity accounttype.User
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[UserDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *UserDao) GetListByCond(ctx context.Context, cond *UserCond) (accounttype.UserList, error) {
	var entityList accounttype.UserList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[UserDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

func (dao *UserDao) GetPageListByCond(ctx context.Context, cond *UserCond) (accounttype.UserList, int64, error) {
	db := dao.DB(ctx).Model(&accounttype.User{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[UserDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList accounttype.UserList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[UserDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *UserDao) CountByCond(ctx context.Context, cond *UserCond) (int64, error) {
	db := dao.DB(ctx).Model(&accounttype.User{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[UserDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

func (dao *UserDao) BuildCondition(db *gorm.DB, cond *UserCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}
}

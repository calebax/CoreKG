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

type KeProjectCond struct {
	BaseCond
	Filters         []apiobj.Filter
	ID              uint
	NameLike        string
	CompanyID       uint
	Uin             uint
	ProjectTypeList []foresttype.ProjectType
}

type KeProjectDao struct {
	BaseModel
}

func NewKeProjectDao() *KeProjectDao {
	return &KeProjectDao{}
}

func (dao *KeProjectDao) TableName() string {
	return foresttype.TableNameKeProject
}

func (dao *KeProjectDao) WithTx(db *gorm.DB) *KeProjectDao {
	return &KeProjectDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *KeProjectDao) Insert(ctx context.Context, entity *foresttype.KnownowProject) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[KeProjectDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *KeProjectDao) BatchInsert(ctx context.Context, entityList foresttype.KnownowProjectList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[KeProjectDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return fmt.Errorf("[KeProjectDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *KeProjectDao) UpdateByID(ctx context.Context, id uint, entity *foresttype.KnownowProject) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[KeProjectDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *KeProjectDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[KeProjectDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *KeProjectDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[KeProjectDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *KeProjectDao) GetByID(ctx context.Context, id uint) (*foresttype.KnownowProject, error) {
	var entity foresttype.KnownowProject
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[KeProjectDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *KeProjectDao) GetByCond(ctx context.Context, cond *KeProjectCond) (*foresttype.KnownowProject, error) {
	var entity foresttype.KnownowProject
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[KeProjectDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *KeProjectDao) GetListByCond(ctx context.Context, cond *KeProjectCond) (foresttype.KnownowProjectList, error) {
	var entityList foresttype.KnownowProjectList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[KeProjectDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

func (dao *KeProjectDao) GetPageListByCond(ctx context.Context, cond *KeProjectCond) (foresttype.KnownowProjectList, int64, error) {
	db := dao.DB(ctx).Model(&foresttype.KnownowProject{}).Table(dao.TableName())
	db.Order(fmt.Sprintf("%s.sort DESC", dao.TableName()))
	dao.BuildCondition(db, cond)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[KeProjectDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList foresttype.KnownowProjectList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[KeProjectDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *KeProjectDao) CountByCond(ctx context.Context, cond *KeProjectCond) (int64, error) {
	db := dao.DB(ctx).Model(&foresttype.KnownowProject{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[KeProjectDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

func (dao *KeProjectDao) BuildCondition(db *gorm.DB, cond *KeProjectCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}
	if cond.NameLike != "" {
		query := fmt.Sprintf("%s.name LIKE ?", dao.TableName())
		db.Where(query, fmt.Sprintf("%%%s%%", cond.NameLike))
	}
	if cond.CompanyID > 0 {
		query := fmt.Sprintf("%s.company_id = ?", dao.TableName())
		db.Where(query, cond.CompanyID)
	}
	if cond.Uin > 0 {
		query := fmt.Sprintf("%s.uin = ?", dao.TableName())
		db.Where(query, cond.Uin)
	}
	if len(cond.ProjectTypeList) > 0 {
		query := fmt.Sprintf("%s.project_type IN ?", dao.TableName())
		db.Where(query, cond.ProjectTypeList)
	}
}

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

type KeCompanyQuotaCond struct {
	BaseCond
	Filters         []apiobj.Filter
	ID              uint
	CompanyID       uint
	SourceType      foresttype.CompanyQuotaSourceType
	ExpireAtStart   *time.Time
	ExpireAtEnd     *time.Time
	MinPackageLevel foresttype.PackageLevel
}

type KeCompanyQuotaDao struct {
	BaseModel
}

func NewKeCompanyQuotaDao() *KeCompanyQuotaDao {
	return &KeCompanyQuotaDao{}
}

func (dao *KeCompanyQuotaDao) TableName() string {
	return foresttype.TableNameKeCompanyQuota
}

func (dao *KeCompanyQuotaDao) WithTx(db *gorm.DB) *KeCompanyQuotaDao {
	return &KeCompanyQuotaDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *KeCompanyQuotaDao) Insert(ctx context.Context, entity *foresttype.KeCompanyQuota) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[KeCompanyQuotaDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *KeCompanyQuotaDao) BatchInsert(ctx context.Context, entityList foresttype.KeCompanyQuotaList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[KeCompanyQuotaDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return fmt.Errorf("[KeCompanyQuotaDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *KeCompanyQuotaDao) UpdateByID(ctx context.Context, id uint, entity *foresttype.KeCompanyQuota) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[KeCompanyQuotaDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *KeCompanyQuotaDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[KeCompanyQuotaDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *KeCompanyQuotaDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[KeCompanyQuotaDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *KeCompanyQuotaDao) GetByID(ctx context.Context, id uint) (*foresttype.KeCompanyQuota, error) {
	var entity foresttype.KeCompanyQuota
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[KeCompanyQuotaDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *KeCompanyQuotaDao) GetByCond(ctx context.Context, cond *KeCompanyQuotaCond) (*foresttype.KeCompanyQuota, error) {
	var entity foresttype.KeCompanyQuota
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[KeCompanyQuotaDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *KeCompanyQuotaDao) GetListByCond(ctx context.Context, cond *KeCompanyQuotaCond) (foresttype.KeCompanyQuotaList, error) {
	var entityList foresttype.KeCompanyQuotaList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[KeCompanyQuotaDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

func (dao *KeCompanyQuotaDao) GetPageListByCond(ctx context.Context, cond *KeCompanyQuotaCond) (foresttype.KeCompanyQuotaList, int64, error) {
	db := dao.DB(ctx).Model(&foresttype.KeCompanyQuota{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[KeCompanyQuotaDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList foresttype.KeCompanyQuotaList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[KeCompanyQuotaDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *KeCompanyQuotaDao) CountByCond(ctx context.Context, cond *KeCompanyQuotaCond) (int64, error) {
	db := dao.DB(ctx).Model(&foresttype.KeCompanyQuota{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[KeCompanyQuotaDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

// GetGroupCountByPackageLevel 按 PackageLevel 分组统计数量，返回 map[PackageLevel]count
func (dao *KeCompanyQuotaDao) GetGroupCountByPackageLevel(ctx context.Context, cond *KeCompanyQuotaCond) (map[foresttype.PackageLevel]int64, error) {
	db := dao.DB(ctx).Model(&foresttype.KeCompanyQuota{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	type packageLevelCount struct {
		PackageLevel foresttype.PackageLevel `gorm:"column:package_level"`
		Count        int64                   `gorm:"column:count"`
	}
	var results []packageLevelCount
	groupField := fmt.Sprintf("%s.package_level", dao.TableName())
	if err := db.Select(groupField + ", COUNT(*) as count").
		Group(groupField).
		Scan(&results).Error; err != nil {
		return nil, fmt.Errorf("[KeCompanyQuotaDao] GetGroupCountByPackageLevel fail, cond:%s, err: %v", logs.JSON(cond), err)
	}

	// 转换为 map
	resultMap := make(map[foresttype.PackageLevel]int64, len(results))
	for _, item := range results {
		resultMap[item.PackageLevel] = item.Count
	}
	return resultMap, nil
}

func (dao *KeCompanyQuotaDao) BuildCondition(db *gorm.DB, cond *KeCompanyQuotaCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}
	if cond.CompanyID > 0 {
		query := fmt.Sprintf("%s.company_id = ?", dao.TableName())
		db.Where(query, cond.CompanyID)
	}
	if cond.ExpireAtStart != nil {
		query := fmt.Sprintf("%s.expire_at >= ?", dao.TableName())
		db.Where(query, cond.ExpireAtStart)
	}
	if cond.ExpireAtEnd != nil {
		query := fmt.Sprintf("%s.expire_at <= ?", dao.TableName())
		db.Where(query, cond.ExpireAtEnd)
	}
	if cond.SourceType != "" {
		query := fmt.Sprintf("%s.source_type = ?", dao.TableName())
		db.Where(query, cond.SourceType)
	}
	if cond.MinPackageLevel > 0 {
		query := fmt.Sprintf("%s.package_level >= ?", dao.TableName())
		db.Where(query, cond.MinPackageLevel)
	}
}

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

type KeResourceScopeCond struct {
	BaseCond
	Filters       []apiobj.Filter
	CompanyID     uint
	ResourceType  foresttype.ResourceType
	ResourceID    uint
	ResourceIDs   []uint
	ScopeTypeList []foresttype.ScopeType
	ScopeID       uint
	Action        foresttype.ActionType
}

type KeResourceScopeDao struct {
	BaseModel
}

func NewKeResourceScopeDao() *KeResourceScopeDao {
	return &KeResourceScopeDao{}
}

func (dao *KeResourceScopeDao) TableName() string {
	return foresttype.TableNameKeResourceScope
}

func (dao *KeResourceScopeDao) WithTx(db *gorm.DB) *KeResourceScopeDao {
	return &KeResourceScopeDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *KeResourceScopeDao) Insert(ctx context.Context, entity *foresttype.KeResourceScope) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[KeResourceScopeDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *KeResourceScopeDao) BatchInsert(ctx context.Context, entityList foresttype.KeResourceScopeList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[KeResourceScopeDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return fmt.Errorf("[KeResourceScopeDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *KeResourceScopeDao) UpdateByID(ctx context.Context, id uint, entity *foresttype.KeResourceScope) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[KeResourceScopeDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *KeResourceScopeDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[KeResourceScopeDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *KeResourceScopeDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_time": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[KeResourceScopeDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *KeResourceScopeDao) DeleteByCond(ctx context.Context, cond *KeResourceScopeCond) error {
	db := dao.DB(ctx).Table(dao.TableName())
	dao.BuildCondition(db, cond)

	if err := db.Delete(&foresttype.KeResourceScope{}).Error; err != nil {
		return fmt.Errorf("[KeResourceScopeDao] DeleteByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return nil
}

func (dao *KeResourceScopeDao) GetByID(ctx context.Context, id uint) (*foresttype.KeResourceScope, error) {
	var entity foresttype.KeResourceScope
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[KeResourceScopeDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *KeResourceScopeDao) GetByCond(ctx context.Context, cond *KeResourceScopeCond) (*foresttype.KeResourceScope, error) {
	var entity foresttype.KeResourceScope
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[KeResourceScopeDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *KeResourceScopeDao) GetAuthResourceIDs(ctx context.Context, cond *KeResourceScopeCond) ([]uint, error) {
	db := dao.DB(ctx).Table(dao.TableName()).Model(foresttype.KeResourceScope{})
	db.Select("resource_id").Where("resource_type = ? AND deleted_at IS NULL", cond.ResourceType).
		Where("("+
			// 1. 公开权限 (action = 'view' 且 scope_type = 'public')
			"(action = ? AND scope_type = ?) OR "+
			// 2. 个人管理权限 (action = 'manage' 且 scope_type = 'user' 且 scope_id = 当前用户)
			"(action = ? AND scope_type = ? AND scope_id = ?) OR "+
			// 3. 个人查看权限 (action = 'view' 且 scope_type = 'user' 且 scope_id = 当前用户)
			"(action = ? AND scope_type = ? AND scope_id = ?) OR "+
			// 4. 公司权限 (action = 'view' 且 scope_type = 'company' 且 scope_id = 当前公司)
			"(action = ? AND scope_type = ? AND scope_id = ?)"+
			")",
			foresttype.ActionView, foresttype.ScopeTypePublic, // 公开
			foresttype.ActionManage, foresttype.ScopeTypeUser, cond.Uin, // 个人管理
			foresttype.ActionView, foresttype.ScopeTypeUser, cond.Uin, // 个人查看
			foresttype.ActionView, foresttype.ScopeTypeCompany, cond.CompanyID) // 公司权限
	var ids []uint
	if err := db.Find(&ids).Error; err != nil {
		return nil, fmt.Errorf("[KeResourceScopeDao] GetAuthResourceIDs fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return ids, nil
}

func (dao *KeResourceScopeDao) GetListByCond(ctx context.Context, cond *KeResourceScopeCond) (foresttype.KeResourceScopeList, error) {
	var entityList foresttype.KeResourceScopeList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[KeResourceScopeDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

func (dao *KeResourceScopeDao) GetPageListByCond(ctx context.Context, cond *KeResourceScopeCond) (foresttype.KeResourceScopeList, int64, error) {
	db := dao.DB(ctx).Model(&foresttype.KeResourceScope{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[KeResourceScopeDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList foresttype.KeResourceScopeList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[KeResourceScopeDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *KeResourceScopeDao) CountByCond(ctx context.Context, cond *KeResourceScopeCond) (int64, error) {
	db := dao.DB(ctx).Model(&foresttype.KeResourceScope{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[KeResourceScopeDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

func (dao *KeResourceScopeDao) BuildCondition(db *gorm.DB, cond *KeResourceScopeCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}
	if cond.ResourceType != "" {
		query := fmt.Sprintf("%s.resource_type = ?", dao.TableName())
		db.Where(query, cond.ResourceType)
	}
	if cond.ResourceID > 0 {
		query := fmt.Sprintf("%s.resource_id = ?", dao.TableName())
		db.Where(query, cond.ResourceID)
	}
	if len(cond.ResourceIDs) > 0 {
		query := fmt.Sprintf("%s.resource_id IN (?)", dao.TableName())
		db.Where(query, cond.ResourceIDs)
	}
	if len(cond.ScopeTypeList) > 0 {
		query := fmt.Sprintf("%s.scope_type IN (?)", dao.TableName())
		db.Where(query, cond.ScopeTypeList)
	}
	if cond.ScopeID > 0 {
		query := fmt.Sprintf("%s.scope_id = ?", dao.TableName())
		db.Where(query, cond.ScopeID)
	}
	if cond.Action != "" {
		query := fmt.Sprintf("%s.action = ?", dao.TableName())
		db.Where(query, cond.Action)
	}
}

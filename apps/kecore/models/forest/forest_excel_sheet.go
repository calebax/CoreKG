package forest

import (
	"context"
	"fmt"
	"time"

	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/types"
	"gorm.io/gorm"
)

type ForestExcelSheetCond struct {
	BaseCond
	Filters       []apiobj.Filter
	ID            uint
	IDS           []uint
	ForestIDs     []uint
	ForestFileID  uint
	ForestFileIDs []uint
	SheetNameLike string
	SheetType     foresttype.ExcelSheetType
	ParentIDs     []uint
	Enable        types.Bool
}

type ForestExcelSheetDao struct {
	BaseModel
}

func NewForestExcelSheetDao() *ForestExcelSheetDao {
	return &ForestExcelSheetDao{}
}

func (dao *ForestExcelSheetDao) TableName() string {
	return foresttype.TableNameKeForestExcelSheet
}

func (dao *ForestExcelSheetDao) WithTx(db *gorm.DB) *ForestExcelSheetDao {
	return &ForestExcelSheetDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *ForestExcelSheetDao) Insert(ctx context.Context, entity *foresttype.ForestExcelSheet) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[KeForestExcelSheetDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *ForestExcelSheetDao) BatchInsert(ctx context.Context, entityList foresttype.ForestExcelSheetList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[KeForestExcelSheetDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return fmt.Errorf("[KeForestExcelSheetDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *ForestExcelSheetDao) UpdateByID(ctx context.Context, id uint, entity *foresttype.ForestExcelSheet) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[KeForestExcelSheetDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *ForestExcelSheetDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("forest_file_id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[KeForestExcelSheetDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *ForestExcelSheetDao) UpdateExcelIDsMap(ctx context.Context, id []uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("forest_file_id IN (?)", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[KeForestExcelSheetDao] UpdateExcelIDsMap fail, id:%s, updateMap:%s, err: %v", logs.JSON(id), logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *ForestExcelSheetDao) UpdateIDsMap(ctx context.Context, id []uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id IN (?)", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[KeForestExcelSheetDao] UpdateIDsMap fail, id:%s, updateMap:%s, err: %v", logs.JSON(id), logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *ForestExcelSheetDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_time": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[KeForestExcelSheetDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *ForestExcelSheetDao) GetByID(ctx context.Context, id uint) (*foresttype.ForestExcelSheet, error) {
	var entity foresttype.ForestExcelSheet
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[KeForestExcelSheetDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *ForestExcelSheetDao) GetByCond(ctx context.Context, cond *ForestExcelSheetCond) (*foresttype.ForestExcelSheet, error) {
	var entity foresttype.ForestExcelSheet
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[KeForestExcelSheetDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *ForestExcelSheetDao) GetListByCond(ctx context.Context, cond *ForestExcelSheetCond) (foresttype.ForestExcelSheetList, error) {
	var entityList foresttype.ForestExcelSheetList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[KeForestExcelSheetDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

func (dao *ForestExcelSheetDao) GetPageListByCond(ctx context.Context, cond *ForestExcelSheetCond) (foresttype.ForestExcelSheetList, int64, error) {
	db := dao.DB(ctx).Model(&foresttype.ForestExcelSheet{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[KeForestExcelSheetDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList foresttype.ForestExcelSheetList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[KeForestExcelSheetDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *ForestExcelSheetDao) CountByCond(ctx context.Context, cond *ForestExcelSheetCond) (int64, error) {
	db := dao.DB(ctx).Model(&foresttype.ForestExcelSheet{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[KeForestExcelSheetDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

func (dao *ForestExcelSheetDao) BuildCondition(db *gorm.DB, cond *ForestExcelSheetCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}
	if len(cond.IDS) > 0 {
		query := fmt.Sprintf("%s.id IN ?", dao.TableName())
		db.Where(query, cond.IDS)
	}
	if len(cond.ForestIDs) > 0 {
		query := fmt.Sprintf("%s.forest_id IN (?)", dao.TableName())
		db.Where(query, cond.ForestIDs)
	}
	if cond.ForestFileID > 0 {
		query := fmt.Sprintf("%s.forest_file_id = ?", dao.TableName())
		db.Where(query, cond.ForestFileID)
	}
	if len(cond.ForestFileIDs) > 0 {
		query := fmt.Sprintf("%s.forest_file_id IN (?)", dao.TableName())
		db.Where(query, cond.ForestFileIDs)
	}
	if cond.SheetNameLike != "" {
		query := fmt.Sprintf("%s.sheet_name LIKE ?", dao.TableName())
		db.Where(query, fmt.Sprintf("%%%s%%", cond.SheetNameLike))
	}
	if cond.SheetType != "" {
		query := fmt.Sprintf("%s.sheet_type = ?", dao.TableName())
		db.Where(query, cond.SheetType)
	}
	if len(cond.ParentIDs) > 0 {
		query := fmt.Sprintf("%s.parent_id IN (?)", dao.TableName())
		db.Where(query, cond.ParentIDs)
	}
	if cond.Enable != 0 {
		db = db.Where(fmt.Sprintf("%s.enable = ?", dao.TableName()), cond.Enable)
	}
}

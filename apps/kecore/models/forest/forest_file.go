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

type ForestFileCond struct {
	BaseCond
	Filters   []apiobj.Filter
	ID        uint
	IDs       []uint
	ForestIDs []uint
	// Status 文件上传状态
	Status foresttype.FileStatus
	// ParseStatus 解析未 markdown 的状态
	ParseStatus foresttype.KnownowForestTaskStatus
	// DescStatus 摘要生成状态
	DescStatus foresttype.KnownowForestTaskStatus
	// KnowledgeStatus 拆 chunk 生成状态
	KnowledgeStatus foresttype.KnownowForestTaskStatus
	IsDir           types.Bool
	NameLike        string
	Enable          types.Bool
	CompanyID       uint
	NotInIDs        []uint
}

type ForestFileDao struct {
	BaseModel
}

func NewForestFileDao() *ForestFileDao {
	return &ForestFileDao{}
}

func (dao *ForestFileDao) TableName() string {
	return foresttype.TableNameKnownowForestFile
}

func (dao *ForestFileDao) WithTx(db *gorm.DB) *ForestFileDao {
	return &ForestFileDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *ForestFileDao) Insert(ctx context.Context, entity *foresttype.KnownowForestFile) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[KeForestFileDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *ForestFileDao) BatchInsert(ctx context.Context, entityList foresttype.KeForestFileList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[KeForestFileDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return fmt.Errorf("[KeForestFileDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *ForestFileDao) UpdateByID(ctx context.Context, id uint, entity *foresttype.KnownowForestFile) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[KeForestFileDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *ForestFileDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[KeForestFileDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *ForestFileDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_time": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[KeForestFileDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *ForestFileDao) GetByID(ctx context.Context, id uint) (*foresttype.KnownowForestFile, error) {
	var entity foresttype.KnownowForestFile
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[KeForestFileDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *ForestFileDao) GetByCond(ctx context.Context, cond *ForestFileCond) (*foresttype.KnownowForestFile, error) {
	var entity foresttype.KnownowForestFile
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[KeForestFileDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *ForestFileDao) GetListByCond(ctx context.Context, cond *ForestFileCond) (foresttype.KeForestFileList, error) {
	var entityList foresttype.KeForestFileList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[KeForestFileDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

func (dao *ForestFileDao) GetPageListByCond(ctx context.Context, cond *ForestFileCond) (foresttype.KeForestFileList, int64, error) {
	db := dao.DB(ctx).Model(&foresttype.KnownowForestFile{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[KeForestFileDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList foresttype.KeForestFileList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[KeForestFileDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *ForestFileDao) CountByCond(ctx context.Context, cond *ForestFileCond) (int64, error) {
	db := dao.DB(ctx).Model(&foresttype.KnownowForestFile{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[KeForestFileDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

func (dao *ForestFileDao) StatSizeByCond(ctx context.Context, cond *ForestFileCond) (int64, error) {
	db := dao.DB(ctx).Model(&foresttype.KnownowForestFile{}).Table(dao.TableName())
	dao.BuildCondition(db, cond)
	var size int64
	if err := db.Select("COALESCE(SUM(size), 0) AS total_size").Scan(&size).Error; err != nil {
		return 0, fmt.Errorf("[KeForestFileDao] StatSizeByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return size, nil
}

func (dao *ForestFileDao) UpdateIDsMap(ctx context.Context, id []uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id IN (?)", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[KeForestFileDao] UpdateIDsMap fail, id:%s, updateMap:%s, err: %v", logs.JSON(id), logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *ForestFileDao) BuildCondition(db *gorm.DB, cond *ForestFileCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}
	if len(cond.IDs) > 0 {
		query := fmt.Sprintf("%s.id IN (?)", dao.TableName())
		db.Where(query, cond.IDs)
	}
	if len(cond.ForestIDs) > 0 {
		query := fmt.Sprintf("%s.forest_id IN (?)", dao.TableName())
		db.Where(query, cond.ForestIDs)
	}
	if cond.ParseStatus != "" {
		query := fmt.Sprintf("%s.parse_status = ?", dao.TableName())
		db.Where(query, cond.ParseStatus)
	}
	if cond.NameLike != "" {
		query := fmt.Sprintf("%s.name LIKE ?", dao.TableName())
		db.Where(query, fmt.Sprintf("%%%s%%", cond.NameLike))
	}
	if cond.IsDir != 0 {
		query := fmt.Sprintf("%s.is_dir = ?", dao.TableName())
		db.Where(query, cond.IsDir)
	}
	if cond.Enable != 0 {
		query := fmt.Sprintf("%s.enable = ?", dao.TableName())
		db.Where(query, cond.Enable)
	}
	if cond.CompanyID > 0 {
		query := fmt.Sprintf("%s.company_id = ?", dao.TableName())
		db.Where(query, cond.CompanyID)
	}
	if cond.Status != "" {
		query := fmt.Sprintf("%s.status = ?", dao.TableName())
		db.Where(query, cond.Status)
	}
	if cond.ParseStatus != "" {
		query := fmt.Sprintf("%s.parse_status = ?", dao.TableName())
		db.Where(query, cond.ParseStatus)
	}
	if cond.KnowledgeStatus != "" {
		query := fmt.Sprintf("%s.knowledge_status = ?", dao.TableName())
		db.Where(query, cond.KnowledgeStatus)
	}
	if cond.DescStatus != "" {
		query := fmt.Sprintf("%s.desc_status = ?", dao.TableName())
		db.Where(query, cond.DescStatus)
	}
	if len(cond.NotInIDs) > 0 {
		query := fmt.Sprintf("%s.id NOT IN (?)", dao.TableName())
		db.Where(query, cond.NotInIDs)
	}
}

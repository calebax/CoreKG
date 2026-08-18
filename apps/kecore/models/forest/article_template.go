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

// Deprecated: 已合并到 ArticleDao，通过 ArticleCond.Type/SourceType/SourceID 条件替代
type ArticleTemplateCond struct {
	BaseCond
	Filters      []apiobj.Filter
	ID           uint
	TemplateType foresttype.TemplateType
	SourceType   foresttype.SourceType
	SourceID     uint
}

// Deprecated: 已合并到 ArticleDao，通过 type 字段区分模板和文章
type ArticleTemplateDao struct {
	BaseModel
}

func NewArticleTemplateDao() *ArticleTemplateDao {
	return &ArticleTemplateDao{}
}

func (dao *ArticleTemplateDao) TableName() string {
	return foresttype.TableNameKeArticleTemplate
}

func (dao *ArticleTemplateDao) WithTx(db *gorm.DB) *ArticleTemplateDao {
	return &ArticleTemplateDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *ArticleTemplateDao) Insert(ctx context.Context, entity *foresttype.KeArticleTemplate) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[ArticleTemplateDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *ArticleTemplateDao) BatchInsert(ctx context.Context, entityList foresttype.KeArticleTemplateList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[ArticleTemplateDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return fmt.Errorf("[ArticleTemplateDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *ArticleTemplateDao) UpdateByID(ctx context.Context, id uint, entity *foresttype.KeArticleTemplate) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[ArticleTemplateDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *ArticleTemplateDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[ArticleTemplateDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *ArticleTemplateDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[ArticleTemplateDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *ArticleTemplateDao) GetByID(ctx context.Context, id uint) (*foresttype.KeArticleTemplate, error) {
	var entity foresttype.KeArticleTemplate
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[ArticleTemplateDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *ArticleTemplateDao) GetByCond(ctx context.Context, cond *ArticleTemplateCond) (*foresttype.KeArticleTemplate, error) {
	var entity foresttype.KeArticleTemplate
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[ArticleTemplateDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *ArticleTemplateDao) GetListByCond(ctx context.Context, cond *ArticleTemplateCond) (foresttype.KeArticleTemplateList, error) {
	var entityList foresttype.KeArticleTemplateList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[ArticleTemplateDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

func (dao *ArticleTemplateDao) GetPageListByCond(ctx context.Context, cond *ArticleTemplateCond) (foresttype.KeArticleTemplateList, int64, error) {
	db := dao.DB(ctx).Model(&foresttype.KeArticleTemplate{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[ArticleTemplateDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList foresttype.KeArticleTemplateList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[ArticleTemplateDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *ArticleTemplateDao) CountByCond(ctx context.Context, cond *ArticleTemplateCond) (int64, error) {
	db := dao.DB(ctx).Model(&foresttype.KeArticleTemplate{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[ArticleTemplateDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

func (dao *ArticleTemplateDao) BuildCondition(db *gorm.DB, cond *ArticleTemplateCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}
	if cond.TemplateType != "" {
		query := fmt.Sprintf("%s.template_type = ?", dao.TableName())
		db.Where(query, cond.TemplateType)
	}
	if cond.SourceType != "" {
		query := fmt.Sprintf("%s.source_type = ?", dao.TableName())
		db.Where(query, cond.SourceType)
	}
	if cond.SourceID > 0 {
		query := fmt.Sprintf("%s.source_id = ?", dao.TableName())
		db.Where(query, cond.SourceID)
	}
}

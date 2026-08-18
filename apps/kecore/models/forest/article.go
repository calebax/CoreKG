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

type ArticleCond struct {
	BaseCond
	Filters      []apiobj.Filter
	ID           uint
	ArticleTypes []foresttype.ArticleType
	SourceType   foresttype.SourceType
	SourceID     uint
}

type ArticleDao struct {
	BaseModel
}

func NewArticleDao() *ArticleDao {
	return &ArticleDao{}
}

func (dao *ArticleDao) TableName() string {
	return foresttype.TableNameKeArticle
}

func (dao *ArticleDao) WithTx(db *gorm.DB) *ArticleDao {
	return &ArticleDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *ArticleDao) Insert(ctx context.Context, entity *foresttype.KeArticle) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[ArticleDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *ArticleDao) BatchInsert(ctx context.Context, entityList foresttype.KeArticleList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[ArticleDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return fmt.Errorf("[ArticleDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *ArticleDao) UpdateByID(ctx context.Context, id uint, entity *foresttype.KeArticle) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[ArticleDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *ArticleDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[ArticleDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *ArticleDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[ArticleDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *ArticleDao) GetByID(ctx context.Context, id uint) (*foresttype.KeArticle, error) {
	var entity foresttype.KeArticle
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[ArticleDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *ArticleDao) GetByCond(ctx context.Context, cond *ArticleCond) (*foresttype.KeArticle, error) {
	var entity foresttype.KeArticle
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[ArticleDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *ArticleDao) GetListByCond(ctx context.Context, cond *ArticleCond) (foresttype.KeArticleList, error) {
	var entityList foresttype.KeArticleList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[ArticleDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

func (dao *ArticleDao) GetPageListByCond(ctx context.Context, cond *ArticleCond) (foresttype.KeArticleList, int64, error) {
	db := dao.DB(ctx).Model(&foresttype.KeArticle{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[ArticleDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList foresttype.KeArticleList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[ArticleDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *ArticleDao) CountByCond(ctx context.Context, cond *ArticleCond) (int64, error) {
	db := dao.DB(ctx).Model(&foresttype.KeArticle{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[ArticleDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

func (dao *ArticleDao) BuildCondition(db *gorm.DB, cond *ArticleCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}
	if len(cond.ArticleTypes) > 0 {
		query := fmt.Sprintf("%s.type IN ?", dao.TableName())
		db.Where(query, cond.ArticleTypes)
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

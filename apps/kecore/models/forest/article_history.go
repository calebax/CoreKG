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

type ArticleHistoryCond struct {
	BaseCond
	Filters    []apiobj.Filter
	ID         uint
	ArticleID  uint
}

type ArticleHistoryDao struct {
	BaseModel
}

func NewArticleHistoryDao() *ArticleHistoryDao {
	return &ArticleHistoryDao{}
}

func (dao *ArticleHistoryDao) TableName() string {
	return foresttype.TableNameKeArticleHistory
}

func (dao *ArticleHistoryDao) WithTx(db *gorm.DB) *ArticleHistoryDao {
	return &ArticleHistoryDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *ArticleHistoryDao) Insert(ctx context.Context, entity *foresttype.KeArticleHistory) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[ArticleHistoryDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *ArticleHistoryDao) BatchInsert(ctx context.Context, entityList foresttype.KeArticleHistoryList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[ArticleHistoryDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return fmt.Errorf("[ArticleHistoryDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *ArticleHistoryDao) UpdateByID(ctx context.Context, id uint, entity *foresttype.KeArticleHistory) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[ArticleHistoryDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *ArticleHistoryDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[ArticleHistoryDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *ArticleHistoryDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[ArticleHistoryDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *ArticleHistoryDao) GetByID(ctx context.Context, id uint) (*foresttype.KeArticleHistory, error) {
	var entity foresttype.KeArticleHistory
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[ArticleHistoryDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *ArticleHistoryDao) GetByCond(ctx context.Context, cond *ArticleHistoryCond) (*foresttype.KeArticleHistory, error) {
	var entity foresttype.KeArticleHistory
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[ArticleHistoryDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *ArticleHistoryDao) GetListByCond(ctx context.Context, cond *ArticleHistoryCond) (foresttype.KeArticleHistoryList, error) {
	var entityList foresttype.KeArticleHistoryList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[ArticleHistoryDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

func (dao *ArticleHistoryDao) GetPageListByCond(ctx context.Context, cond *ArticleHistoryCond) (foresttype.KeArticleHistoryList, int64, error) {
	db := dao.DB(ctx).Model(&foresttype.KeArticleHistory{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[ArticleHistoryDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList foresttype.KeArticleHistoryList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[ArticleHistoryDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *ArticleHistoryDao) CountByCond(ctx context.Context, cond *ArticleHistoryCond) (int64, error) {
	db := dao.DB(ctx).Model(&foresttype.KeArticleHistory{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[ArticleHistoryDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

func (dao *ArticleHistoryDao) BuildCondition(db *gorm.DB, cond *ArticleHistoryCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}
	if cond.ArticleID > 0 {
		query := fmt.Sprintf("%s.article_id = ?", dao.TableName())
		db.Where(query, cond.ArticleID)
	}
}

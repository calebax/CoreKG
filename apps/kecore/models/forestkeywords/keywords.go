package forestkeywords

import (
	"context"
	"fmt"
	"time"

	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/user"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

type KeywordsCond struct {
	BaseCond
	Filters       []apiobj.Filter
	ID            uint
	SubjectID     int
	WordType      foresttype.WordType
	Words         []string
	LikeWord      string
	SubjectIDNot0 bool
}

type KeywordsDao struct {
	BaseModel
}

func NewKeywordsDao() *KeywordsDao {
	return &KeywordsDao{}
}

func (dao *KeywordsDao) TableName() string {
	return foresttype.TableNameKeywords
}

func (dao *KeywordsDao) WithTx(db *gorm.DB) *KeywordsDao {
	return &KeywordsDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *KeywordsDao) Insert(ctx context.Context, entity *foresttype.Keywords) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[KeywordsDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *KeywordsDao) BatchInsert(ctx context.Context, entityList foresttype.KeywordsList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[KeywordsDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return fmt.Errorf("[KeywordsDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *KeywordsDao) UpdateByID(ctx context.Context, id uint, entity *foresttype.Keywords) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[KeywordsDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *KeywordsDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[KeywordsDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *KeywordsDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[KeywordsDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *KeywordsDao) DeleteByIDs(ctx context.Context, ids []uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id IN (?)", ids).Delete(&foresttype.Keywords{}).Error; err != nil {
		return fmt.Errorf("[KeywordsDao] Delete fail, id:%d, err: %v", ids, err)
	}
	return nil
}

func (dao *KeywordsDao) DeleteBySubjectID(ctx context.Context, subjectID uint, softDelete bool) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if !softDelete {
		db = db.Unscoped()
	}
	if err := db.Where("subject_id = ?", subjectID).Delete(&foresttype.Keywords{}).Error; err != nil {
		return fmt.Errorf("[KeywordsDao] Delete fail, subjectID:%d, err: %v", subjectID, err)
	}
	return nil
}

func (dao *KeywordsDao) GetByID(ctx context.Context, id uint) (*foresttype.Keywords, error) {
	var entity foresttype.Keywords
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[KeywordsDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *KeywordsDao) GetByCond(ctx context.Context, cond *KeywordsCond) (*foresttype.Keywords, error) {
	var entity foresttype.Keywords
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[KeywordsDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *KeywordsDao) GetListByCond(ctx context.Context, cond *KeywordsCond) (foresttype.KeywordsList, error) {
	var entityList foresttype.KeywordsList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[KeywordsDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

func (dao *KeywordsDao) GetPageListByCond(ctx context.Context, cond *KeywordsCond) (foresttype.KeywordsList, int64, error) {
	db := dao.DB(ctx).Model(&foresttype.Keywords{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[KeywordsDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList foresttype.KeywordsList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[KeywordsDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *KeywordsDao) CountByCond(ctx context.Context, cond *KeywordsCond) (int64, error) {
	db := dao.DB(ctx).Model(&foresttype.Keywords{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[KeywordsDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

func (dao *KeywordsDao) BuildCondition(db *gorm.DB, cond *KeywordsCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.SubjectID != -1 {
		query := fmt.Sprintf("%s.subject_id = ?", dao.TableName())
		db.Where(query, cond.SubjectID)
	}
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}
	if cond.WordType != "" {
		query := fmt.Sprintf("%s.word_type = ?", dao.TableName())
		db.Where(query, cond.WordType)
	}
	if len(cond.Words) > 0 {
		query := fmt.Sprintf("%s.word IN (?)", dao.TableName())
		db.Where(query, cond.Words)
	}
	if cond.LikeWord != "" {
		query := fmt.Sprintf("%s.word LIKE ?", dao.TableName())
		db.Where(query, fmt.Sprintf("%%%s%%", cond.LikeWord))
	}
	if cond.SubjectIDNot0 {
		query := fmt.Sprintf("%s.subject_id != 0", dao.TableName())
		db.Where(query)
	}
}

// ListSynonymKeywords 获取同义词列表
func (dao *KeywordsDao) ListSynonymKeywords(ctx context.Context, cond *KeywordsCond) ([]*SynonymKeywordDetail, int64, error) {
	db := dao.DB(ctx).Model(&foresttype.Keywords{}).Table(dao.TableName())
	res := []*SynonymKeywordDetail{}
	cond.WordType = foresttype.WordTypeSynonym
	list, count, err := dao.GetPageListByCond(ctx, cond)
	if err != nil {
		logs.ErrorContextf(ctx, "ListSynonymKeywords GetPageListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
		return nil, 0, err
	}

	if count == 0 || len(list) == 0 {
		return res, 0, nil
	}
	// 获取所有id
	subjectIDs := make([]uint, 0, len(list))
	subjectIDSet := make(map[uint]struct{})

	for _, k := range list {
		if _, ok := subjectIDSet[k.ID]; !ok {
			subjectIDSet[k.ID] = struct{}{}
			subjectIDs = append(subjectIDs, k.ID)
		}
	}
	// 获取所有subject_id为列表中的子元素
	var childrenAll []foresttype.Keywords
	err = db.
		Where("subject_id IN ?", subjectIDs).
		Find(&childrenAll).Error
	if err != nil {
		logs.ErrorContextf(ctx,
			"ListSynonymKeywords query children fail, subjectIDs:%v, err:%v",
			subjectIDs, err,
		)
		return nil, 0, err
	}
	// subject_id 分组
	group := make(map[uint][]foresttype.Keywords, len(subjectIDs))
	for _, k := range childrenAll {
		group[k.SubjectID] = append(group[k.SubjectID], k)
	}
	userMap := map[uint]*accounttype.User{}
	// map组装
	for _, parent := range list {
		userEntity, exists := userMap[parent.Uin]
		if !exists {
			userEntity, err = user.GetUserByUin(ctx, parent.Uin)
			if err != nil {
				logs.ErrorContextf(ctx, "GetUserByUin error: %v", err)
				continue
			}
			userMap[parent.Uin] = userEntity
		}
		res = append(res, &SynonymKeywordDetail{
			Keywords:        parent,
			UserName:        userEntity.Name,
			SynonymKeywords: group[parent.ID],
		})
	}

	return res, count, nil
}

// GetSynonymKeywords 获取详情
func (dao *KeywordsDao) GetSynonymKeywords(ctx context.Context, id uint) (*SynonymKeywordDetail, error) {
	parent, err := dao.GetByID(ctx, id)
	if err != nil {
		logs.ErrorContextf(ctx, "GetSynonymKeywords GetByID err: %v", err)
		return nil, err
	}
	if parent == nil {
		return nil, gorm.ErrRecordNotFound
	}
	children, err := dao.GetListByCond(ctx, &KeywordsCond{
		SubjectID: int(parent.ID),
		WordType:  foresttype.WordTypeSynonym,
	})
	res := &SynonymKeywordDetail{
		Keywords:        *parent,
		SynonymKeywords: children,
	}
	return res, nil
}

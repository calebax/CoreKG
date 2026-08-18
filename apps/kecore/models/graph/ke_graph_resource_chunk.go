package graph

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/concqueue"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

type KeGraphResourceChunkCond struct {
	BaseCond
	Filters []apiobj.Filter
	ID      uint
}

type KeGraphResourceChunkDao struct {
	BaseModel
}

func NewKeGraphResourceChunkDao() *KeGraphResourceChunkDao {
	return &KeGraphResourceChunkDao{}
}

func (dao *KeGraphResourceChunkDao) TableName() string {
	return foresttype.TableNameKeGraphResourceChunk
}

func (dao *KeGraphResourceChunkDao) WithTx(db *gorm.DB) *KeGraphResourceChunkDao {
	return &KeGraphResourceChunkDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *KeGraphResourceChunkDao) Insert(ctx context.Context, entity *foresttype.KeGraphResourceChunk) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[KeGraphResourceChunkDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *KeGraphResourceChunkDao) BatchInsert(ctx context.Context, entityList foresttype.KeGraphResourceChunkList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[KeGraphResourceChunkDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return fmt.Errorf("[KeGraphResourceChunkDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *KeGraphResourceChunkDao) UpdateByID(ctx context.Context, id uint, entity *foresttype.KeGraphResourceChunk) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[KeGraphResourceChunkDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *KeGraphResourceChunkDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[KeGraphResourceChunkDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *KeGraphResourceChunkDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[KeGraphResourceChunkDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *KeGraphResourceChunkDao) GetByID(ctx context.Context, id uint) (*foresttype.KeGraphResourceChunk, error) {
	var entity foresttype.KeGraphResourceChunk
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[KeGraphResourceChunkDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *KeGraphResourceChunkDao) GetByCond(ctx context.Context, cond *KeGraphResourceChunkCond) (*foresttype.KeGraphResourceChunk, error) {
	var entity foresttype.KeGraphResourceChunk
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[KeGraphResourceChunkDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *KeGraphResourceChunkDao) GetListByCond(ctx context.Context, cond *KeGraphResourceChunkCond) (foresttype.KeGraphResourceChunkList, error) {
	var entityList foresttype.KeGraphResourceChunkList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[KeGraphResourceChunkDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

func (dao *KeGraphResourceChunkDao) GetPageListByCond(ctx context.Context, cond *KeGraphResourceChunkCond) (foresttype.KeGraphResourceChunkList, int64, error) {
	db := dao.DB(ctx).Model(&foresttype.KeGraphResourceChunk{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[KeGraphResourceChunkDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList foresttype.KeGraphResourceChunkList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[KeGraphResourceChunkDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *KeGraphResourceChunkDao) CountByCond(ctx context.Context, cond *KeGraphResourceChunkCond) (int64, error) {
	db := dao.DB(ctx).Model(&foresttype.KeGraphResourceChunk{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[KeGraphResourceChunkDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

func (dao *KeGraphResourceChunkDao) BuildCondition(db *gorm.DB, cond *KeGraphResourceChunkCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}
	for _, filter := range cond.Filters {
		switch filter.Field {
		case "resource_type":
			db.Where(fmt.Sprintf("%s.resource_type = ?", dao.TableName()), filter.Value[0])
		case "resource_id":
			db.Where(fmt.Sprintf("%s.resource_id = ?", dao.TableName()), filter.Value[0])
		}
	}
}

// GetListByChunkIDs 并发分组查询大量 chunkIDs 对应的实体
// 内部按批次并发查询以避免 IN 过长和单次查询过慢的问题
func (dao *KeGraphResourceChunkDao) GetListByChunkIDs(ctx context.Context, graphID, graphVersion uint, chunkIDs []string) (foresttype.KeGraphResourceChunkList, error) {
	if len(chunkIDs) == 0 {
		return nil, nil
	}

	// 默认批大小与最大并发
	const (
		batchSize      = 50
		maxConcurrency = 4
	)

	// 分批
	batches := make([][]string, 0, (len(chunkIDs)+batchSize-1)/batchSize)
	for i := 0; i < len(chunkIDs); i += batchSize {
		end := i + batchSize
		if end > len(chunkIDs) {
			end = len(chunkIDs)
		}
		batches = append(batches, chunkIDs[i:end])
	}

	var (
		result   foresttype.KeGraphResourceChunkList
		mu       sync.Mutex
		firstErr error
		errOnce  sync.Once
	)

	q := concqueue.New(maxConcurrency, len(batches), concqueue.WithContext(ctx), concqueue.WithErrorHandler(func(err error) {
		errOnce.Do(func() { firstErr = err })
	}))
	for _, batch := range batches {
		batchCopy := batch
		q.Submit(func(_ context.Context) error {
			var list foresttype.KeGraphResourceChunkList
			db := dbutil.Knownow().WithContext(ctx).
				Table(dao.TableName()).
				Session(&gorm.Session{})

			if err := db.
				Where("graph_id = ?", graphID).
				Where("graph_version_id = ?", graphVersion).
				Where("chunk_id IN ?", batchCopy).
				Find(&list).Error; err != nil {
				return fmt.Errorf("[KeGraphResourceChunkDao] GetListByChunkIDs batch query fail, err: %v", err)
			}

			if len(list) > 0 {
				mu.Lock()
				result = append(result, list...)
				mu.Unlock()
			}
			return nil
		})

	}

	_ = q.StopAndWait()
	if firstErr != nil {
		return nil, firstErr
	}
	return result, nil
}

// GetListByChunkIDs 并发分组查询大量 chunkIDs 对应的实体
// 内部按批次并发查询以避免 IN 过长和单次查询过慢的问题
func (dao *KeGraphResourceChunkDao) GetListByChunkID(ctx context.Context, graphID, graphVersion uint, chunkID string) (foresttype.KeGraphResourceChunkList, error) {

	var (
		result foresttype.KeGraphResourceChunkList
	)

	db := dao.DB(ctx).Table(dao.TableName())

	if err := db.Where("graph_id = ?", graphID).
		Where("graph_version_id = ?", graphVersion).Where("chunk_id = ?", chunkID).Find(&result).Error; err != nil {
		return nil, fmt.Errorf("[KeGraphResourceChunkDao] GetListByChunkIDs batch query fail, err: %v", err)
	}
	return result, nil
}

// BatchReplace 先按 chunkIDs 与 resourceType 软删除历史数据，再批量写入新的实体列表
// 要求 entityList 必须为同一 resource_type，并与传入 resourceType 一致
func (dao *KeGraphResourceChunkDao) BatchReplace(ctx context.Context, entityList foresttype.KeGraphResourceChunkList) error {
	if len(entityList) == 0 {
		// return fmt.Errorf("[KeGraphResourceChunkDao] BatchReplace fail, entityList is empty")
		return nil
	}
	var chunkIDs []string
	// 校验同一 resource_type 且与入参一致
	resourceType := entityList[0].ResourceType
	for _, e := range entityList {
		if e.ResourceType == "" || e.ResourceType != resourceType {
			return fmt.Errorf("[KeGraphResourceChunkDao] BatchReplace fail, inconsistent resource_type in entityList")
		}
		chunkIDs = append(chunkIDs, e.ChunkID)
	}
	txDB := dao.DB(ctx).Table(dao.TableName())

	// 分批软删除，避免 IN 过长
	const insertBatchSize = 500
	for i := 0; i < len(chunkIDs); i += insertBatchSize {
		end := i + insertBatchSize
		if end > len(chunkIDs) {
			end = len(chunkIDs)
		}
		batch := chunkIDs[i:end]

		// updatedField := map[string]interface{}{
		// 	"deleted_at": time.Now(),
		// }
		if err := txDB.Session(&gorm.Session{}).Unscoped().
			Where("graph_id = ?", entityList[0].GraphID).
			Where("resource_id = ?", entityList[0].ResourceID).
			Where("graph_version_id = ?", entityList[0].GraphVersionID).
			Where("resource_type = ? AND chunk_id IN ?", resourceType, batch).
			Delete(&foresttype.KeGraphResourceChunk{}).Error; err != nil {
			return fmt.Errorf("[KeGraphResourceChunkDao] BatchReplace soft delete fail, err: %v", err)
		}
	}

	// 将列表按 chunkIDs 分组，然后分组批量插入
	chunkIDToEntities := make(map[string]foresttype.KeGraphResourceChunkList)
	for _, e := range entityList {
		chunkIDToEntities[e.ChunkID] = append(chunkIDToEntities[e.ChunkID], e)
	}

	const batchSize = 500
	for i := 0; i < len(chunkIDs); i += batchSize {
		end := i + batchSize
		if end > len(chunkIDs) {
			end = len(chunkIDs)
		}
		batch := chunkIDs[i:end]

		var toInsert foresttype.KeGraphResourceChunkList
		for _, cid := range batch {
			if lst, ok := chunkIDToEntities[cid]; ok && len(lst) > 0 {
				toInsert = append(toInsert, lst...)
				// for _, v := range lst {
				// 	v.ID = 0
				// 	toInsert = append(toInsert, v)
				// }
			}
		}
		if len(toInsert) == 0 {
			continue
		}
		if err := txDB.Session(&gorm.Session{}).Create(toInsert).Error; err != nil {
			return fmt.Errorf("[KeGraphResourceChunkDao] BatchReplace insert fail, err: %v", err)
		}
	}

	return nil
}

func (dao *KeGraphResourceChunkDao) BatchDelete(ctx context.Context, ids []uint) error {
	return dao.DB(ctx).Table(dao.TableName()).
		Where("id IN ?", ids).
		Delete(&foresttype.KeGraphResourceChunk{}).Error
}

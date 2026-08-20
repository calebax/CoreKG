package admin

import (
	"context"
	"fmt"
	"time"

	"github.com/insmtx/corekg/pkgs/platform/admintype"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

type AnnouncementCond struct {
	BaseCond
	Filters []apiobj.Filter
	ID      uint
}

type AnnouncementDao struct {
	BaseModel
}

func NewAdminAnnouncementDao() *AnnouncementDao {
	return &AnnouncementDao{}
}

func (dao *AnnouncementDao) TableName() string {
	return admintype.TableNameAdminAnnouncement
}

func (dao *AnnouncementDao) WithTx(db *gorm.DB) *AnnouncementDao {
	return &AnnouncementDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *AnnouncementDao) Insert(ctx context.Context, entity *admintype.AdminAnnouncement) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entity).Error; err != nil {
		return fmt.Errorf("[AdminAnnouncementDao] Insert fail, entity:%s, err: %v", logs.JSON(entity), err)
	}
	return nil
}

func (dao *AnnouncementDao) BatchInsert(ctx context.Context, entityList admintype.AdminAnnouncementList) error {
	if len(entityList) == 0 {
		return fmt.Errorf("[AdminAnnouncementDao] BatchInsert fail, entityList is empty")
	}

	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return fmt.Errorf("[AdminAnnouncementDao] BatchInsert fail, entityList:%s, err: %v", logs.JSON(entityList), err)
	}
	return nil
}

func (dao *AnnouncementDao) UpdateByID(ctx context.Context, id uint, entity *admintype.AdminAnnouncement) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return fmt.Errorf("[AdminAnnouncementDao] UpdateByID fail, id:%d, entity:%s, err: %v", id, logs.JSON(entity), err)
	}
	return nil
}

func (dao *AnnouncementDao) UpdateMap(ctx context.Context, id uint, updateMap map[string]interface{}) error {
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return fmt.Errorf("[AdminAnnouncementDao] UpdateMap fail, id:%d, updateMap:%s, err: %v", id, logs.JSON(updateMap), err)
	}
	return nil
}

func (dao *AnnouncementDao) Delete(ctx context.Context, id uint) error {
	db := dao.DB(ctx).Table(dao.TableName())
	updatedField := map[string]interface{}{
		"deleted_at": time.Now(),
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return fmt.Errorf("[AdminAnnouncementDao] Delete fail, id:%d, err: %v", id, err)
	}
	return nil
}

func (dao *AnnouncementDao) GetByID(ctx context.Context, id uint) (*admintype.AdminAnnouncement, error) {
	var entity admintype.AdminAnnouncement
	db := dao.DB(ctx).Table(dao.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[AdminAnnouncementDao] GetByID fail, id:%d, err: %v", id, err)
	}
	return &entity, nil
}

func (dao *AnnouncementDao) GetByCond(ctx context.Context, cond *AnnouncementCond) (*admintype.AdminAnnouncement, error) {
	var entity admintype.AdminAnnouncement
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, fmt.Errorf("[AdminAnnouncementDao] GetByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return &entity, nil
}

func (dao *AnnouncementDao) GetListByCond(ctx context.Context, cond *AnnouncementCond) (admintype.AdminAnnouncementList, error) {
	var entityList admintype.AdminAnnouncementList
	db := dao.DB(ctx).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, fmt.Errorf("[AdminAnnouncementDao] GetListByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, nil
}

type Announcement struct {
	//公告id
	ID uint `json:"id"`
	//创建时间
	CreatedAt time.Time `json:"created_at"`
	//创建人uin
	Uin uint `json:"uin"`
	//公司id
	CompanyID uint `json:"company_id"`
	//创建人昵称
	Creator string `json:"creator"`
	//版本tag
	Tag string `json:"tag"`
	//内容
	Content string `json:"content"`
}
type AnnouncementList []Announcement

func (dao *AnnouncementDao) GetPageListByCond(ctx context.Context, cond *AnnouncementCond) (AnnouncementList, int64, error) {
	db := dao.DB(ctx).Model(&admintype.AdminAnnouncement{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)

	for _, v := range cond.Filters {
		switch v.Field {
		case "tag":
			query := fmt.Sprintf("%s.tag LIKE ?", dao.TableName())
			db.Where(query, fmt.Sprintf("%%"+v.Value[0]+"%%"))
		default:
			logs.ErrorContextf(ctx, "unknown field %v for filter", v.Field)
			return nil, 0, fmt.Errorf("[AdminAnnouncementDao] GetPageListByCond invalid filter: %v", v.Field)
		}
	}

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, fmt.Errorf("[AdminAnnouncementDao] GetPageListByCond count fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	if cond.Limit > 0 {
		db.Limit(cond.Limit)
	}
	if cond.Offset > 0 {
		db.Offset(cond.Offset)
	}
	var entityList AnnouncementList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, fmt.Errorf("[AdminAnnouncementDao] GetPageListByCond find fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return entityList, count, nil
}

func (dao *AnnouncementDao) CountByCond(ctx context.Context, cond *AnnouncementCond) (int64, error) {
	db := dao.DB(ctx).Model(&admintype.AdminAnnouncement{}).Table(dao.TableName())

	dao.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, fmt.Errorf("[AdminAnnouncementDao] CountByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return count, nil
}

func (dao *AnnouncementDao) BuildCondition(db *gorm.DB, cond *AnnouncementCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}
}

package task

import (
	"context"
	"fmt"

	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

type TaskCond struct {
	BaseCond
	Filters    []apiobj.Filter
	ID         uint
	IDs        []uint
	TaskType   string
	TaskStatus TaskStatus
}

type TaskDao struct {
	BaseModel
}

func NewTaskDao() *TaskDao {
	return &TaskDao{}
}

func (dao *TaskDao) TableName() string {
	return TableNameCoreTask
}

func (dao *TaskDao) WithTx(db *gorm.DB) *TaskDao {
	return &TaskDao{
		BaseModel: BaseModel{DBClient: db},
	}
}

func (dao *TaskDao) AvgCostTimeByCond(ctx context.Context, cond *TaskCond) (float64, error) {
	db := dao.DB(ctx).Model(&Task{}).Table(dao.TableName())
	dao.BuildCondition(db, cond)

	var avgCost float64
	if err := db.Select("COALESCE(AVG(cost), 0)").Scan(&avgCost).Error; err != nil {
		return 0, fmt.Errorf("[TaskDao] AvgCostTimeByCond fail, cond:%s, err: %v", logs.JSON(cond), err)
	}
	return avgCost, nil
}

func (dao *TaskDao) BuildCondition(db *gorm.DB, cond *TaskCond) {
	db = dao.BaseModel.BuildBaseCondition(db, dao.TableName(), cond.BaseCond)
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", dao.TableName())
		db.Where(query, cond.ID)
	}
	if len(cond.IDs) > 0 {
		query := fmt.Sprintf("%s.id in ?", dao.TableName())
		db.Where(query, cond.IDs)
	}
	if cond.TaskType != "" {
		query := fmt.Sprintf("%s.task_type = ?", dao.TableName())
		db.Where(query, cond.TaskType)
	}
	if cond.TaskStatus != "" {
		query := fmt.Sprintf("%s.task_status = ?", dao.TableName())
		db.Where(query, cond.TaskStatus)
	}
}

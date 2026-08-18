package forestkeywords

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"gorm.io/gorm"
)

func BaseQuery(db *gorm.DB, query apiobj.PageQuery) *gorm.DB {

	if !query.BeginTime.IsZero() {
		db = db.Where("created_at >= ?", query.BeginTime)
	}
	if !query.EndTime.IsZero() {
		db = db.Where("created_at <= ?", query.EndTime)
	}

	for _, order := range query.OrderBy {
		db = db.Order(order)
	}

	if !query.ListAll {
		if query.Limit > 0 {
			db = db.Limit(query.Limit)
		}
		if query.Offset >= 0 {
			db = db.Offset(query.Offset)
		}
	}
	return db
}

type BaseModel struct {
	DBClient *gorm.DB
}

type BaseCond struct {
	ID          uint
	IDs         []uint
	CompanyID   uint
	IsDelete    bool
	Offset      int
	Limit       int
	Uin         uint
	BeginTime   time.Time
	EndTime     time.Time
	OrderBy     []string
	OrCondition OrCondition
}

type OrCondition struct {
	Conditions []string
	Args       []any
}

// DB 获取DB
func (m *BaseModel) DB(ctx context.Context) *gorm.DB {
	if m.DBClient != nil {
		return m.DBClient.WithContext(ctx)
	}

	return dbutil.Knownow().WithContext(ctx)
}

func (m *BaseModel) BuildBaseCondition(db *gorm.DB, tableName string, cond BaseCond) *gorm.DB {
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", tableName)
		db.Where(query, cond.ID)
	}
	if len(cond.IDs) > 0 {
		query := fmt.Sprintf("%s.id in ?", tableName)
		db.Where(query, cond.IDs)
	}
	if cond.CompanyID > 0 {
		query := fmt.Sprintf("%s.company_id = ?", tableName)
		db.Where(query, cond.CompanyID)
	}
	if cond.Uin > 0 {
		query := fmt.Sprintf("%s.uin = ?", tableName)
		db.Where(query, cond.Uin)
	}
	if !cond.BeginTime.IsZero() {
		query := fmt.Sprintf("%s.created_at >= ?", tableName)
		db.Where(query, cond.BeginTime)
	}
	if !cond.EndTime.IsZero() {
		query := fmt.Sprintf("%s.created_at <= ?", tableName)
		db.Where(query, cond.EndTime)
	}
	if cond.IsDelete {
		db.Unscoped()
	}

	if len(cond.OrderBy) > 0 {
		db.Order(strings.Join(cond.OrderBy, ","))
	}

	if len(cond.OrCondition.Conditions) > 0 {
		if len(cond.OrCondition.Args) == len(cond.OrCondition.Conditions) {
			whereClause := strings.Join(cond.OrCondition.Conditions, " OR ")
			db = db.Where("("+whereClause+")", cond.OrCondition.Args...)
		}
	}
	return db
}

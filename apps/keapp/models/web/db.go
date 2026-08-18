package web

import (
	"context"

	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"gorm.io/gorm"
)

const (
	TableNameKeWebResource  = "ke_web_resource"
	TableNameKeWebCrawlRule = "ke_web_crawl_rule"
	TableNameKeCrawlTask    = "ke_crawl_task"
)

type BaseModel struct {
	DBClient *gorm.DB
}

func (m *BaseModel) DB(ctx context.Context) *gorm.DB {
	if m.DBClient != nil {
		return m.DBClient.WithContext(ctx)
	}
	return dbutil.Knownow().WithContext(ctx)
}

func InitDB() error {
	return dbtools.InitModel(dbutil.Knownow(),
		&KeWebResource{},
		&KeWebCrawlRule{},
		&KeCrawlTask{},
	)
}

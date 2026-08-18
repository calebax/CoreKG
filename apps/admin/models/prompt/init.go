package prompt

import (
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"github.com/ygpkg/yg-go/prompt/model"
	"gorm.io/gorm"
)

// InitModel 初始化 core_prompt 与 core_prompt_version 表的 AutoMigrate
func InitModel(db *gorm.DB) error {
	return dbtools.InitModel(db,
		&model.CorePrompt{},
		&model.CorePromptVersion{},
	)
}

// NewPromptDao 创建 PromptDao 实例，使用 Core DB 连接
func NewPromptDao() *model.PromptDao {
	return model.NewPromptDao(dbutil.Core())
}

// NewPromptVersionDao 创建 PromptVersionDao 实例，使用 Core DB 连接
func NewPromptVersionDao() *model.PromptVersionDao {
	return model.NewPromptVersionDao(dbutil.Core())
}

// CoreDB 返回 Core 数据库连接，供事务等场景使用
func CoreDB() *gorm.DB {
	return dbutil.Core()
}

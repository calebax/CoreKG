package dbplugins

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/go-sql-driver/mysql"
	"gorm.io/gorm"
)

type Plugin struct {
	Type DatabaseType
	PluginHandler
}

type PluginHandler interface {
	IsAvailable(ctx context.Context, config *PluginConfig) (bool, error)
	GetDatabases(ctx context.Context, config *PluginConfig) ([]string, error)
	GetStorageGroups(ctx context.Context, config *PluginConfig, opt *QueryOption) (*StorageGroupRes, error)
	GetStorageUnits(ctx context.Context, config *PluginConfig, schema string, opt *QueryOption) (*StorageUnitRes, error)
}

type Credentials struct {
	ConnectionID string
	Type         string
	Hostname     string
	Username     string
	Password     string
	Database     string
	Port         uint
	Advanced     []Record
	IsProfile    bool
}

func (c *Credentials) BuildConnectionKey() string {
	return fmt.Sprintf("%s_%s_%s", c.Type, c.ConnectionID, c.Database)
}

type ExternalModel struct {
	Type  string
	Token string
}

type PluginConfig struct {
	Credentials   *Credentials
	ExternalModel *ExternalModel
}

type StorageGroupRes struct {
	List  []StorageGroup `json:"list"`
	Total int64          `json:"total"`
}

type StorageGroup struct {
	// Name 存储分组名称
	Name string `json:"name"`
	// Attributes 存储分组属性
	Attributes []Record `json:"attributes"`
}

type StorageUnitRes struct {
	List  []StorageUnit `json:"list"`
	Total int64         `json:"total"`
}

type StorageUnit struct {
	// Name 存储单元名称
	Name string `json:"name"`
	// TableAttributes 表属性
	TableAttributes []Record `json:"table_attributes"`
	// ColumnAttributes 列属性
	ColumnAttributes []Record `json:"column_attributes"`
}

type Record struct {
	// Key 键
	Key string `json:"key"`
	// Value 值
	Value string `json:"value"`
	// Extra 额外信息
	Extra map[string]string `json:"extra"`
}

type QueryOption struct {
	Limit      int
	Offset     int
	OrderRules []string
	Filters    []Filter
}

type FilterKey string

const (
	FilterKeySchema  FilterKey = "schema"
	FilterKeySchemas FilterKey = "schemas"
	FilterKeyTable   FilterKey = "table"
	FilterKeyTables  FilterKey = "tables"
)

type Filter struct {
	// Key 键
	Key FilterKey `json:"key"`
	// Values 值
	Values []string `json:"values"`
}

type DBOperation[T any] func(*gorm.DB) (T, error)
type DBCreationFunc func(ctx context.Context, pluginConfig *PluginConfig) (*gorm.DB, error)

// WithConnection handles database connection lifecycle and executes the operation
func WithConnection[T any](ctx context.Context, config *PluginConfig, DB DBCreationFunc, operation DBOperation[T]) (T, error) {
	db, err := DB(ctx, config)
	if err != nil {
		var zero T
		return zero, err
	}
	sqlDb, err := db.DB()
	if err != nil {
		var zero T
		return zero, err
	}
	defer sqlDb.Close()
	return operation(db)
}

// type GetStorageGroupsOption struct {
// 	Schemas []string
// }

// type GetStorageUnitsOption struct {
// 	Tables []string
// }

// FriendlyMySQLError 将 MySQL 错误转换为用户友好的中文信息。
func FriendlyMySQLError(err error) string {
	if err == nil {
		return ""
	}

	// 检查错误是否是 MySQL 错误
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) {
		switch mysqlErr.Number {
		case 1045: // 拒绝访问
			return "数据库连接失败：用户名或密码错误。请检查您的登录凭证。"
		case 1049: // 未知数据库
			return "数据库连接失败：指定的数据库不存在。请确认数据库名称是否正确。"
		case 1146: // 表不存在
			return fmt.Sprintf("操作失败：表 '%s' 不存在。请联系管理员。", mysqlErr.Message)
		case 1062: // 主键重复
			return "数据添加失败：此记录已存在，请勿重复提交。"
		case 1064: // SQL 语法错误
			return "操作失败：您的查询语句存在语法错误。请检查您的输入。"
		case 1054: // 未知列
			return fmt.Sprintf("操作失败：列名错误。请确认表是否存在该列。错误详情：%s", mysqlErr.Message)
		case 1105: // 一般错误
			return "操作失败：服务器内部错误。请稍后重试或联系技术支持。"
		case 1216, 1217: // 外键约束失败
			return "数据操作失败：违反了外键约束。请确认关联的数据是否存在。"
		case 1364: // 字段没有默认值
			return "数据添加失败：某些必填字段没有提供值。"
		default:
			return fmt.Sprintf("数据库操作失败：发生未知错误，错误代码 %d。请联系管理员。错误详情：%s", mysqlErr.Number, mysqlErr.Message)
		}
	}

	// 如果不是 MySQL 错误，返回原始错误信息
	if errors.Is(err, sql.ErrNoRows) {
		return "查询失败：没有找到匹配的数据。"
	}

	return fmt.Sprintf("操作失败：发生意外错误。错误详情：%s", err.Error())
}

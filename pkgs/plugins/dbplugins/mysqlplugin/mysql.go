package mysqlplugin

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"
	"strings"

	"github.com/insmtx/corekg/pkgs/plugins/dbplugins"
	"github.com/insmtx/corekg/pkgs/plugins/dbplugins/gormplugin"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

type MySQLPlugin struct {
	gormplugin.GormPlugin
}

func NewMySQLPlugin() *dbplugins.Plugin {
	mysqlPlugin := &MySQLPlugin{}
	mysqlPlugin.Type = dbplugins.DatabaseTypeMySQL
	mysqlPlugin.PluginHandler = mysqlPlugin
	mysqlPlugin.GormPluginHandler = mysqlPlugin
	return &mysqlPlugin.Plugin
}

const (
	defaultSchemaMetaOrderRule   = "ORDER BY SCHEMA_NAME"
	defaultSchemaTablesOrderRule = "ORDER BY TABLE_NAME"
	defaultSchemaColumnOrderRule = "ORDER BY TABLE_NAME, ORDINAL_POSITION"
)

// GetStorageGroups 获取存储分组
func (p *MySQLPlugin) GetStorageGroups(ctx context.Context, config *dbplugins.PluginConfig, opt *dbplugins.QueryOption) (*dbplugins.StorageGroupRes, error) {
	return dbplugins.WithConnection(ctx, config, p.DB, func(db *gorm.DB) (*dbplugins.StorageGroupRes, error) {
		var schemas []string
		if opt != nil {
			for _, v := range opt.Filters {
				switch v.Key {
				case dbplugins.FilterKeySchemas:
					schemas = v.Values

				}
			}
		}
		statSql := p.BuildSchemaStatSQL(schemas)

		type StatItem struct {
			SchemaName string `gorm:"column:schema_name"`
			TableCount int    `gorm:"column:table_count"`
			RowCount   int    `gorm:"column:row_count"`
			TotalSize  int64  `gorm:"column:total_size"`
			DataSize   int64  `gorm:"column:data_size"`
		}
		var statItems []StatItem
		if err := db.WithContext(ctx).Raw(statSql).Scan(&statItems).Error; err != nil {
			logs.ErrorContextf(ctx, "[MySQLPlugin.GetStorageGroups] Failed to get storage groups, err: %s", err)
			return nil, err
		}

		type MetadataItem struct {
			SchemaName string `gorm:"column:schema_name"`
			Charset    string `gorm:"column:charset"`
			Collation  string `gorm:"column:collation"`
		}
		var metadata []MetadataItem
		sql := p.BuildSchemaMetadataSQL(schemas)

		var total int64
		countSQL := fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS subquery;", strings.TrimRight(sql, " \n\t;"))
		if err := db.WithContext(ctx).Raw(countSQL).Count(&total).Error; err != nil {
			logs.ErrorContextf(ctx, "[MySQLPlugin.GetStorageGroups] Failed to count storage groups, err: %s", err)
			return nil, err
		}

		if opt != nil {
			// 去掉末尾分号和空格
			sql = strings.TrimRight(sql, " \n\t;")
			if opt.Limit > 0 {
				sql += fmt.Sprintf(" LIMIT %d", opt.Limit)
			}
			if opt.Offset > 0 {
				sql += fmt.Sprintf(" OFFSET %d", opt.Offset)
			}
			sql += ";"
		}
		if err := db.WithContext(ctx).Raw(sql).Scan(&metadata).Error; err != nil {
			logs.ErrorContextf(ctx, "[MySQLPlugin.GetStorageGroups] Failed to get storage groups, err: %s", err)
			return nil, err
		}
		statMap := make(map[string]StatItem)
		for _, v := range statItems {
			statMap[v.SchemaName] = v
		}
		var storageGroups []dbplugins.StorageGroup
		for _, v := range metadata {
			stat := statMap[v.SchemaName]
			attributes := []dbplugins.Record{
				{
					Key:   dbplugins.RecordKeyCharset,
					Value: v.Charset,
				},
				{
					Key:   dbplugins.RecordKeyCollation,
					Value: v.Collation,
				},
				{
					Key:   dbplugins.RecordKeyTableCount,
					Value: strconv.Itoa(stat.TableCount),
				},
				{
					Key:   dbplugins.RecordKeyRowCount,
					Value: strconv.Itoa(stat.RowCount),
				},
				{
					Key:   dbplugins.RecordKeyTotalSize,
					Value: fmt.Sprintf("%.d", stat.TotalSize),
				},
				{
					Key:   dbplugins.RecordKeyDataSize,
					Value: fmt.Sprintf("%.d", stat.DataSize),
				},
			}

			storageGroups = append(storageGroups, dbplugins.StorageGroup{
				Name:       v.SchemaName,
				Attributes: attributes,
			})

		}
		return &dbplugins.StorageGroupRes{
			List:  storageGroups,
			Total: total,
		}, nil
	})

}

func (p *MySQLPlugin) BuildSchemaTableMetadataSQL(schema string, queryOption *dbplugins.QueryOption) string {
	sql := `
		SELECT
			TABLE_NAME AS table_name,
			TABLE_TYPE AS table_type,
			IFNULL(DATA_LENGTH + INDEX_LENGTH, 0) AS total_size,
			IFNULL(DATA_LENGTH, 0) AS data_size,
			IFNULL(TABLE_ROWS, 0) AS row_count
		FROM
			INFORMATION_SCHEMA.TABLES
		WHERE
			TABLE_SCHEMA = ` + "'" + schema + "'"
	if queryOption != nil {
		for _, filter := range queryOption.Filters {
			sql += p.buildFilterCondition(filter)
		}
		sql += " " + p.buildOrderBy(queryOption.OrderRules, defaultSchemaTablesOrderRule)
		// if queryOption.Limit > 0 {
		// 	sql += fmt.Sprintf(" LIMIT %d", queryOption.Limit)
		// }
		// if queryOption.Offset > 0 {
		// 	sql += fmt.Sprintf(" OFFSET %d", queryOption.Offset)
		// }
		sql += ";"
	}
	return sql
}

func (p *MySQLPlugin) BuildSchemaTableColumnMetadataSQL(schema string, queryOption *dbplugins.QueryOption) string {
	sql := `SELECT TABLE_NAME AS table_name, COLUMN_NAME AS column_name, DATA_TYPE AS data_type, COLUMN_TYPE AS column_type, COLUMN_COMMENT AS column_comment
			FROM INFORMATION_SCHEMA.COLUMNS
			WHERE TABLE_SCHEMA = ` + "'" + schema + "'"
	if queryOption != nil {
		for _, filter := range queryOption.Filters {
			sql += p.buildFilterCondition(filter)
		}
		sql += ";"
	}

	return sql
}

func (p *MySQLPlugin) GetDatabases(ctx context.Context, config *dbplugins.PluginConfig) ([]string, error) {
	return dbplugins.WithConnection(ctx, config, p.DB, func(db *gorm.DB) ([]string, error) {
		sql := `SELECT SCHEMA_NAME FROM INFORMATION_SCHEMA.SCHEMATA`
		var databases []string
		if err := db.WithContext(ctx).Raw(sql).Scan(&databases).Error; err != nil {
			logs.ErrorContextf(ctx, "[MySQLPlugin.GetDatabases] Failed to get databases, err: %s", err)
			return nil, err
		}
		return databases, nil
	})
}

func (p *MySQLPlugin) BuildSchemaStatSQL(schemas []string) string {
	sql := `SELECT table_schema AS 'schema_name', COUNT(table_name) AS 'table_count', SUM(table_rows) AS 'row_count', SUM(data_length + index_length) AS 'total_size', SUM(data_length) AS 'data_size'
		 FROM information_schema.tables`
	if len(schemas) > 0 {
		quotedSchemas := make([]string, 0, len(schemas))
		for _, s := range schemas {
			quotedSchemas = append(quotedSchemas, "'"+s+"'")
		}
		sql += " WHERE table_schema IN (" + strings.Join(quotedSchemas, ",") + ")"
	}
	sql = sql + " GROUP BY table_schema ORDER BY table_schema DESC;"
	return sql
}

func (p *MySQLPlugin) BuildSchemaMetadataSQL(schemas []string) string {
	sql := "SELECT SCHEMA_NAME AS schema_name, DEFAULT_CHARACTER_SET_NAME AS charset, DEFAULT_COLLATION_NAME AS collation FROM INFORMATION_SCHEMA.SCHEMATA"
	if len(schemas) > 0 {
		quotedSchemas := make([]string, 0, len(schemas))
		for _, s := range schemas {
			quotedSchemas = append(quotedSchemas, "'"+s+"'")
		}
		sql += " WHERE SCHEMA_NAME IN (" + strings.Join(quotedSchemas, ",") + ")"
	}
	sql = sql + " ORDER BY SCHEMA_NAME DESC;"
	return sql
}

func (p *MySQLPlugin) GetTableNameAndAttributes(ctx context.Context, rows *sql.Rows, db *gorm.DB) (string, []dbplugins.Record) {
	var tableName, tableType string
	var totalSize, dataSize int64
	var rowCount int64
	if err := rows.Scan(&tableName, &tableType, &totalSize, &dataSize, &rowCount); err != nil {
		logs.ErrorContextf(ctx, "[MySQLPlugin.GetTableNameAndAttributes] Failed to get table name and attributes, err: %s", err)
		return "", nil
	}

	// 如果行数为 0 或者可疑地过低，就执行一次 SELECT COUNT，代价应该不会太高
	// MySQL 的 TABLE_ROWS 只是一个估算值，可能非常不准确
	if rowCount < 100 {
		var actualCount int64
		countQuery := db.Table(tableName).Select("COUNT(*)")
		if err := countQuery.Scan(&actualCount).Error; err == nil {
			rowCount = actualCount
		}
	}

	attributes := []dbplugins.Record{
		{Key: dbplugins.RecordKeyTableType, Value: tableType},
		{Key: dbplugins.RecordKeyTotalSize, Value: fmt.Sprintf("%d", totalSize)},
		{Key: dbplugins.RecordKeyDataSize, Value: fmt.Sprintf("%.d", dataSize)},
		{Key: dbplugins.RecordKeyRowCount, Value: fmt.Sprintf("%d", rowCount)},
	}
	return tableName, attributes
}

// buildFilterCondition 根据Filter构建WHERE条件
func (p *MySQLPlugin) buildFilterCondition(filter dbplugins.Filter) string {
	if len(filter.Values) == 0 {
		return ""
	}

	switch filter.Key {
	case dbplugins.FilterKeySchema:
		// 单个schema
		return fmt.Sprintf(" AND TABLE_SCHEMA = '%s'", filter.Values[0])

	case dbplugins.FilterKeySchemas:
		// 多个schema
		quotedSchemas := make([]string, len(filter.Values))
		for i, schema := range filter.Values {
			quotedSchemas[i] = "'" + schema + "'"
		}
		return fmt.Sprintf(" AND TABLE_SCHEMA IN (%s)", strings.Join(quotedSchemas, ","))

	case dbplugins.FilterKeyTables:
		// 精确表名匹配
		quotedTables := make([]string, len(filter.Values))
		for i, table := range filter.Values {
			quotedTables[i] = "'" + table + "'"
		}
		return fmt.Sprintf(" AND TABLE_NAME IN (%s)", strings.Join(quotedTables, ","))

	case dbplugins.FilterKeyTable:
		// LIKE 处理
		return fmt.Sprintf(" AND TABLE_NAME LIKE '%%%s%%'", filter.Values[0])
	}
	return ""
}

// buildOrderBy 构建排序条件，返回完整的 ORDER BY 表达式
// 传参示例：[]string{"column_name DESC", "table_name", "ordinal_position ASC"}
func (p *MySQLPlugin) buildOrderBy(orderBy []string, defaultOrderBy string) string {
	fieldMap := map[string]string{
		"table_name":       "TABLE_NAME",
		"column_name":      "COLUMN_NAME",
		"data_type":        "DATA_TYPE",
		"column_type":      "COLUMN_TYPE",
		"ordinal_position": "ORDINAL_POSITION",
	}

	validOrders := make([]string, 0, len(orderBy))

	for _, field := range orderBy {
		// 去除首尾空白字符并转换为小写
		field = strings.TrimSpace(strings.ToLower(field))
		if field == "" {
			continue
		}

		// 按空白字符分割，支持 "field DESC" 或 "field ASC" 或单独的 "field" 格式
		parts := strings.Fields(field)
		if len(parts) == 0 {
			continue
		}

		fieldName := parts[0]

		// 检查字段是否在映射表中
		if dbField, exists := fieldMap[fieldName]; exists {
			orderStr := dbField

			// 处理排序方向
			if len(parts) >= 2 {
				direction := strings.ToUpper(parts[1])
				if direction == "DESC" || direction == "ASC" {
					orderStr += " " + direction
				} else {
					// 如果第二个参数不是有效的排序方向，默认使用 ASC
					orderStr += " ASC"
				}
			} else {
				// 如果没有指定排序方向，默认使用 ASC
				orderStr += " ASC"
			}

			validOrders = append(validOrders, orderStr)
		}
		// 如果字段不在映射表中，直接忽略（也可以选择记录日志）
	}

	if len(validOrders) > 0 {
		return "ORDER BY " + strings.Join(validOrders, ", ")
	}

	// 如果没有有效的排序字段，使用默认排序
	return defaultOrderBy
}

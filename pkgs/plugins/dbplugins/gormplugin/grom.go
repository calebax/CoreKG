package gormplugin

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/insmtx/corekg/pkgs/plugins/dbplugins"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

type GormPlugin struct {
	dbplugins.Plugin
	GormPluginHandler
}

type GormPluginHandler interface {
	DB(ctx context.Context, config *dbplugins.PluginConfig) (*gorm.DB, error)
	BuildSchemaTableMetadataSQL(schema string, queryOption *dbplugins.QueryOption) string
	BuildSchemaTableColumnMetadataSQL(schema string, queryOption *dbplugins.QueryOption) string
	BuildSchemaMetadataSQL(schemas []string) string
	BuildSchemaStatSQL(schemas []string) string
	GetTableNameAndAttributes(ctx context.Context, rows *sql.Rows, db *gorm.DB) (string, []dbplugins.Record)
}

// GetStorageUnits 获取存储单元
func (p *GormPlugin) GetStorageUnits(ctx context.Context, config *dbplugins.PluginConfig, schema string, opt *dbplugins.QueryOption) (*dbplugins.StorageUnitRes, error) {
	return dbplugins.WithConnection(ctx, config, p.DB, func(db *gorm.DB) (*dbplugins.StorageUnitRes, error) {
		sql := p.BuildSchemaTableMetadataSQL(schema, opt)

		var total int64
		countSQL := fmt.Sprintf("SELECT COUNT(*) FROM (%s) AS subquery;", strings.TrimRight(sql, " \n\t;"))
		if err := db.WithContext(ctx).Raw(countSQL).Count(&total).Error; err != nil {
			logs.ErrorContextf(ctx, "[GormPlugin.GetStorageUnits] Failed to count table schema, err: %s", err)
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

		rows, err := db.WithContext(ctx).Raw(sql).Rows()
		if err != nil {
			logs.ErrorContextf(ctx, "[GormPlugin.GetStorageUnits] Failed to get storage units, err: %s", err)
			return nil, err
		}
		defer rows.Close()

		tablesWithColumns, err := p.GetTableSchema(ctx, db, schema, opt)
		if err != nil {
			logs.ErrorContextf(ctx, "[GormPlugin.GetStorageUnits] Failed to get table schema, err: %s", err)
			return nil, err
		}

		var storageUnits []dbplugins.StorageUnit
		for rows.Next() {
			tableName, tableAttributes := p.GetTableNameAndAttributes(ctx, rows, db)
			if len(tableAttributes) == 0 && tableName == "" {
				continue
			}

			storageUnits = append(storageUnits, dbplugins.StorageUnit{
				Name:             tableName,
				TableAttributes:  tableAttributes,
				ColumnAttributes: tablesWithColumns[tableName],
			})
		}
		return &dbplugins.StorageUnitRes{
			Total: total,
			List:  storageUnits,
		}, nil
	})
}

func (p *GormPlugin) GetTableSchema(ctx context.Context, db *gorm.DB, schema string, opt *dbplugins.QueryOption) (map[string][]dbplugins.Record, error) {
	var result []struct {
		TableName     string `gorm:"column:table_name"`
		ColumnName    string `gorm:"column:column_name"`
		DataType      string `gorm:"column:data_type"`
		ColumnType    string `gorm:"column:column_type"`
		ColumnComment string `gorm:"column:column_comment"`
	}

	sql := p.BuildSchemaTableColumnMetadataSQL(schema, opt)

	if err := db.WithContext(ctx).Raw(sql).Scan(&result).Error; err != nil {
		logs.ErrorContextf(ctx, "[GormPlugin.GetTableSchema] Failed to get table schema, err: %s", err)
		return nil, err
	}

	tableColumnsMap := make(map[string][]dbplugins.Record)
	for _, row := range result {
		tableColumnsMap[row.TableName] = append(tableColumnsMap[row.TableName],
			dbplugins.Record{
				Key:   row.ColumnName,
				Value: row.DataType,
				Extra: map[string]string{
					dbplugins.RecordKeyColumnType:    row.ColumnType,
					dbplugins.RecordKeyColumnComment: row.ColumnComment,
				}},
		)
	}

	return tableColumnsMap, nil
}

package mysqlplugin

import (
	"context"
	"net"
	"strconv"

	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/insmtx/corekg/pkgs/plugins/dbplugins"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

func (p *MySQLPlugin) DB(ctx context.Context, config *dbplugins.PluginConfig) (*gorm.DB, error) {
	connectionInput, err := p.ParseConnectionConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	mysqlConfig := mysqldriver.NewConfig()
	mysqlConfig.User = connectionInput.Username
	mysqlConfig.Passwd = connectionInput.Password
	mysqlConfig.Net = "tcp"
	mysqlConfig.Addr = net.JoinHostPort(connectionInput.Hostname, strconv.Itoa(connectionInput.Port))
	mysqlConfig.DBName = connectionInput.Database
	mysqlConfig.AllowCleartextPasswords = connectionInput.AllowClearTextPasswords
	mysqlConfig.ParseTime = connectionInput.ParseTime
	mysqlConfig.Loc = connectionInput.Loc
	mysqlConfig.Params = connectionInput.ExtraOptions

	db, err := gorm.Open(mysql.Open(mysqlConfig.FormatDSN()), &gorm.Config{})
	if err != nil {
		logs.WarnContextf(ctx, "[MySQLPlugin.DB] Failed to connect to database, hostname: %s, port: %d, database: %s, err: %s", connectionInput.Hostname, connectionInput.Port, connectionInput.Database, err)
		return nil, err
	}
	db.Logger = logs.GetGorm("gorm")
	return db, nil
}

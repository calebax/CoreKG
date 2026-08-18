package gormplugin

import (
	"context"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/insmtx/corekg/pkgs/plugins/dbplugins"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

type ConnectionInput struct {
	//common
	Username string `validate:"required"`
	Password string `validate:"required"`
	Database string `validate:"required"`
	Hostname string `validate:"required"`
	Port     int    `validate:"required"`

	//mysql/mariadb
	ParseTime               bool           `validate:"boolean"`
	Loc                     *time.Location `validate:"required"`
	AllowClearTextPasswords bool           `validate:"boolean"`

	ConnectionTimeout int

	ExtraOptions map[string]string `validate:"omitnil"`
}

func (p *GormPlugin) IsAvailable(ctx context.Context, config *dbplugins.PluginConfig) (bool, error) {
	addr := net.JoinHostPort(config.Credentials.Hostname, strconv.Itoa(int(config.Credentials.Port)))
	conn, err := net.DialTimeout("tcp", addr, 5*time.Second)
	if err != nil {
		logs.WarnContextf(ctx, "[GormPlugin.IsAvailable] Failed to dial %s for database type %s, err: %v", addr, p.Type, err)
		return false, nil
	}
	conn.Close()

	config.Credentials.ConnectionID = ""
	available, err := dbplugins.WithConnection(ctx, config, p.DB, func(db *gorm.DB) (bool, error) {
		sqlDb, err := db.WithContext(ctx).DB()
		if err != nil {
			logs.WarnContextf(ctx, "[GormPlugin.IsAvailable] Failed to get SQL DB instance for database type %s", p.Type)
			return false, err
		}
		if err = sqlDb.Ping(); err != nil {
			logs.WarnContextf(ctx, "[GormPlugin.IsAvailable] Failed to ping database for type %s", p.Type)
			return false, nil
		}
		defer sqlDb.Close()
		return true, nil
	})
	if err != nil {
		return false, err
	}

	return available, nil
}

func (p *GormPlugin) ParseConnectionConfig(ctx context.Context, config *dbplugins.PluginConfig) (*ConnectionInput, error) {

	parseTime, err := strconv.ParseBool(dbplugins.GetRecordValueOrDefault(config.Credentials.Advanced, dbplugins.ConfigKeyParseTime, dbplugins.ConfigDefaultParseTime))
	if err != nil {
		logs.ErrorContextf(ctx, "[GormPlugin.ParseConnectionConfig] Failed to parse parseTime setting for database type %s", p.Type)
		return nil, err
	}
	loc, err := time.LoadLocation(dbplugins.GetRecordValueOrDefault(config.Credentials.Advanced, dbplugins.ConfigKeyLoc, dbplugins.ConfigDefaultLoc))
	if err != nil {
		logs.ErrorContextf(ctx, "[GormPlugin.ParseConnectionConfig] Failed to load time location for database type %s", p.Type)
		return nil, err
	}

	// 基本连接信息
	database := url.PathEscape(config.Credentials.Database)
	username := url.PathEscape(config.Credentials.Username)
	password := url.PathEscape(config.Credentials.Password)
	hostname := url.PathEscape(config.Credentials.Hostname)

	return &ConnectionInput{
		Username:  username,
		Password:  password,
		Database:  database,
		Hostname:  hostname,
		Port:      int(config.Credentials.Port),
		ParseTime: parseTime,
		Loc:       loc,
	}, nil
}

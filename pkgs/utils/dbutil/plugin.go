package dbutil

import (
	"github.com/insmtx/corekg/pkgs/plugins/dbplugins"
	"github.com/insmtx/corekg/pkgs/plugins/dbplugins/mysqlplugin"
)

var dbPluginEngin *dbplugins.Engine

func InitializePlugins() {
	dbPluginEngin = &dbplugins.Engine{}
	dbPluginEngin.RegistryPlugin(mysqlplugin.NewMySQLPlugin())
}

func GetDBPluginEngine() *dbplugins.Engine {
	return dbPluginEngin
}

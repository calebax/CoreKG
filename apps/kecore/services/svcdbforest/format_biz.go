package svcdbforest

import (
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/plugins/dbplugins"
)

type BuildEntityReq struct {
	CompanyID          uint
	Uin                uint
	PluginDatabaseType dbplugins.DatabaseType
	PluginConfig       *dbplugins.PluginConfig
	ForestEntity       *foresttype.KnownowForest
}

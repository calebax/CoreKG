package dbutil

import (
	"context"
	"testing"

	"github.com/insmtx/corekg/pkgs/plugins/dbplugins"
	"github.com/stretchr/testify/assert"
)

func TestDBPluginIsAvailable(t *testing.T) {
	InitializePlugins()
	config := &dbplugins.PluginConfig{
		Credentials: &dbplugins.Credentials{
			ConnectionID: "test-connect-id",
			Type:         "test-type",
			Hostname:     "127.0.0.1",
			Username:     "root",
			Password:     "123456",
			Database:     "demo",
			Port:         3306,
			Advanced: []dbplugins.Record{
				{Key: dbplugins.ConfigKeyParseTime, Value: "True"},
				{Key: dbplugins.ConfigKeyLoc, Value: "Local"},
			},
		},
	}
	ctx := context.Background()
	available, _ := GetDBPluginEngine().ChoosePlugin(dbplugins.DatabaseTypeMySQL).IsAvailable(ctx, config)

	assert.True(t, available)
}

func TestDBPluginGetDatabases(t *testing.T) {
	InitializePlugins()
	config := &dbplugins.PluginConfig{
		Credentials: &dbplugins.Credentials{
			ConnectionID: "test-connect-id",
			Type:         "test-type",
			Hostname:     "127.0.0.1",
			Username:     "root",
			Password:     "123456",
			Database:     "demo",
			Port:         3306,
			Advanced: []dbplugins.Record{
				{Key: dbplugins.ConfigKeyParseTime, Value: "True"},
				{Key: dbplugins.ConfigKeyLoc, Value: "Local"},
			},
		},
	}
	ctx := context.Background()
	databases, err := GetDBPluginEngine().ChoosePlugin(dbplugins.DatabaseTypeMySQL).GetDatabases(ctx, config)

	assert.Nil(t, err)
	t.Log(databases)
}

func TestDBPluginGetStorageGroups(t *testing.T) {
	InitializePlugins()
	config := &dbplugins.PluginConfig{
		Credentials: &dbplugins.Credentials{
			ConnectionID: "test-connect-id",
			Type:         "test-type",
			Hostname:     "127.0.0.1",
			Username:     "root",
			Password:     "123456",
			Database:     "demo",
			Port:         3306,
			Advanced: []dbplugins.Record{
				{Key: dbplugins.ConfigKeyParseTime, Value: "True"},
				{Key: dbplugins.ConfigKeyLoc, Value: "Local"},
			},
		},
	}
	ctx := context.Background()
	databases, err := GetDBPluginEngine().ChoosePlugin(dbplugins.DatabaseTypeMySQL).GetStorageGroups(ctx, config, &dbplugins.QueryOption{})

	assert.Nil(t, err)
	t.Log(databases)
}

func TestDBPluginGetStorageUnits(t *testing.T) {
	InitializePlugins()
	config := &dbplugins.PluginConfig{
		Credentials: &dbplugins.Credentials{
			ConnectionID: "test-connect-id",
			Type:         "test-type",
			Hostname:     "127.0.0.1",
			Username:     "root",
			Password:     "123456",
			Database:     "demo",
			Port:         3306,
			Advanced: []dbplugins.Record{
				{Key: dbplugins.ConfigKeyParseTime, Value: "True"},
				{Key: dbplugins.ConfigKeyLoc, Value: "Local"},
			},
		},
	}
	ctx := context.Background()
	storageUnits, err := GetDBPluginEngine().ChoosePlugin(dbplugins.DatabaseTypeMySQL).GetStorageUnits(ctx, config, "demo", &dbplugins.QueryOption{})

	assert.Nil(t, err)
	t.Log(storageUnits)
}

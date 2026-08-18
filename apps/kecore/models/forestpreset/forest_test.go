package forestpreset

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/insmtx/corekg/apps/kecore/models/fs"
	"github.com/insmtx/corekg/apps/kecore/models/nbgraph"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"github.com/ygpkg/yg-go/logs"
)

func TestPreset(t *testing.T) {
	// 通过环境变量读取真实密码，避免把凭据写入源码。
	password := os.Getenv("PRESET_MYSQL_PASSWORD")
	if password == "" {
		t.Skip("PRESET_MYSQL_PASSWORD 未设置，跳过")
	}
	dbtools.InitMultiDBConn(map[string]string{
		"knownow": "mysql://root:" + password + "@CHANGE_ME_HOST:26323/roc_knownow?charset=utf8mb4&parseTime=true&loc=Local",
		"core":    "mysql://root:" + password + "@CHANGE_ME_HOST:26323/roc_core?charset=utf8mb4&parseTime=true&loc=Local",
	})
	ctx := context.Background()
	// 初始化文件存储
	err := fs.InitForestStorage()
	if err != nil {
		logs.FatalContextf(ctx, "[main] InitForestStorage failed, %s", err)
		return
	}

	err = nbgraph.InitNebulaConf(context.Background())
	if err != nil {
		logs.FatalContextf(ctx, "[main] InitNebulaConf failed, %s", err)
		return
	}
	err = PresetForest(ctx, 1, 318) //海雄
	// err = PresetForest(0, 417) // 客户
	// err = PresetForest(0, 175) // 玉荣
	// err = PresetForest(0, 185) // 雨龙
	// err = PresetForest(0, 420) // 客户2
	// err = PresetForest(1, 378) // 宋浩
	if err != nil {
		fmt.Println(err.Error())
	}
}

func TestPresetESForests(t *testing.T) {
	dbtools.InitMultiDBConn(map[string]string{
		// "chat":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"knownow": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"core":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	})
	// 初始化文件存储
	err := fs.InitForestStorage()
	if err != nil {
		logs.FatalContextf(context.Background(), "[main] InitForestStorage failed, %s", err)
		return
	}

	err = nbgraph.InitNebulaConf(context.Background())
	if err != nil {
		logs.FatalContextf(context.Background(), "[main] InitNebulaConf failed, %s", err)
		return
	}
	err = PresetESForests(context.Background(), 72, 581)
	if err != nil {
		fmt.Println(err.Error())
	}

}

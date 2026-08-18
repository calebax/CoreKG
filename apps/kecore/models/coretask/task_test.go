package coretask

import (
	"context"
	"fmt"
	"testing"

	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/fs"
	"github.com/insmtx/corekg/apps/kecore/models/graph"
	"github.com/insmtx/corekg/pkgs/task"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"github.com/ygpkg/yg-go/dbtools/redispool"
	"github.com/ygpkg/yg-go/logs"
)

func TestCoreTask(t *testing.T) {
	dbtools.InitMultiDBConn(map[string]string{
		"core":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=True&loc=Local",
		"knownow": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	})
	_ = redispool.InitRedis("knowledge", "redis")

	// var tasks []*foresttype.KnownowForestFile
	// _ = dbutil.Core().Where("forest_id = ?", 41).Find(&tasks).Error
	// // aa, _ := forest.GetForestByID(41)
	// for _, t := range tasks {
	// 	// dbutil.Core().Create(GenerateSuccessFileTask(t, aa, "test_bucket"))
	// 	task.PushTaskQueue(context.Background(), SuccessFileTask)
	// 	fmt.Println(t)
	// }
	// 初始化文件存储
	err := fs.InitForestStorage()
	if err != nil {
		logs.FatalContextf(context.Background(), "[main] InitForestStorage failed, %s", err)
		return
	}

	var tasks []*task.Task
	_ = dbutil.Core().Where("task_type = ?", "ke.knowledge_task").Find(&tasks).Error
	forestIDSet := make(map[uint]struct{})
	for _, t := range tasks {
		fmt.Println("----------------------t.subject_id----------------------", t.SubjectID)
		fmt.Println(t.SubjectID)
		file, err := forest.GetForestFileByID(t.SubjectID)
		if err != nil {
			continue
		}
		forestIDSet[file.ForestID] = struct{}{}
	}
	for forestID := range forestIDSet {
		graphInfo, err := graph.GetForestGraph(context.Background(), forestID)
		if err != nil {
			continue
		}
		_ = GenerateForestGraphTask(context.Background(), graphInfo, false)
	}
	// PushTaskQueue(context.Background(), "ke.prase_pdf_task")
}

func TestStatus(t *testing.T) {
	dbtools.InitMultiDBConn(map[string]string{
		"core":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=True&loc=Local",
		"knownow": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	})
	// SuccessStatus(context.Background())
	a, _ := ListGraphTaskCount(context.Background(), []uint{256, 254, 10000})
	logs.Infof("%+v", logs.JSON(a))
}

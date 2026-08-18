package graph

import (
	"context"
	"fmt"
	"testing"

	"github.com/insmtx/corekg/apps/account/models/employee"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"github.com/ygpkg/yg-go/logs"
)

func DBInit() {
	if err := dbtools.InitMultiDBConn(map[string]string{
		"chat":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"knownow": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"core":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"account": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	}); err != nil {
		panic(err)
	}
}

func TestGetEdgeTagInfoByTagID(t *testing.T) {
	DBInit()
	res, err := GetEdgeTagInfoByTagID(context.Background(), 3, 3)
	// fmt.Println(a.Edges)
	// logs.Infof("%+v", a.Edges)
	fmt.Println(err)
	logs.InfoContextf(context.Background(), "%+v", res)
}

func TestGraphPermMigrate(t *testing.T) {
	DBInit()
	var gs []*foresttype.ForestGraphInfo
	if err := dbutil.Knownow().
		Where("deleted_at IS NULL").
		Find(&gs).
		Error; err != nil {
		panic(err)
	}
	var rss []*foresttype.KeResourceScope

	cmpAdminMap := make(map[uint][]uint)
	for _, g := range gs {
		empIDs, ok := cmpAdminMap[g.CompanyID]
		if !ok {
			emps, err := employee.GetAdminEmployeeByCompanyUD(g.CompanyID)
			if err != nil {
				panic(err)
			}
			var s []uint
			for _, e := range emps {
				s = append(s, e.Uin)
			}
			cmpAdminMap[g.CompanyID] = append(cmpAdminMap[g.CompanyID], s...)
			empIDs = cmpAdminMap[g.CompanyID]
		}
		rss = append(rss, &foresttype.KeResourceScope{
			ResourceType: foresttype.ResourceTypeGraph,
			ResourceID:   g.ID,
			ScopeType:    foresttype.ScopeTypeCompany,
			ScopeID:      g.CompanyID,
			Action:       foresttype.ActionView,
		})

		for _, emp := range empIDs {
			rss = append(rss, &foresttype.KeResourceScope{
				ResourceType: foresttype.ResourceTypeGraph,
				ResourceID:   g.ID,
				ScopeType:    foresttype.ScopeTypeUser,
				ScopeID:      emp,
				Action:       foresttype.ActionManage,
			})
		}

	}
	if err := dbutil.Knownow().CreateInBatches(rss, len(rss)).Error; err != nil {
		panic(err)
	}
}

func TestMigrateTNodeToNode(t *testing.T) {
	DBInit()
	err := MigrateTNodeToNode(context.Background())
	logs.ErrorContextf(context.Background(), "%v", err)
}

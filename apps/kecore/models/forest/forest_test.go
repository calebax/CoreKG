package forest

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/perm"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
)

var skipTest = true

func TestSplitPath(t *testing.T) {
	paths := []string{
		"a/b/c/d.pdf",
		"a/b/c/",
	}
	for _, path := range paths {
		names := strings.Split(path, "/")
		t.Logf("%d----%v", len(names), names)

		dir, f := filepath.Split(path)
		t.Logf("%v--%v", dir, f)

		l := filepath.SplitList(path)
		t.Logf("%d-------%v", len(l), l)
	}
}

func TestCreateForest(t *testing.T) {
	if skipTest {
		return
	}
	forest_info := &foresttype.KnownowForest{
		CompanyID:   0,
		Uin:         1,
		Name:        "secondforest",
		AvatarUrl:   "req.Request.AvatarUrl",
		Description: "req.Request.Description",
	}
	if _, err := CreateForest(context.Background(), dbutil.Knownow(), forest_info); err != nil {
		fmt.Println(err.Error())
	}
}

// func TestListForest(t *testing.T) {
// 	if skipTest {
// 		return
// 	}
// 	ls, err := ListForest(1)
// 	if err != nil {
// 		t.Logf(err.Error())
// 	}
// 	t.Logf("%+v", ls)
// }

// func TestModifyForest(t *testing.T) {
// 	if skipTest {
// 		return
// 	}
// 	err := ModifyForest(1, "firstforest", "ok1")
// 	if err != nil {
// 		t.Logf(err.Error())
// 	}

// }

//func TestDeleteForest(t *testing.T) {
//	if skipTest {
//		return
//	}
//	err := DeleteForest(1, 1)
//	if err != nil {
//		fmt.Println(err.Error())
//	}
//}

func TestPreviewFile(t *testing.T) {
	if skipTest {
		return
	}
	c, err := PreviewFile("1/4/images.pdf")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	// t.Logf(ct)
	f, err := os.Create("a.pdf")
	if err != nil {
		fmt.Println(err.Error())
		return
	}
	f.Write(c)
	defer f.Close()
}

func TestUnMarshal(t *testing.T) {
	fps := &perm.Set{
		Forest: &foresttype.KnownowForest{
			CompanyID:   0,
			Uin:         1,
			Name:        "secondforest",
			AvatarUrl:   "req.Request.AvatarUrl",
			Description: "req.Request.Description",
		},
		ManagePerm: true,
		UserPerm:   false,
	}

	fpsBts, err := json.Marshal(*fps)
	if err != nil {
		t.Fatal(err.Error())
	}
	fmt.Println(string(fpsBts))
}

func DBInit() {
	if err := dbtools.InitMultiDBConn(map[string]string{
		"knownow": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"core":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"chat":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	}); err != nil {
		panic(err)
	}
}

func TestRefreshForest(t *testing.T) {
	DBInit()
	//if err := RefreshForest(context.TODO(), 101); err != nil {
	//	t.Fatal(err)
	//}

	if err := RefreshForests(context.TODO(), []uint{176}); err != nil {
		t.Fatal(err)
	}

}

func TestGetDirFiles(t *testing.T) {
	DBInit()

	//res, err := GetDirFiles(3082)
	res, err := GetDirsFiles(context.TODO(), []uint{3093, 3094})
	if err != nil {
		t.Fatal(err)
	}
	for _, v := range res {
		fmt.Println(v.Name)
	}
}

func TestMigrateForestPerm(t *testing.T) {
	DBInit()

	var (
		frss = make([]*foresttype.KnownowForest, 0)
		pss  = make([]*foresttype.KnownowForestPublicScope, 0)
		rss  = make([]*foresttype.KeResourceScope, 0)
	)
	if err := dbutil.Knownow().
		Where("deleted_at is null").
		Find(&frss).Error; err != nil {
		t.Fatal(err)
	}

	if err := dbutil.Knownow().
		Where("deleted_at is null").
		Find(&pss).Error; err != nil {
		t.Fatal(err)
	}

	psMap := make(map[uint][]*foresttype.KnownowForestPublicScope)
	for _, scope := range pss {
		if scope != nil {
			psMap[scope.ForestID] = append(psMap[scope.ForestID], scope)
		}
	}

	for _, v := range frss {
		mgIDs := v.ManagerIDs.Slice()

		//construct manager
		for _, mgID := range mgIDs {
			//for manage action, scope_type will be user
			rs := &foresttype.KeResourceScope{
				ResourceType: foresttype.ResourceTypeForest,
				ResourceID:   v.ID,
				ScopeType:    foresttype.ScopeTypeUser,
				ScopeID:      mgID,
				Action:       foresttype.ActionManage,
			}
			//append to insert waited slice
			rss = append(rss, rs)
		}
		//construct viewer
		//for view action, scope_type would be v.company
		//Get resource's public_scope record

		switch v.PublicScope {
		case foresttype.PublicScopeCompany:
			rss = append(rss, &foresttype.KeResourceScope{
				ResourceType: foresttype.ResourceTypeForest,
				ResourceID:   v.ID,
				ScopeType:    foresttype.ScopeTypeCompany,
				ScopeID:      v.CompanyID,
				Action:       foresttype.ActionView,
			})
		case foresttype.PublicScopePrivate:
			rss = append(rss, &foresttype.KeResourceScope{
				ResourceType: foresttype.ResourceTypeForest,
				ResourceID:   v.ID,
				ScopeType:    foresttype.ScopeTypeUser,
				ScopeID:      v.Uin,
				Action:       foresttype.ActionView,
			})
		case foresttype.PublicScopePublic:
			rss = append(rss, &foresttype.KeResourceScope{
				ResourceType: foresttype.ResourceTypeForest,
				ResourceID:   v.ID,
				ScopeType:    foresttype.ScopeTypePublic,
				Action:       foresttype.ActionView,
			})
		case foresttype.PublicScopeCustom:
			psFit := psMap[v.ID]
			for _, pf := range psFit {
				if pf != nil {
					rss = append(rss, &foresttype.KeResourceScope{
						ResourceType: foresttype.ResourceTypeForest,
						ResourceID:   v.ID,
						ScopeType:    foresttype.ScopeTypeUser,
						ScopeID:      pf.ScopeID,
						Action:       foresttype.ActionView,
					})
				}
			}
		}
	}

	if err := dbutil.Knownow().CreateInBatches(&rss, len(rss)).Error; err != nil {
		t.Fatal(err)
	}
}

func TestQueryListForest(t *testing.T) {
	DBInit()
	var resp ForestInfoItemList
	if err := QueryListForest(context.TODO(), apiobj.PageQuery{
		Uin:       581,
		CompanyID: 72,
	}, &resp); err != nil {
		t.Fatal(err)
	}
	for _, v := range resp.Data {
		fmt.Printf("%+v\n", *v)
	}
}

func TestCanUse(t *testing.T) {
	DBInit()

	fmt.Println(CanUseForest(context.TODO(), 164, 479, 18))
	fmt.Println("=======================================================")
}

func TestCanViewForests(t *testing.T) {
	DBInit()
	uins, err := ViewAbleForests(581, 72)
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println(uins)
}

type (
	foo string
	bar string
)

func TestType(t *testing.T) {

	var a foo = "cc"
	var b bar = "cc"
	println(bar(a) == b)

}

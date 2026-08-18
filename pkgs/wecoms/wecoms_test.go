package wecoms

import (
	"os"
	"testing"

	"github.com/xen0n/go-workwx"
)

var (
	testCompanyID       = os.Getenv("TEST_COMPANYID")
	testAppID     int64 = 1000002
	testAppSecret       = os.Getenv("TEST_APP_SECRET")
)

func TestListUsers(t *testing.T) {
	if testCompanyID == "" {
		return
	}
	t.Logf("company_id: %s", testCompanyID)
	cli := workwx.New(testCompanyID).WithApp(testAppSecret, testAppID)
	depts, err := cli.ListAllDepts()
	if err != nil {
		t.Fatal(err)
	}
	for _, dept := range depts {
		t.Logf("dept: %+v", dept)
		uis, err := cli.ListUsersByDeptID(dept.ID, false)
		if err != nil {
			t.Fatal(err)
		}
		for _, u := range uis {
			t.Logf("user: %+v", u)
		}
	}
}

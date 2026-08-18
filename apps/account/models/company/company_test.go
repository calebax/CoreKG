package company

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/employee"
	"github.com/insmtx/corekg/apps/kechat/models/chatagent"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"github.com/ygpkg/yg-go/random"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func TestCreateCompanyWithUsers(t *testing.T) {
	t.Skip()
	ctx := context.TODO()
	if err := dbtools.InitMultiDBConn(map[string]string{
		"account": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	}); err != nil {
		panic(err)
	}

	password := os.Getenv("COREKG_TEST_USER_PASSWORD")
	if password == "" {
		t.Fatalf("COREKG_TEST_USER_PASSWORD 未设置")
	}

	var (
		companyName = os.Getenv("COREKG_TEST_COMPANY_NAME")
		phonePrefix = os.Getenv("COREKG_TEST_USER_PHONE_PREFIX")
		userCount   = 1
		issuer      = "yygu"
	)

	crypedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		panic(err)
	}
	strPwd := string(crypedPassword)

	if err = dbtools.Account().Transaction(func(tx *gorm.DB) error {
		//	1.create company
		cmp, err := CreateCompany(ctx, tx, &CompanyInfo{Name: companyName})
		if err != nil {
			return err
		}
		for i := 0; i < userCount; i++ {
			p := fmt.Sprintf("%v%v", phonePrefix, i)
			//2. create every user
			u := &accounttype.User{
				Phone:    &p,
				Identify: fmt.Sprintf("%v%v", "dx", random.String(4)),
				Password: &strPwd,
				Name:     fmt.Sprintf("%v_%v", companyName, i),
			}

			if err := tx.Create(u).Error; err != nil {
				return err
			}
			//3. create every uins
			ui := &accounttype.UserIdentification{
				UserID:      u.ID,
				SubjectType: accounttype.SubjectTypeCompany,
				SubjectID:   cmp.ID,
				Issuer:      issuer,
				UinStatus:   accounttype.UinStatusNormal,
				Name:        u.Name,
			}
			if err := tx.Create(ui).Error; err != nil {
				return err
			}
			var role accounttype.SysRole
			if i == 0 {
				role = accounttype.SysRoleSysAdmin
			} else {
				role = accounttype.SysRoleSysEmployee
			}
			emp := &accounttype.Employee{
				CompanyID: cmp.ID,
				UserID:    u.ID,
				Uin:       ui.ID,
				SysRole:   role,
			}
			if err := tx.Create(emp).Error; err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		panic(err)
	}
}

func TestGetCompanyQuota(t *testing.T) {
	quota := &accounttype.ResourceQuota{
		DiskQuota:     100 * forest.GB,
		QAQuota:       3000,
		AgentQuota:    300,
		EmployeeQuota: 300,
	}
	bytes, err := json.Marshal(quota)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(bytes))
}

func InitDB() {
	if err := dbtools.InitMultiDBConn(map[string]string{
		"account": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"chat":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	}); err != nil {
		panic(err)
	}
}

func TestCreateCompany(t *testing.T) {
	t.Skip()
	if err := dbutil.Account().Create(&accounttype.Company{
		Name: "test-company-0-0-1",
		Quota: &accounttype.ResourceQuota{
			DiskQuota:     100 * forest.GB,
			QAQuota:       3000,
			AgentQuota:    300,
			EmployeeQuota: 300,
		},
	}).Error; err != nil {
		panic(err)
	}

}

func TestGetEmployees(t *testing.T) {
	InitDB()
	emps, err := employee.GetEmployeeByCompanyID(72)
	if err != nil {
		panic(err)
	}
	fmt.Println("len: ", len(emps))
	for _, emp := range emps {
		fmt.Printf("%+v\n", emp)
	}
}

func TestGetAgentCount(t *testing.T) {
	InitDB()
	ags, err := chatagent.GetALLAgentsByCompanyID(context.Background(), 72)
	if err != nil {
		panic(err)
	}
	fmt.Println("len: ", len(ags))
	for _, ag := range ags {
		fmt.Printf("%+v\n", ag)
	}
}

type Foo struct {
	Bar *string
}

func TestFoo(t *testing.T) {
	f := &Foo{}

	if *(*f).Bar == "" {
		println("null")
	}

	fmt.Printf("%+v", f)
}

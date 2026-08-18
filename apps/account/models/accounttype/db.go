package accounttype

import (
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"github.com/ygpkg/yg-go/storage"
)

const (
	// TableNameUser 用户登录表
	TableNameUser = "user"
	// TableNameUserIdentification 用户标识
	TableNameUserIdentification = "user_identification"
	// TableNameIndividual 个人用户表
	TableNameIndividual = "individual"
	// TableNameCompany 公司用户表
	TableNameCompany             = "company"
	TableNameCompanyInvitation   = "company_invitation"
	TableNameCompanyUpgradeApply = "company_upgrade_apply"

	TableNameEmployee = "account_employee"
	// TableNameEmployeeThirdBinding = "account_emp_third_binding"
	TableNamePosition             = "account_position"
	TableNamePrivilege            = "account_privilege"
	TableNameRelPositionPrivilege = "account_rel_position_privilege"
	TableNameRelEmployeePosition  = "account_rel_employee_position"

	TableNameWechatBinding = "account_wechat_binding"

	TableNameStudent      = "account_student"
	TableNameClass        = "account_class"
	TableNameClassStudent = "account_class_student"
	TableNameClassTeacher = "account_class_teacher"

	// APIkey
	TableNameAPIKey           = "account_api_key"
	TableNameAPIKeyPrivilege  = "account_api_key_privilege"
	TableNameAPIPrivilege     = "account_api_privilege"
	TableNameAPIService       = "account_api_service"
	TableNameAPIAuthorization = "account_api_authorization"

	// Account ExternalBinding
	TableNameUserExternalBinding          = "account_external_binding"
	TableNameAccountDepartment            = "account_department"
	TableNameAccountRelEmployeeDepartment = "account_rel_employee_department"
)

// InitModel init db tables
func InitDB() error {
	err := dbtools.InitModel(dbutil.Account(),
		&User{},
		&UserIdentification{},
		&Individual{},
		&Company{},
		&CompanyInvitation{},
		&CompanyUpgradeApply{},

		&Employee{},
		&Position{},
		&Privilege{},
		&RelEmployeePosition{},
		&RelPositionPrivilege{},
		&WechatBinding{},
		&AccountDepartment{},
		&AccountRelEmployeeDepartment{},

		//
		&APIKey{},
		&APIKeyPrivilege{},
		&APIPrivilege{},
		&APIService{},
		&APIAuthorization{},

		&ExternalBinding{},
	)
	if err != nil {
		return err
	}

	{
		// db_pre
		if err := presetDatabase(); err != nil {
			return err
		}
	}
	return nil
}

// 数据库初始化准备 生成职位权限
func presetDatabase() error {
	return nil
}

func InitCoreDB() error {
	err := dbtools.InitModel(dbutil.Core(),
		&storage.FileInfo{},
	)
	if err != nil {
		return err
	}

	{ // db_pre
		if err := presetDatabase(); err != nil {
			return err
		}
	}
	return nil
}

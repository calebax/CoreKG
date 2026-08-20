package svcorganize

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"mime/multipart"
	stdurl "net/url"
	"path/filepath"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/insmtx/corekg/apps/account/internal/dto/dtoorganize"
	"github.com/insmtx/corekg/apps/account/models/account"
	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/employee"
	"github.com/insmtx/corekg/apps/account/models/user"
	"github.com/insmtx/corekg/apps/kecore/services/svccoze"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/pkgs/platform/company"
	"github.com/insmtx/corekg/pkgs/utils"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/insmtx/corekg/version"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/config"
	"github.com/ygpkg/yg-go/i18n"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/random"
	"github.com/ygpkg/yg-go/settings"
	"github.com/ygpkg/yg-go/storage"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

func CreateDepartment(ctx *gin.Context, req *dtoorganize.CreateDepartmentRequest) (res *dtoorganize.CreateDepartmentResponse, err error) {
	res = &dtoorganize.CreateDepartmentResponse{}
	companyID := runtime.CompanyID(ctx)
	//check parent exist
	if req.Request.ParentId > 0 {
		if cnt, err := account.NewAccountDepartmentDao().CountByCond(ctx, &account.DepartmentCond{
			ID:        req.Request.ParentId,
			CompanyID: companyID,
		}); err != nil {
			logs.ErrorContextf(ctx, "check parent department exist faild %v", err)
			return res, err
		} else if cnt <= 0 {
			res.Code = errcode.ErrCode_BadRequest
			res.Message = i18n.T(runtime.GetLanguage(ctx), "account_department_not_exist")
			return res, err
		}
	}

	//check name exist
	if cnt, err := account.NewAccountDepartmentDao().CountByCond(ctx, &account.DepartmentCond{
		Name:      req.Request.Name,
		CompanyID: companyID,
	}); err != nil {
		logs.ErrorContextf(ctx, "check department name exist faild %v", err)
		return res, err
	} else if cnt > 0 {
		res.Code = errcode.ErrCode_BadRequest
		res.Message = i18n.T(runtime.GetLanguage(ctx), "account_department_name_exist")
		return res, err
	}

	//get the closest department sort
	dep, err := account.NewAccountDepartmentDao().GetByCond(ctx, &account.DepartmentCond{
		BaseCond: account.BaseCond{
			Limit:   1,
			OrderBy: []string{"sort desc"}},
		ParentID: req.Request.ParentId})
	if err != nil {
		logs.ErrorContextf(ctx, "get the closest department failed %v", err)
		return res, err
	}

	dept := &accounttype.AccountDepartment{
		Name:      req.Request.Name,
		ParentID:  req.Request.ParentId,
		Sort:      dep.Sort + account.SortGap,
		CompanyID: companyID,
	}

	if err = account.NewAccountDepartmentDao().Insert(ctx, dept); err != nil {
		logs.ErrorContextf(ctx, "insert department failed %v", err)
		return res, err
	}

	res.Response.AccountDepartment = dept

	return res, nil
}

func DeleteDepartment(ctx *gin.Context, req *dtoorganize.DeleteDepartmentRequest) (res *dtoorganize.DeleteDepartmentResponse, err error) {
	res = &dtoorganize.DeleteDepartmentResponse{}
	companyID := runtime.CompanyID(ctx)
	//check if department has any employee
	if emps, err := account.NewAccountRelEmployeeDepartmentDao().CountByCond(ctx, &account.RelEmployeeDepartmentCond{DepartmentID: req.Request.ID}); err != nil {
		logs.ErrorContextf(ctx, "check department employee exist faild %v", err)
		return res, err
	} else if emps > 0 {
		res.Code = errcode.ErrCode_BadRequest
		res.Message = i18n.T(runtime.GetLanguage(ctx), "account_department_exist_employee")
		return res, err
	}

	//check department exist
	if req.Request.ID > 0 {
		if cnt, err := account.NewAccountDepartmentDao().CountByCond(ctx, &account.DepartmentCond{
			ID:        req.Request.ID,
			CompanyID: companyID,
		}); err != nil {
			logs.ErrorContextf(ctx, "check department exist faild %v", err)
			return res, err
		} else if cnt <= 0 {
			res.Code = errcode.ErrCode_BadRequest
			res.Message = i18n.T(runtime.GetLanguage(ctx), "account_department_not_exist")
			return res, err
		}
	}

	//delete department
	if err = account.NewAccountDepartmentDao().Delete(ctx, req.Request.ID); err != nil {
		logs.ErrorContextf(ctx, "delete department(id:%v) failed %v", req.Request.ID, err)
		return res, err
	}

	return res, nil
}

func RenameDepartment(ctx *gin.Context, req *dtoorganize.RenameDepartmentRequest) (res *dtoorganize.RenameDepartmentResponse, err error) {
	res = &dtoorganize.RenameDepartmentResponse{}
	//check department exit
	dep, err := account.NewAccountDepartmentDao().GetByID(ctx, req.Request.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "check department exist faild %v", err)
		return nil, err
	}

	if dep.Name == req.Request.Name {
		logs.WarnContextf(ctx, "rename department[id:%v,name:%v] but do nothing", req.Request.ID, req.Request.Name)
		return res, nil
	}

	//check if target name exist
	if cnt, err := account.NewAccountDepartmentDao().CountByCond(ctx, &account.DepartmentCond{Name: req.Request.Name}); err != nil {
		logs.ErrorContextf(ctx, "check department name exist faild %v", err)
		return nil, err
	} else if cnt > 0 {
		res.Code = errcode.ErrCode_BadRequest
		res.Message = i18n.T(runtime.GetLanguage(ctx), "account_department_name_exist")
		return nil, err
	}
	dep.Name = req.Request.Name

	if err = account.NewAccountDepartmentDao().UpdateByID(ctx, dep.ID, dep); err != nil {
		logs.ErrorContextf(ctx, "update department(id:%v) failed %v", dep.ID, err)
		return nil, err
	}
	res.Response.AccountDepartment = dep

	return res, nil
}

func MoveDepartment(ctx *gin.Context, req *dtoorganize.MoveDepartmentRequest) (res *dtoorganize.MoveDepartmentResponse, err error) {
	res = &dtoorganize.MoveDepartmentResponse{}

	// 1. 在单个事务中执行所有操作，保证数据一致性
	if err = dbutil.Account().Transaction(func(tx *gorm.DB) error {
		var targetDept, preDept, postDept *accounttype.AccountDepartment
		var preSort, postSort uint

		// 2. 获取被移动的部门信息
		targetDept, err = account.NewAccountDepartmentDao().GetByID(ctx, req.Request.DepartmentId)
		if err != nil {
			logs.ErrorContextf(ctx, "check department(id:%v) exist faild %v", req.Request.DepartmentId, err)
			return fmt.Errorf("target department with id %d not found: %w", req.Request.PostID, err)
		}

		// 3. 根据 PreID 和 PostID 计算排序边界
		// --------------------------------------------------------------------
		// Case 1: 移动到顶部 (PreID 为 0)
		if req.Request.PreID == 0 {
			preSort = 0

			postDept, err = account.NewAccountDepartmentDao().GetByID(ctx, req.Request.PostID)
			if err != nil {
				logs.ErrorContextf(ctx, "check post department(id:%v) exist faild %v", req.Request.PostID, err)
				return fmt.Errorf("post department with id %d not found: %w", req.Request.PostID, err)
			}

			if postDept.ParentID != targetDept.ParentID {
				logs.ErrorContextf(ctx, "departments must be in the same level to be sorted[post's parent_id:%v|targ's parent_id:%v]", postDept.ParentID, targetDept.ParentID)
				return fmt.Errorf("departments must be in the same level to be sorted")
			}
			postSort = postDept.Sort
		} else if req.Request.PostID == 0 {
			// Case 2: 移动到底部 (PostID 为 0)
			preDept, err = account.NewAccountDepartmentDao().GetByID(ctx, req.Request.PreID)
			if err != nil {
				logs.ErrorContextf(ctx, "pre department with id %d not found: %w", req.Request.PreID, err)
				return fmt.Errorf("pre department with id %d not found: %w", req.Request.PreID, err)
			}

			if preDept.ParentID != targetDept.ParentID {
				logs.ErrorContextf(ctx, "departments must be in the same level to be sorted[pre's parent_id:%v|targ's parent_id:%v]", preDept.ParentID, targetDept.ParentID)
				return fmt.Errorf("departments must be in the same level to be sorted")
			}
			preSort = preDept.Sort
			// 设定一个足够大的值作为末尾
			postSort = preDept.Sort + account.SortGap*2
		} else {
			// Case 3: 移动到两个部门之间
			preDept, err = account.NewAccountDepartmentDao().GetByID(ctx, req.Request.PreID)
			if err != nil {
				logs.ErrorContextf(ctx, "pre department with id %d not found: %w", req.Request.PreID, err)
				return fmt.Errorf("pre department with id %d not found: %w", req.Request.PreID, err)
			}

			postDept, err = account.NewAccountDepartmentDao().GetByID(ctx, req.Request.PostID)
			if err != nil {
				logs.ErrorContextf(ctx, "check post department(id:%v) exist faild %v", req.Request.PostID, err)
				return fmt.Errorf("post department with id %d not found: %w", req.Request.PostID, err)
			}
			if preDept.ParentID != targetDept.ParentID || postDept.ParentID != targetDept.ParentID {
				logs.ErrorContextf(ctx, "departments must be in the same level to be sorted[pre_parent_id:%v|targ_parent_id:%v|post_parent_id:%v]", preDept.ParentID, targetDept.ParentID, postDept.ParentID)
				return fmt.Errorf("departments must be in the same level to be sorted")
			}
			preSort = preDept.Sort
			postSort = postDept.Sort
		}
		// --------------------------------------------------------------------

		// 4. 计算新的排序值
		newSort := (preSort + postSort) / 2

		// 5. 检查是否需要重排序
		// 当计算出的 newSort 与 preSort 相等时，意味着 pre 和 post 之间没有足够的整数空间
		// 例如 preSort=1000, postSort=1001, newSort=(1000+1001)/2 = 1000
		if newSort == preSort {
			logs.DebugContextf(ctx, "Sort gap exhausted between %d and %d. Rebalancing parent ID %d...\n", preSort, postSort, targetDept.ParentID)

			// 调用重排序方法
			if err := account.NewAccountDepartmentDao().RebalanceSiblings(ctx, tx, targetDept.ParentID, req.Request.DepartmentId, req.Request.PreID, req.Request.PostID); err != nil {
				logs.ErrorContextf(ctx, "rebalance department(id:%v) failed %v", targetDept.ParentID, err)
				return fmt.Errorf("failed to rebalance: %w", err)
			}
			return nil
		}

		// 6. 如果不需要重排，直接更新目标部门的排序值
		logs.DebugContextf(ctx, "Moving department %d to new sort value: %d\n", targetDept.ID, newSort)
		return tx.Model(&targetDept).Update("sort", newSort).Error
	}); err != nil {
		logs.ErrorContextf(ctx, "move department failed: %v", err)
		return nil, err
	}
	return res, nil
}

func GetDepartmentTree(ctx *gin.Context, req *dtoorganize.GetDepartmentTreeRequest) (res *dtoorganize.GetDepartmentTreeResponse, err error) {
	res = &dtoorganize.GetDepartmentTreeResponse{}
	companyID := runtime.CompanyID(ctx)
	// get company's all departments
	depts, err := account.NewAccountDepartmentDao().GetListByCond(ctx, &account.DepartmentCond{
		CompanyID: companyID,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "get departments(companyID:%v) failed: %v", companyID, err)
		return nil, fmt.Errorf("get departments(companyID:%v) failed: %w", companyID, err)
	}
	res.Response.Departments = depts

	if req.Request.IncludeEmployee {
		emps, err := account.NewAccountRelEmployeeDepartmentDao().GetCompanyEmployInfo(ctx, companyID)
		if err != nil {
			logs.ErrorContextf(ctx, "get employees(companyID:%v) failed: %v", companyID, err)
			return nil, fmt.Errorf("get employees(companyID:%v) failed: %w", companyID, err)
		}
		empMaps := make(map[uint]dtoorganize.EmployeeInfo, len(emps))
		for _, v := range emps {
			empMaps[v.Uin] = v
		}
		relEmps, err := account.NewAccountRelEmployeeDepartmentDao().GetListByCond(ctx, &account.RelEmployeeDepartmentCond{CompanyID: companyID})
		if err != nil {
			logs.ErrorContextf(ctx, "get employeesRels(companyID:%v) failed: %v", companyID, err)
			return nil, fmt.Errorf("get employeesRels(companyID:%v) failed: %w", companyID, err)
		}

		for _, v := range relEmps {
			tempEmp := empMaps[v.Uin]
			switch v.IsPrimary {
			case 1:
				tempEmp.DepartmentIDs = append([]uint{v.DepartmentID}, tempEmp.DepartmentIDs...)
			case -1:
				tempEmp.DepartmentIDs = append(tempEmp.DepartmentIDs, v.DepartmentID)
			}
			empMaps[v.Uin] = tempEmp
		}
		for _, v := range empMaps {
			res.Response.Employees = append(res.Response.Employees, v)
		}
	}
	return res, nil
}

const DefaultPasswdLen = 8

func CreateEmployee(ctx *gin.Context, info *dtoorganize.EmployeeInfo) (*apiobj.BaseResponse, error) {
	//check department list valid
	var (
		res       = &apiobj.BaseResponse{}
		companyID = runtime.CompanyID(ctx)
		pwdStr    string
		newUser   bool
		u         *accounttype.User
		ui        *accounttype.UserIdentification
	)
	defer func() {
		_ = svccoze.SpaceSync(ctx)
	}()
	deps, err := account.NewAccountDepartmentDao().GetListByCond(ctx, &account.DepartmentCond{
		IDs:       info.DepartmentIDs,
		CompanyID: companyID,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "get departments(companyID:%v) failed: %v", info.DepartmentIDs, err)
		return nil, fmt.Errorf("get departments(companyID:%v) failed: %w", info.DepartmentIDs, err)
	} else if len(info.DepartmentIDs) != len(deps) {
		logs.ErrorContextf(ctx, "department list exist invalid id [desire:%v] [actually:%v]", info.DepartmentIDs, deps)
		return nil, fmt.Errorf("department list exist invalid id [desire:%v] [actually:%v]", info.DepartmentIDs, deps)
	}

	// check name || phone || email exist
	var phone *string
	if len(info.Phone) == 0 {
		phone = nil
	} else {
		phone = &info.Phone
	}

	//*  ===============================================
	//  check user's email or phone exist
	u, err = user.GetUserByPhoneAndEmail(ctx, phone, &info.Email)
	if err != nil {
		// ! unexpected error
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			logs.ErrorContextf(ctx, "GetUserByPhoneAndEmail(phone:%v,email:%v)", utils.PtrValue(phone), utils.PtrValue(&info.Email))
			return res, err
		}
		// expected
		// get user by phone and check user's email valid
		u, err = user.GetUserByPhone(*phone)
		if err != nil {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				logs.ErrorContextf(ctx, "GetUserByPhone(phone:%v) failed: %v", *phone, err)
				res.Code = errcode.ErrCode_InternalError
				res.Message = "account_get_user_by_phone_failed"
				return res, err
			}

			newUser = true

			//? target user no existed, create a new user
			pwd, err := bcrypt.GenerateFromPassword([]byte(info.Phone[len(info.Phone)-DefaultPasswdLen:]), bcrypt.DefaultCost)
			if err != nil {
				logs.ErrorContextf(ctx, "bcrypt.GenerateFromPassword:uin[%v] desire to create pwd failed, %+v", runtime.Uin(ctx), err)
				res.Code = errcode.ErrCode_InternalError
				res.Message = "account_generate_password_failed" // 生成密码失败
				return res, nil
			}
			pwdStr = string(pwd)
			u = &accounttype.User{
				Identify:        uuid.NewString(),
				Name:            info.Name,
				Email:           &info.Email,
				Phone:           phone,
				Password:        &pwdStr,
				AvatarURL:       "/images/default-user-avatar.svg",
				PasswordChanged: -1,
			}
		}
	}

	if !newUser {
		//? when accepted email is not matching stored email, return error
		if len(info.Email) > 0 &&
			(u.Email == nil || *u.Email == "" || (*u.Email != "" && *u.Email != info.Email)) {
			logs.ErrorContextf(ctx, "User email is invalid, phone:%v, email:%v but stored email is %v", *phone, info.Email, utils.PtrValue(u.Email))
			res.Code = errcode.ErrCode_BadRequest
			res.Message = "account_email_not_match"
			return res, fmt.Errorf("user email is invalid, phone:%v, email:%v but stored email is %v", *phone, info.Email, utils.PtrValue(u.Email))
		}

		//? if user found from GetUserByPhoneAndEmail's company id == companyID
		_, err = user.GetCompanyUinByUserID(ctx, u.ID, companyID)
		if err == nil || !errors.Is(err, gorm.ErrRecordNotFound) {
			if !errors.Is(err, gorm.ErrRecordNotFound) {
				//! abnormal
				logs.ErrorContextf(ctx, "GetCompanyUinByUserID faild err: :%v", err)
				res.Code = errcode.ErrCode_InternalError
				res.Message = "account_get_company_uin_by_user_id_failed"
				return res, err
			}
			//! already exist this user in local company
			logs.ErrorContextf(ctx, "Already exist user(phone:%v,email:%v) in local-company", utils.PtrValue(phone), utils.PtrValue(&info.Email))
			res.Code = errcode.ErrCode_BadRequest
			res.Message = "account_user_info_already_exists_local"
			return res, nil
		}
	}

	// * expected
	if user.ExistUser(ctx, phone, &info.Email, companyID) {
		// ? user has already existed in local company
		logs.ErrorContextf(ctx, "ExistUser check faild user's phone(%v) or email(%v) has already exist", utils.PtrValue(phone), utils.PtrValue(&info.Email))
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "account_user_info_already_exists"
		return res, nil
	}

	// ? user exist but not existed in local company
	ui = &accounttype.UserIdentification{
		UserID:      u.ID,
		SubjectType: accounttype.SubjectTypeCompany,
		SubjectID:   companyID,
		UinStatus:   accounttype.UinStatusNormal,
		Issuer:      global.IssuerYYGU,
		Name:        info.Name,
	}

	//* ===============================================

	var (
		emp = &accounttype.Employee{
			CompanyID: companyID,
			SysRole:   info.SysRole,
		}

		relEmps []accounttype.AccountRelEmployeeDepartment
	)
	if len(info.Email) <= 0 {
		u.Email = nil
	}

	if err = dbutil.Account().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if newUser {
			if err := tx.Create(u).Error; err != nil {
				return err
			}
		}
		ui.UserID = u.ID
		//create uin
		if err := tx.Create(ui).Error; err != nil {
			return err
		}
		emp.Uin = ui.ID
		emp.UserID = u.ID
		//create employee
		if err := tx.Create(emp).Error; err != nil {
			return err
		}

		for i, v := range info.DepartmentIDs {
			r := accounttype.AccountRelEmployeeDepartment{
				Uin:          ui.ID,
				DepartmentID: v,
				EmployeeID:   emp.ID,
				CompanyID:    companyID,
				IsPrimary:    -1,
			}
			if i == 0 {
				r.IsPrimary = 1
			}
			relEmps = append(relEmps, r)
		}

		//批量创建部门规则
		return dbutil.Account().CreateInBatches(relEmps, len(relEmps)).Error
	}); err != nil {
		logs.ErrorContextf(ctx, "CreateEmployee failed, %+v", err)
		return nil, err
	}
	info.Uin = ui.ID
	info.EmployeeID = emp.ID
	info.CreatedAt = u.CreatedAt
	return res, nil
}

const DefaultPrivatePasswordPrefix = "pwd"

func CreateEmployeePrivate(ctx *gin.Context, info *dtoorganize.EmployeeInfo) (*apiobj.BaseResponse, error) {
	//check department list valid
	var (
		res       = &apiobj.BaseResponse{}
		companyID = runtime.CompanyID(ctx)
	)
	defer func() {
		_ = svccoze.SpaceSync(ctx)
	}()
	deps, err := account.NewAccountDepartmentDao().GetListByCond(ctx, &account.DepartmentCond{
		IDs:       info.DepartmentIDs,
		CompanyID: companyID,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "get departments(companyID:%v) failed: %v", info.DepartmentIDs, err)
		return nil, fmt.Errorf("get departments(companyID:%v) failed: %w", info.DepartmentIDs, err)
	} else if len(info.DepartmentIDs) != len(deps) {
		logs.ErrorContextf(ctx, "department list exist invalid id [desire:%v] [actually:%v]", info.DepartmentIDs, deps)
		return nil, fmt.Errorf("department list exist invalid id [desire:%v] [actually:%v]", info.DepartmentIDs, deps)
	}

	// check name || phone || email exist
	var phone *string
	if len(info.Phone) == 0 {
		phone = nil
	} else {
		phone = &info.Phone
	}
	if exist := user.ExistUser(ctx, phone, &info.Email, companyID); exist {
		logs.WarnContextf(ctx, "User info has already exist about name[%v] phone[%v] email[%v] in company[id:%v]", info.Name, info.Phone, info.Email, companyID)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "account_user_info_already_exists" // 用户名,手机号或邮箱已存在
		return res, nil
	}

	// 检查用户标识是否存在
	idt := random.String(6)
	exi, err := user.ExistsUserByIIdentify(ctx, idt)
	if err != nil {
		logs.ErrorContextf(ctx, "ExistsUserByIIdentify: exists user failed, %+v", err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "account_check_user_identify_failed")) // 检查用户标识失败
		return res, nil
	}
	if exi {
		idt = fmt.Sprintf("%s%d", idt, rand.Intn(10))
	}

	//default password = pwd + email
	passwd := []byte(DefaultPrivatePasswordPrefix + info.Email)

	pwd, err := bcrypt.GenerateFromPassword(passwd, bcrypt.DefaultCost)
	if err != nil {
		logs.ErrorContextf(ctx, "bcrypt.GenerateFromPassword:uin[%v] desire to create pwd failed, %+v", runtime.Uin(ctx), err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "account_generate_password_failed" // 生成密码失败
		return res, nil
	}

	pwdStr := string(pwd)
	//create user
	var (
		u = &accounttype.User{
			Identify:        idt,
			Name:            info.Name,
			Email:           &info.Email,
			Phone:           phone,
			Password:        &pwdStr,
			AvatarURL:       "/images/default-user-avatar.svg",
			PasswordChanged: -1,
		}
		ui = &accounttype.UserIdentification{
			Name:        u.Name,
			SubjectType: accounttype.SubjectTypeCompany,
			SubjectID:   companyID,
			UinStatus:   accounttype.UinStatusNormal,
			Issuer:      global.IssuerYYGU,
		}
		emp = &accounttype.Employee{
			CompanyID: companyID,
			SysRole:   info.SysRole,
		}

		relEmps []accounttype.AccountRelEmployeeDepartment
	)

	if len(info.Phone) <= 0 {
		u.Phone = nil
	}

	if err = dbutil.Account().Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(u).Error; err != nil {
			return err
		}
		ui.UserID = u.ID
		//create uin
		if err := tx.Create(ui).Error; err != nil {
			return err
		}
		emp.Uin = ui.ID
		emp.UserID = u.ID
		//create employee
		if err := tx.Create(emp).Error; err != nil {
			return err
		}

		for i, v := range info.DepartmentIDs {
			r := accounttype.AccountRelEmployeeDepartment{
				Uin:          ui.ID,
				DepartmentID: v,
				EmployeeID:   emp.ID,
				CompanyID:    companyID,
				IsPrimary:    -1,
			}
			if i == 0 {
				r.IsPrimary = 1
			}
			relEmps = append(relEmps, r)
		}

		//批量创建部门规则
		return dbutil.Account().CreateInBatches(relEmps, len(relEmps)).Error
	}); err != nil {
		logs.ErrorContextf(ctx, "CreateEmployee failed, %+v", err)
		return nil, err
	}
	info.Uin = ui.ID
	info.EmployeeID = emp.ID
	info.CreatedAt = u.CreatedAt
	return res, nil
}

func EditEmployee(ctx context.Context, info dtoorganize.EmployeeInfo, companyID uint) (*apiobj.BaseResponse, *dtoorganize.EmployeeInfo, error) {
	var (
		res                                                                    = &apiobj.BaseResponse{}
		uinUpdate, userUpdate, employeeUpdate, departmentUpdate, primaryUpdate bool
		toDelDepIDs, toAddDepIDs                                               []uint
		desirePrimDepID                                                        uint
	)

	u, err := user.GetUserByUin(ctx, info.Uin)
	if err != nil {
		logs.ErrorContextf(ctx, "user.GetUserByUin(uin:%v) failed, %+v", info.Uin, err)
		return res, nil, err
	}

	ui, err := user.GetUserIdentificationByUIN(ctx, info.Uin)
	if err != nil {
		logs.ErrorContextf(ctx, "user.GetUserIdentificationByUIN(uin:%v) failed, %+v", info.Uin, err)
		return res, nil, err
	}

	emp, err := employee.GetEmployeeByID(ctx, info.EmployeeID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetEmployeeByID(id:%v) failed :%v", info.EmployeeID, err)
		return res, nil, err
	}

	if len(info.Name) > 0 {
		//if user.ExistUinName(ctx, info.Name, ui.ID, companyID) {
		//	logs.ErrorContextf(ctx, "user.ExistUinName(name:%v) failed, %+v", info.Name, ui)
		//	res.Code = errcode.ErrCode_BadRequest
		//	res.Message = "account_name_already_exists"
		//	return res, nil, nil
		//}
		ui.Name = info.Name
		uinUpdate = true
	}

	if u.Phone == nil || *u.Phone != info.Phone {
		if len(info.Phone) > 0 {
			if existPhone, err := user.ExistPhone(ctx, u.ID, info.Phone); err != nil {
				logs.ErrorContextf(ctx, "user.ExistPhone(phone:%v) failed, %+v", info.Phone, err)
				return res, nil, err
			} else if existPhone {
				logs.WarnContextf(ctx, "phone(%v) has already exist in company(id:%v)", utils.PtrValue(&info.Phone), emp.CompanyID)
				res.Code = errcode.ErrCode_BadRequest
				res.Message = "account_phone_already_exists"
				return res, nil, nil
			}
		}
		userUpdate = true
		u.Phone = &info.Phone
	}

	if u.Email == nil || *u.Email != info.Email {
		if len(info.Email) > 0 {
			if exist, err := user.ExistEmail(ctx, u.ID, info.Email); err != nil {
				logs.ErrorContextf(ctx, "user.ExistEmail(email:%v) failed, %+v", info.Email, err)
				return res, nil, err
			} else if exist {
				logs.WarnContextf(ctx, "email(%v) has already exist in company(id:%v)", utils.PtrValue(&info.Email), emp.CompanyID)
				res.Code = errcode.ErrCode_BadRequest
				res.Message = "account_email_already_exist"
				return res, nil, nil
			}
		}
		userUpdate = true
		u.Email = &info.Email
	}

	if len(info.SysRole) > 0 && emp.SysRole != info.SysRole {
		employeeUpdate = true
		emp.SysRole = info.SysRole
	}

	if len(info.DepartmentIDs) > 0 {
		//check department ids valid
		deps, err := account.NewAccountDepartmentDao().GetListByCond(ctx, &account.DepartmentCond{
			IDs:       info.DepartmentIDs,
			CompanyID: companyID,
		})
		if err != nil {
			logs.ErrorContextf(ctx, "get departments(companyID:%v) failed: %v", info.DepartmentIDs, err)
			return res, nil, fmt.Errorf("get departments(companyID:%v) failed: %w", info.DepartmentIDs, err)
		} else if len(info.DepartmentIDs) != len(deps) {
			logs.ErrorContextf(ctx, "department list exist invalid id [desire:%v] [actually:%v]", info.DepartmentIDs, deps)
			return res, nil, fmt.Errorf("department list exist invalid id [desire:%v] [actually:%v]", info.DepartmentIDs, deps)
		}

		desirePrimDepID = info.DepartmentIDs[0]

		relEmpDeps, err := account.NewAccountRelEmployeeDepartmentDao().GetListByCond(ctx, &account.RelEmployeeDepartmentCond{
			Uin:       info.Uin,
			CompanyID: emp.CompanyID,
		})
		if err != nil {
			logs.ErrorContextf(ctx, "GetRelEmpDeptListByCond(uin:%v) failed, %+v", info.Uin, err)
			return res, nil, err
		}
		var (
			OldDeptIDs []uint
			primDepID  uint
		)

		for _, r := range relEmpDeps {
			if r.IsPrimary == 1 {
				primDepID = r.DepartmentID
			}
			OldDeptIDs = append(OldDeptIDs, r.DepartmentID)
		}
		mockDeptIDs := info.DepartmentIDs
		slices.Sort(OldDeptIDs)
		slices.Sort(mockDeptIDs)

		if len(mockDeptIDs) == len(OldDeptIDs) && primDepID == desirePrimDepID {
			for i := range OldDeptIDs {
				if OldDeptIDs[i] != mockDeptIDs[i] {
					departmentUpdate = true
					break
				}
			}
		} else {
			departmentUpdate = true
		}

		oldRelMap := relEmpDeps.ToMap()
		NewRelMap := make(map[uint]struct{}, 0)
		for _, v := range info.DepartmentIDs {
			NewRelMap[v] = struct{}{}
		}

		//Get toDel
		for depID, dep := range oldRelMap {
			if _, exists := NewRelMap[depID]; !exists {
				if dep.IsPrimary == 1 {
					primaryUpdate = true
				}
				toDelDepIDs = append(toDelDepIDs, depID)
			}
		}

		// Get toAdd
		for depID := range NewRelMap {
			if _, exists := oldRelMap[depID]; !exists {
				toAddDepIDs = append(toAddDepIDs, depID)
			}
		}
	}
	if err = dbutil.Account().Transaction(func(tx *gorm.DB) error {
		if uinUpdate {
			if err := tx.Save(ui).Error; err != nil {
				logs.ErrorContextf(ctx, "tx.Save(uin:%v) failed, %+v", ui.ID, err)
				return err
			}
		}
		if userUpdate {
			if err := tx.Save(u).Error; err != nil {
				logs.ErrorContextf(ctx, "tx.Save(user_id:%v) failed, %+v", u.ID, err)
				return err
			}
		}

		if departmentUpdate {
			//to del
			if err := tx.Table(accounttype.TableNameAccountRelEmployeeDepartment).
				Where("deleted_at IS NULL").
				Where("id in (?)", toDelDepIDs).
				Where("uin  = ?", info.Uin).
				Where("employee_id = ?", info.EmployeeID).
				Delete(&accounttype.AccountRelEmployeeDepartment{}).
				Error; err != nil {
				logs.ErrorContextf(ctx, "tx.Delete(department_ids:%v) failed, %+v", toDelDepIDs, err)
				return err
			}

			var toAddDeps []accounttype.AccountRelEmployeeDepartment
			for _, id := range toAddDepIDs {
				dep := accounttype.AccountRelEmployeeDepartment{
					Uin:          info.Uin,
					EmployeeID:   emp.ID,
					CompanyID:    companyID,
					DepartmentID: id,
					IsPrimary:    -1,
				}
				if primaryUpdate && id == desirePrimDepID {
					dep.IsPrimary = 1
				}

				toAddDeps = append(toAddDeps, dep)
			}

			if err := tx.CreateInBatches(&toAddDeps, len(toAddDeps)).Error; err != nil {
				logs.ErrorContextf(ctx, "tx.CreateInBatches(%v) failed, %+v", toAddDeps, err)
				return err
			}
		}

		if employeeUpdate {
			if err := tx.Save(emp).Error; err != nil {
				logs.ErrorContextf(ctx, "tx.Save(emp:%v) failed, %+v", emp, err)
				return err
			}
		}

		return nil
	}); err != nil {
		logs.ErrorContextf(ctx, "EditDepartmentEmployeeDepartment failed, %+v", err)
		return res, nil, err
	}

	empInfo := &dtoorganize.EmployeeInfo{
		DepartmentIDs: info.DepartmentIDs,
		Uin:           ui.ID,
		EmployeeID:    emp.ID,
		SysRole:       emp.SysRole,
		Name:          ui.Name,
	}
	if u.Email != nil {
		email := *u.Email
		logs.DebugContextf(ctx, "user(id:%v)'s email :%v", u.ID, email)
		empInfo.Email = *u.Email
	}
	if u.Phone != nil {
		logs.DebugContextf(ctx, "user(id:%v)'s phone :%v", u.ID, *u.Phone)
		empInfo.Phone = *u.Phone
	}
	return res, empInfo, nil
}

func CreateDepartmentEmployee(ctx *gin.Context, req *dtoorganize.CreateDepartmentEmployeeRequest) (res *dtoorganize.CreateDepartmentEmployeeResponse, err error) {
	res = &dtoorganize.CreateDepartmentEmployeeResponse{}
	resp, err := CreateEmployee(ctx, &req.Request.EmployeeInfo)
	if err != nil {
		logs.ErrorContextf(ctx, "CreateDepartmentEmployee: CreateEmployee failed, %+v", err)
	}
	if resp != nil {
		res.BaseResponse = *resp
	}
	res.Response.EmployeeInfo = &req.Request.EmployeeInfo

	return res, nil
}

func EditDepartmentEmployee(ctx *gin.Context, req *dtoorganize.EditDepartmentEmployeeRequest) (res *dtoorganize.EditDepartmentEmployeeResponse, err error) {
	res = &dtoorganize.EditDepartmentEmployeeResponse{}
	resp, emp, err := EditEmployee(ctx, req.Request.EmployeeInfo, runtime.CompanyID(ctx))
	if err != nil {
		logs.ErrorContextf(ctx, "EditDepartmentEmployee(%+v) failed, %+v", req, err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "account_edit_employee_failed"
		return nil, err
	}
	if resp != nil {
		res.BaseResponse = *resp
	}

	if emp != nil {
		res.Response.EmployeeInfo = emp
	}
	return res, nil
}

func EditCompanyInfo(ctx *gin.Context, req *dtoorganize.EditCompanyInfoRequest) (res *dtoorganize.EditCompanyInfoResponse, err error) {
	res = &dtoorganize.EditCompanyInfoResponse{}
	companyID := runtime.CompanyID(ctx)
	defer func() {
		_ = svccoze.SpaceSync(ctx)
	}()
	var count int64
	//check company name exist
	if err = dbutil.Account().Table(accounttype.TableNameCompany).
		Where("deleted_at IS NULL").
		Where("id != ?", companyID).
		Where("name = ?", req.Request.Name).
		Count(&count).
		Error; err != nil {
		logs.ErrorContextf(ctx, "check name exist failed, %v", err)
		return nil, err
	}
	if count > 0 {
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "account_company_name_exist"
		return res, nil
	}

	if len(req.Request.Name) > 0 {
		if err := dbutil.Account().Table(accounttype.TableNameCompany).
			Where("id = ?", companyID).
			Update("name", req.Request.Name).
			Error; err != nil {
			logs.ErrorContextf(ctx, "EditCompanyInfo: update name failed :%v", err)
			return nil, err
		}
	}

	return res, nil
}

func UploadOrganizeLogo(ctx *gin.Context) (res dtoorganize.UploadOrganizeLogoResponse, err error) {
	purpose := ctx.Request.FormValue("purpose")
	companyID := runtime.CompanyID(ctx)
	defer func() {
		_ = svccoze.SpaceSync(ctx)
	}()
	// 白名单
	if purpose != "company-logo" {
		logs.WarnContextf(ctx, "[UploadOrganizeLogo]: invalid purpose: %s", purpose)
		runtime.BadRequest(ctx, i18n.T(runtime.GetLanguage(ctx), "account_invalid_purpose")) // 无效的上传目的
		return
	}
	f, fh, err := ctx.Request.FormFile("file")
	if err != nil {
		logs.ErrorContextf(ctx, "upload image error: %v", err)
		runtime.BadRequest(ctx, i18n.T(runtime.GetLanguage(ctx), "account_invalid_parameters")) // 参数错误
		return
	}
	defer func(f multipart.File) {
		err := f.Close()
		if err != nil {
			logs.ErrorContextf(ctx, "upload image close faild error: %v", err)
		}
	}(f)

	fi := &storage.FileInfo{
		CompanyID: runtime.CompanyID(ctx),
		Uin:       runtime.Uin(ctx),
		Filename:  fh.Filename,
		Size:      fh.Size,
	}

	ext := strings.ToLower(filepath.Ext(fh.Filename))
	fi.FileExt = ext

	fi.StoragePath = storage.GenerateFileStoragePath(purpose, fi.Uin, ext)

	st, err := storage.LoadStorager(purpose)
	if err != nil {
		logs.ErrorContextf(ctx, "upload image error: %v", err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "account_storage_load_failed")) // 加载存储服务失败
		return
	}
	err = st.Save(ctx, fi, f)
	if err != nil {
		logs.ErrorContextf(ctx, "upload image error: %v", err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "account_storage_save_failed")) // 保存文件失败
		return
	}
	fi.PublicURL = st.GetPublicURL(fi.StoragePath, false)
	//
	if version.DeployMode() != "" && version.DeployMode() != global.DeployModeOpenPO {
		//私有化且非海外
		var (
			cfg config.StorageConfig
		)
		if err = settings.GetYaml(settings.SettingGroupCore, storage.SettingPrefix+purpose, &cfg); err != nil {
			logs.ErrorContextf(ctx, "get storage config [group:%v|key:%v] error: %v", settings.SettingGroupCore, storage.SettingPrefix+purpose, err)
			runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_get_upload_config_failed")) // 获取上传配置失败
			return
		}
		referer := ctx.GetHeader("Referer")
		// 解析 URL
		parsedURL, err := stdurl.Parse(referer)
		if err != nil {
			logs.ErrorContextf(ctx, "get referer error: %v", err)
			runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_parse_url_failed")) // 解析url失败
			return res, nil
		}

		cfg.S3.EndPoint = fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
		st, err := storage.NewStorageWithCfg(cfg)
		if err != nil {
			logs.ErrorContextf(ctx, "[PreviewFileByURL] new storage error: %v cfg[+%v]", err, cfg)
			runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_create_storage_failed")) // 创建存储器失败
			return res, nil
		}
		url := st.GetPublicURL(fi.StoragePath, false)

		fi.PublicURL = url
		res.Response.PublicUrl = url
	}

	if err = dbutil.Account().Transaction(func(tx *gorm.DB) error {
		if err := tx.Table(accounttype.TableNameCompany).
			Where("id = ?", companyID).
			Update("logo", fi.PublicURL).
			Error; err != nil {
			logs.ErrorContextf(ctx, "update logo url error: %v", err)
			return err
		}
		if err := dbutil.Core().Create(fi).Error; err != nil {
			logs.ErrorContextf(ctx, "save core file error: %v", err)
			return err
		}
		return nil
	}); err != nil {
		logs.ErrorContextf(ctx, "upload logo transaction error: %v", err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "account_database_save_failed")) // 数据库保存失败
		return
	}

	res.Response.FileID = fi.ID
	res.Response.PublicUrl = fi.PublicURL

	return res, nil
}

func GetCompanyInfo(ctx *gin.Context, _ *dtoorganize.GetCompanyInfoRequest) (res *dtoorganize.GetCompanyInfoResponse, err error) {
	res = &dtoorganize.GetCompanyInfoResponse{}
	companyID := runtime.CompanyID(ctx)

	cmp, err := company.GetCompanyByID(companyID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetCompanyInfo: GetCompanyByID(id:%v) failed, err: %v", companyID, err)
		return nil, err
	}

	res.Response.Company = *cmp
	return res, nil
}

func CreateDepartmentEmployeePrivate(ctx *gin.Context, req *dtoorganize.CreateDepartmentEmployeePrivateRequest) (res *dtoorganize.CreateDepartmentEmployeePrivateResponse, err error) {
	//res
	res = &dtoorganize.CreateDepartmentEmployeePrivateResponse{}
	resp, err := CreateEmployeePrivate(ctx, &req.Request.EmployeeInfo)
	if err != nil {
		logs.ErrorContextf(ctx, "CreateDepartmentEmployeePrivate: CreateEmployee failed, %+v", err)
	}
	if resp != nil {
		res.BaseResponse = *resp
	}
	res.Response.EmployeeInfo = &req.Request.EmployeeInfo
	return res, nil
}

func EditDepartmentEmployeePrivate(ctx *gin.Context, req *dtoorganize.EditDepartmentEmployeePrivateRequest) (res *dtoorganize.EditDepartmentEmployeePrivateResponse, err error) {
	//res
	res = &dtoorganize.EditDepartmentEmployeePrivateResponse{}
	resp, emp, err := EditEmployee(ctx, *req.Request.EmployeeInfo, runtime.CompanyID(ctx))
	if err != nil {
		logs.ErrorContextf(ctx, "EditDepartmentEmployeePrivate(%+v) failed, %+v", req, err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "account_edit_employee_failed"
		return nil, err
	}
	if resp != nil {
		res.BaseResponse = *resp
	}

	if emp != nil {
		res.Response.EmployeeInfo = emp
	}
	return res, nil
}

func ChangePasswordNotice(ctx *gin.Context, req *dtoorganize.ChangePasswordNoticeRequest) (res *dtoorganize.ChangePasswordNoticeResponse, err error) {
	res = &dtoorganize.ChangePasswordNoticeResponse{}

	if req.Request.AlwaysIgnore {
		u, err := user.GetUserByID(req.Request.UserID)
		if err != nil {
			logs.ErrorContextf(ctx, "ChangePasswordNotice: GetUserByID(user_id:%v) failed, %+v", req.Request.UserID, err)
			return nil, fmt.Errorf("GetUserByID(user_id:%v) failed :%w", req.Request.UserID, err)
		}
		u.PasswordChanged = 1
		if err = dbutil.Account().Save(u).Error; err != nil {
			logs.ErrorContextf(ctx, "ChangePasswordNotice: Save(%+v) failed, %+v", *u, err)
			return nil, err
		}
		res.Response.PasswordChanged = true
		return res, nil
	}
	return res, nil
}

func ChangeDefaultPassword(ctx *gin.Context, req *dtoorganize.ChangeDefaultPasswordRequest) (res *dtoorganize.ChangeDefaultPasswordResponse, err error) {
	res = &dtoorganize.ChangeDefaultPasswordResponse{}

	u, err := user.GetUserByID(req.Request.UserID)
	if err != nil {
		logs.ErrorContextf(ctx, "ChangePasswordNotice: GetUserByID(user_id:%v) failed, %+v", req.Request.UserID, err)
		return nil, fmt.Errorf("GetUserByID(user_id:%v) failed :%w", req.Request.UserID, err)
	}

	if err = bcrypt.CompareHashAndPassword([]byte(*u.Password), []byte(req.Request.OldPassword)); err != nil {
		logs.ErrorContextf(ctx, "LoginByPassword: password not match, %s", err)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "account_invalid_password" // 用户或密码错误
		return
	}

	if err = user.UpdateAccountPassword(u.ID, req.Request.NewPassword); err != nil {
		logs.ErrorContextf(ctx, "update account password failed, user_id=%d, error=%s", u.ID, err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "account_update_password_failed" // 修改密码失败，请稍后再试
		return
	}

	return res, nil
}

func ResetPassword(ctx *gin.Context, req *dtoorganize.ResetPasswordRequest) (res *dtoorganize.ResetPasswordResponse, err error) {
	userEntity, err := user.GetUserByUin(ctx, req.Request.Uin)
	if err != nil {
		logs.ErrorContextf(ctx, "[ResetPassword] GetUserByUin failed, uin=%d, error=%s", req.Request.Uin, err)
		return nil, fmt.Errorf("GetUserByUin failed, uin=%d, error=%v", req.Request.Uin, err)
	}
	res = &dtoorganize.ResetPasswordResponse{}
	if userEntity.ID == 0 {
		logs.WarnContextf(ctx, "[ResetPassword] user not found, uin=%d", req.Request.Uin)
		res.Code = errcode.ErrCode_BadRequest
		res.Message = "account_user_not_found"
		return res, nil
	}
	var passwd string
	if version.DeployMode() != "" {
		if userEntity.Email == nil {
			logs.ErrorContextf(ctx, "[ResetPassword] user email is empty, uin=%d", req.Request.Uin)
			res.Code = errcode.ErrCode_BadRequest
			res.Message = "account_user_email_empty"
			return res, nil
		}
		passwd = DefaultPrivatePasswordPrefix + *userEntity.Email
	} else {
		if userEntity.Phone == nil {
			logs.ErrorContextf(ctx, "[ResetPassword] user phone is empty, uin=%d", req.Request.Uin)
			res.Code = errcode.ErrCode_BadRequest
			res.Message = "account_user_phone_empty"
			return res, nil
		}
		phone := *userEntity.Phone
		passwd = phone[len(phone)-DefaultPasswdLen:]
	}

	if err = user.UpdateAccountPassword(userEntity.ID, passwd); err != nil {
		logs.ErrorContextf(ctx, "[ResetPassword] update account password failed, uin=%d, error=%s", req.Request.Uin, err)
		res.Code = errcode.ErrCode_InternalError
		res.Message = "account_update_password_failed"
	}

	return res, nil
}

func UploadWebSiteLogo(ctx *gin.Context) (res dtoorganize.UploadOrganizeLogoResponse, err error) {
	purpose := ctx.Request.FormValue("purpose")

	// 白名单
	if purpose != "company-logo" {
		logs.WarnContextf(ctx, "[UploadWebSiteLogo]: invalid purpose: %s", purpose)
		runtime.BadRequest(ctx, i18n.T(runtime.GetLanguage(ctx), "account_invalid_purpose")) // 无效的上传目的
		return
	}
	f, fh, err := ctx.Request.FormFile("file")
	if err != nil {
		logs.ErrorContextf(ctx, "upload image error: %v", err)
		runtime.BadRequest(ctx, i18n.T(runtime.GetLanguage(ctx), "account_invalid_parameters")) // 参数错误
		return
	}
	defer func(f multipart.File) {
		err := f.Close()
		if err != nil {
			logs.ErrorContextf(ctx, "upload image close faild error: %v", err)
		}
	}(f)

	fi := &storage.FileInfo{
		CompanyID: runtime.CompanyID(ctx),
		Uin:       runtime.Uin(ctx),
		Filename:  fh.Filename,
		Size:      fh.Size,
	}

	ext := strings.ToLower(filepath.Ext(fh.Filename))
	fi.FileExt = ext

	fi.StoragePath = storage.GenerateFileStoragePath(purpose, fi.Uin, ext)

	st, err := storage.LoadStorager(purpose)
	if err != nil {
		logs.ErrorContextf(ctx, "upload image error: %v", err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "account_storage_load_failed")) // 加载存储服务失败
		return
	}
	err = st.Save(ctx, fi, f)
	if err != nil {
		logs.ErrorContextf(ctx, "upload image error: %v", err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "account_storage_save_failed")) // 保存文件失败
		return
	}
	fi.PublicURL = st.GetPublicURL(fi.StoragePath, false)
	//
	if version.DeployMode() != "" && version.DeployMode() != global.DeployModeOpenPO {
		//私有化且非海外
		var (
			cfg config.StorageConfig
		)
		if err = settings.GetYaml(settings.SettingGroupCore, storage.SettingPrefix+purpose, &cfg); err != nil {
			logs.ErrorContextf(ctx, "get storage config [group:%v|key:%v] error: %v", settings.SettingGroupCore, storage.SettingPrefix+purpose, err)
			runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_get_upload_config_failed")) // 获取上传配置失败
			return
		}
		referer := ctx.GetHeader("Referer")
		// 解析 URL
		parsedURL, err := stdurl.Parse(referer)
		if err != nil {
			logs.ErrorContextf(ctx, "get referer error: %v", err)
			runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_parse_url_failed")) // 解析url失败
			return res, nil
		}

		cfg.S3.EndPoint = fmt.Sprintf("%s://%s", parsedURL.Scheme, parsedURL.Host)
		st, err := storage.NewStorageWithCfg(cfg)
		if err != nil {
			logs.ErrorContextf(ctx, "[PreviewFileByURL] new storage error: %v cfg[+%v]", err, cfg)
			runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_create_storage_failed")) // 创建存储器失败
			return res, nil
		}
		url := st.GetPublicURL(fi.StoragePath, false)

		fi.PublicURL = url
		res.Response.PublicUrl = url
	}

	if err := dbutil.Core().WithContext(ctx).Create(fi).Error; err != nil {
		logs.ErrorContextf(ctx, "save core file error: %v", err)
		return res, err
	}

	res.Response.FileID = fi.ID
	res.Response.PublicUrl = fi.PublicURL

	return res, nil
}

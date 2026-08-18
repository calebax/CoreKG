package privite

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/employee"
	"github.com/insmtx/corekg/apps/account/models/user"
	"github.com/insmtx/corekg/apps/admin/models/login_setting"
	"github.com/insmtx/corekg/apps/kecore/services/svccoze"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/dbtools/redispool"
	"github.com/ygpkg/yg-go/i18n"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/random"
	"github.com/ygpkg/yg-go/settings"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// CreateEmployee 创建员工
// @Tags 员工管理
// @Summary 创建员工
// @Description 创建员工
// @Router /account.CreateEmployee [post]
// @Param user body CreateEmployeeRequest true "入参"
// @Success 200 {object} apiobj.BaseResponse "返回值"
func CreateEmployee(ctx *gin.Context, req *CreateEmployeeRequest, resp *apiobj.BaseResponse) {
	if req.Valid(resp); resp.Code != 0 {
		logs.WarnContextf(ctx, "CreateEmployee validate params failed")
		return
	}
	defer func() {
		_ = svccoze.SpaceSync(ctx)
	}()
	companyID := runtime.CompanyID(ctx)
	var phone *string
	if len(req.Request.Phone) == 0 {
		phone = nil
	} else {
		phone = &req.Request.Phone
	}
	if exist := user.ExistUser(ctx, phone, &req.Request.Email, companyID); exist {
		logs.ErrorContextf(ctx, "User info has already exist about name[%v] phone[%v] email[%v]", req.Request.UserName, req.Request.Phone, req.Request.Email)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_user_info_already_exists" // 用户名,手机号或邮箱已存在
		return
	}

	// 检查用户标识是否存在
	idt := random.String(6)
	exi, err := user.ExistsUserByIIdentify(ctx, idt)
	if err != nil {
		logs.ErrorContextf(ctx, "ExistsUserByIIdentify: exists user failed, %+v", err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "account_check_user_identify_failed")) // 检查用户标识失败
		return
	}
	if exi {
		idt = fmt.Sprintf("%s%d", idt, rand.Intn(10))
	}

	pwd, err := bcrypt.GenerateFromPassword([]byte(req.Request.Password), bcrypt.DefaultCost)
	if err != nil {
		logs.ErrorContextf(ctx, "bcrypt.GenerateFromPassword:uin[%v] desire to create pwd failed, %+v", runtime.Uin(ctx), err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_generate_password_failed" // 生成密码失败
		return
	}

	pwdStr := string(pwd)

	if err := dbutil.Account().Transaction(func(tx *gorm.DB) error {
		//create user
		var (
			u = &accounttype.User{
				Identify:  idt,
				Name:      req.Request.UserName,
				Email:     &req.Request.Email,
				Phone:     phone,
				Password:  &pwdStr,
				AvatarURL: "/images/default-user-avatar.svg",
			}
			ui = &accounttype.UserIdentification{
				Name:        u.Name,
				SubjectType: accounttype.SubjectTypeCompany,
				SubjectID:   companyID,
				UinStatus:   accounttype.UinStatusNormal,
				Issuer:      global.IssuerYYGU,
			}
		)

		if err := tx.Create(u).Error; err != nil {
			return err
		}
		ui.UserID = u.ID
		//create uin
		if err := tx.Create(ui).Error; err != nil {
			return err
		}

		//create employee
		return tx.Create(&accounttype.Employee{
			UserID:    u.ID,
			Uin:       ui.ID,
			CompanyID: companyID,
			SysRole:   accounttype.SysRoleSysEmployee,
		}).Error
	}); err != nil {
		logs.ErrorContextf(ctx, "CreateEmployee failed, %+v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_create_employee_failed" // 创建员工失败
	}
	return
}

// EditEmployee 编辑员工信息
// @Tags 员工管理
// @Summary 编辑员工信息
// @Description 编辑员工信息
// @Router /account.EditEmployee [post]
// @Param user body EditEmployeeRequest true "入参"
// @Success 200 {object} apiobj.BaseResponse "返回值"
func EditEmployee(ctx *gin.Context, req *EditEmployeeRequest, resp *apiobj.BaseResponse) {
	if req.Valid(resp); resp.Code != 0 {
		logs.WarnContextf(ctx, "EditEmployee validate params failed")
		return
	}
	emp, err := employee.GetEmployeeByUin(req.Request.Uin)
	if err != nil {
		logs.ErrorContextf(ctx, "GetEmployeeByUin(%d):%+v", req.Request.Uin, err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "account_get_employee_info_failed")) // 获取员工信息失败
		return
	}

	uin := runtime.Uin(ctx)
	cmpID := runtime.CompanyID(ctx)
	if emp.CompanyID != cmpID {
		logs.ErrorContextf(ctx, "uin[%v] desire to edit other employee uin[%v]", uin, emp.Uin)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "account_edit_illegal_employee_uin")) // 修改非法员工uin
		return
	}

	if emp.Uin != uin && emp.SysRole == accounttype.SysRoleSysAdmin {
		logs.ErrorContextf(ctx, "uin[%v] desire to edit other uin[%v]admin's info", emp.Uin, uin)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "account_edit_other_admin_info_forbidden")) // 禁止修改其他管理员信息
		return
	}
	var phone *string
	if len(req.Request.Phone) == 0 {
		phone = nil
	} else {
		phone = &req.Request.Phone
	}

	if exist := user.ExistUserAfterEdit(ctx, req.Request.UserName, phone, &req.Request.Email, cmpID, emp.UserID); exist {
		logs.ErrorContextf(ctx, "userinfo has already exist detail[%v]", req.Request)
		runtime.BadRequest(ctx, i18n.T(runtime.GetLanguage(ctx), "account_duplicate_personal_info")) // 存在重复个人信息
		return
	}

	uinObj, err := user.GetUserIdentificationByUIN(ctx, req.Request.Uin)
	if err != nil {
		logs.ErrorContextf(ctx, "GetUserIdentificationByUIN(%d):%+v", req.Request.Uin, err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "account_get_uin_failed")) // 获取uin失败
		return
	}
	var (
		nameChanged = false
		infoChanged = false
		pwdStr      string
	)
	if uinObj.Name != req.Request.UserName {
		uinObj.Name = req.Request.UserName
		nameChanged = true
	}
	u, err := user.GetUserByID(uinObj.UserID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetUserByID(%d):%+v", uinObj.UserID, err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "account_get_user_info_failed")) // 获取用户信息失败
		return
	}
	if (len(req.Request.Phone) > 0 && (u.Phone == nil || *u.Phone != req.Request.Phone)) || *u.Email != req.Request.Email {
		u.Phone = &req.Request.Phone
		u.Email = &req.Request.Email
		infoChanged = true
	}
	if phone == nil {
		u.Phone = phone
		infoChanged = true
	}

	if len(req.Request.Password) > 0 {
		pwd, err := bcrypt.GenerateFromPassword([]byte(req.Request.Password), bcrypt.DefaultCost)
		if err != nil {
			logs.ErrorContextf(ctx, "bcrypt.GenerateFromPassword:%+v", err)
			runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "account_generate_password_failed")) // 生成密码失败
			return
		}
		pwdStr = string(pwd)
		u.Password = &pwdStr
		infoChanged = true

	}

	if err := dbutil.Account().Transaction(func(tx *gorm.DB) error {
		if nameChanged {
			if err := tx.Save(uinObj).Error; err != nil {
				return err
			}
		}
		if infoChanged {
			return tx.Save(u).Error
		}
		return nil
	}); err != nil {
		logs.ErrorContextf(ctx, "Edit Employee info faild :%+v", err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "account_update_employee_info_failed")) // 更新员工信息失败
		return
	}
}

// LoginByPasswordPrivate 私有化账户密码登陆
// @Tags NewAccount
// @Summary 私有化账户密码登陆
// @Description 私有化账户密码登陆
// @Router /account.LoginByPasswordPrivate [post]
// @Param user body LoginPasswordRequest true "入参"
// @Success 200 {object} LoginThirdResponse
func LoginByPasswordPrivate(ctx *gin.Context, req *LoginPasswordRequest, resp *LoginThirdResponse) {
	if req.Request.Password == "" || req.Request.Username == "" {
		logs.ErrorContextf(ctx, "LoginByPassword: invalid password or username")
		runtime.BadRequest(ctx, i18n.T(runtime.GetLanguage(ctx), "account_invalid_username_or_password")) // 用户名或密码无效
		return
	}

	//logs.DebugContextf(ctx, "LoginByPassword: request: %+v", *req)
	var (
		delay = 5 * time.Minute

		//origin       = ctx.ClientIP()
		accountRdsKey = user.RedisKeyLoginAccount(req.Request.Username)
		failCount     = 0
		errAPI        error
		err           error
	)

	maxLoginTry, errAPI := settings.GetValue("account", "max_login_try")
	if errAPI != nil {
		logs.ErrorContextf(ctx, "LoginByPasswordPrivate: get max_login_try error, %s", errAPI)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "account_get_login_try_failed")) // 获取登陆尝试配置失败
		return
	}

	maxTry, errAPI := strconv.Atoi(maxLoginTry)
	if errAPI != nil {
		logs.ErrorContextf(ctx, "LoginByPassword: try conv str[%v] to int failed %v", maxLoginTry, errAPI)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "account_login_try_conversion_failed")) // 尝试配置转换失败
		return
	}

	logs.DebugContextf(ctx, "LoginByPasswordPrivate: account: %s, accountRdsKey: %s", req.Request.Username, accountRdsKey)
	//check if account has already called this login api multi times
	if exist, ttl := redispool.IsExistKey(accountRdsKey); !exist {
		logs.DebugContextf(ctx, "LoginByPasswordPrivate: redis key %s not exist,this is first login in last [%v] seconds", accountRdsKey, delay.Seconds())
		if errAPI = redispool.SetString(accountRdsKey, fmt.Sprintf("%v", 0), delay); errAPI != nil {
			logs.ErrorContextf(ctx, "set redis key [%v] failed", accountRdsKey)
			runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "account_account_cache_validation_failed")) // 登陆Account缓存验证失败
			return
		}
	} else {
		countStr, errAPI := redispool.GetString(accountRdsKey)
		if errAPI != nil {
			logs.ErrorContextf(ctx, "LoginByPasswordPrivate: get redis key [%v] failed", accountRdsKey)
			runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "account_get_account_cache_failed")) // 获取登陆Account缓存失败
			return
		}

		logs.DebugContextf(ctx, "Get account redis key [%v] value[%v]", accountRdsKey, countStr)

		count, errAPI := strconv.Atoi(countStr)
		if errAPI != nil {
			logs.ErrorContextf(ctx, "LoginByPasswordPrivate: convert redis key [%v] value[%v] to int failed %v", accountRdsKey, countStr, errAPI)
			runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "account_account_cache_parsing_failed")) // 登陆Account缓存解析失败
			return
		}

		if count > maxTry {
			logs.ErrorContextf(ctx, "account[%v] has already up to MaxLoginTry[%v]\nAPI[loginByPassword] has been banned for %v seconds, debug with key:%v",
				req.Request.Username, maxTry, ttl.Seconds(), accountRdsKey)
			runtime.InternalError(ctx, i18n.TWithData(runtime.GetLanguage(ctx), "account_login_attempts_exceeded", map[string]interface{}{"max_try": maxTry})) // 登陆Account失败次数已达上限
			return
		}
		failCount = count
	}

	defer func() {
		if err != nil {
			if err := redispool.SetString(accountRdsKey, fmt.Sprintf("%v", failCount+1), delay); err != nil {
				logs.ErrorContextf(ctx, "LoginByPasswordPrivate: set redis key [%v] value[%v] failed", accountRdsKey, failCount+1)
				runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "account_update_account_cache_failed")) // 更新Account登陆缓存失败
				return
			}
		}
	}()

	lst, err := login_setting.GetLoginSettingByPath(req.Request.DomainName, "")
	if err != nil {
		logs.ErrorContextf(ctx, "LoginByPasswordPrivate: get login setting failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_get_login_setting_failed" // 获取登录设置失败
		return
	}

	var usr *accounttype.User
	if strings.Contains(req.Request.Username, "@") {
		usr, err = user.GetUserByEmail(req.Request.Username)
	} else {
		usr, err = user.GetUserByPhone(req.Request.Username)
	}
	if err != nil || usr == nil || usr.ID == 0 || usr.Password == nil || *usr.Password == "" {
		logs.ErrorContextf(ctx, "LoginByPasswordPrivate: user not found or invalid password")
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_user_or_password" // 用户或密码错误
		return
	}
	if usr.Password == nil || *usr.Password == "" {
		logs.ErrorContextf(ctx, "LoginByPasswordPrivate: user password not set")
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_user_or_password" // 用户或密码错误
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(*usr.Password), []byte(req.Request.Password))
	if err != nil {
		logs.ErrorContextf(ctx, "LoginByPasswordPrivate: password not match, %s", err)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_user_or_password" // 用户或密码错误
		return
	}

	resp.Response.UserInfo = &user.UserInfo{
		Identify:        usr.Identify,
		AvatarURL:       usr.AvatarURL,
		Bio:             usr.Bio,
		Name:            usr.Name,
		PasswordChanged: usr.PasswordChanged,
	}
	resp.Response.Issuer = lst.Issuer
	resp.Response.AllowRegister = lst.AllowRegister
	resp.Response.LoginWay = 8
	resp.Response.UserID = usr.ID
	resp.Response.LoginStatus = "success"

	if usr.Email != nil {
		resp.Response.UserInfo.Email = *usr.Email
	}

	uins, err := user.GetUserUinsByUserID(ctx, usr.ID, resp.Response.Issuer)
	if err != nil {
		logs.ErrorContextf(ctx, "LoginByPasswordPrivate: get user uins failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_get_user_uins_failed" // 获取用户标识信息失败
		return
	}

	luin, err := GetUinType(ctx, uins, resp.Response.Issuer)
	if err != nil {
		logs.ErrorContextf(ctx, "LoginByPasswordPrivate: get uin type failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_uin_fetch_failed" // 获取用户信息失败
		return
	}
	resp.Response.Uin = luin

	token := user.GenerateJwtToken(ctx, uins[0].ID, resp.Response.LoginWay, runtime.GetRealIP(ctx.Request), resp.Response.Issuer)
	if token == "" {
		logs.ErrorContextf(ctx, "LoginByPasswordPrivate: generate jwt token failed")
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_jwt_token_generation_failed" // 切换身份失败
		return
	}
	resp.Response.UserInfo.Uin = uins[0].ID
	resp.Response.JwtToken = token
	refresh, err := user.GenerateRefreshToken(ctx, usr.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "loginSuccess: generate refresh token failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		return
	}
	resp.Response.RefreshToken = refresh
	return

}

// GetUinType 获取登录用户信息
func GetUinType(ctx context.Context, uins []*accounttype.UserIdentification, issuer string) ([]*LoginUin, error) {
	var (
		loginUin []*LoginUin
		uinids   []uint
		userID   uint
	)

	// 收集公司 ID 和用户 ID
	for _, uin := range uins {
		if uin.SubjectType == accounttype.SubjectTypeCompany && uin.Issuer == issuer {
			uinids = append(uinids, uin.ID)
		} else if uin.SubjectType == accounttype.SubjectTypeIndividual {
			userID = uin.UserID
		}
	}

	// 获取公司员工信息
	emps, err := employee.GetCompanyEmployeeInfo(uinids)
	if err != nil {
		logs.ErrorContextf(ctx, "GetUinType: account_get_company_employee_info_failed, %s", err)
		return nil, fmt.Errorf("account_get_company_employee_info_failed")
	}

	// 按 CompanyID 分组
	employeeMap := make(map[uint]*employee.CompanyEmployeeInfo)
	for _, emp := range emps {
		employeeMap[emp.CompanyID] = emp
	}

	// 构建登录用户信息
	for _, uin := range uins {
		if uin.SubjectType == accounttype.SubjectTypeCompany {
			emp, exists := employeeMap[uin.SubjectID]
			if !exists || emp == nil {
				logs.WarnContextf(ctx, "GetUinType: account_company_info_not_found, SubjectID: %d", uin.SubjectID)
				continue
			}
			loginUin = append(loginUin, &LoginUin{
				Uin:           *uin,
				CompanyName:   emp.CompanyName,
				CompanyLogo:   emp.CompanyLogo,
				Role:          emp.SysRole,
				CompanyStatus: emp.CompanyStatus,
			})
		} else if uin.SubjectType == accounttype.SubjectTypeIndividual {
			u, err := user.GetUserByID(userID)
			if err != nil {
				logs.ErrorContextf(ctx, "GetUinType: account_get_user_failed, %s", err)
				return nil, fmt.Errorf("account_get_user_failed")
			}
			loginUin = append(loginUin, &LoginUin{
				Uin:  *uin,
				Name: u.Name,
			})
		}
	}

	return loginUin, nil
}

// DeleteEmployeePrivate 删除员工及用户(私有化专用)
// @Tags 账号管理
// @Summary 删除员工及用户(私有化专用)
// @Description 删除员工及用户(私有化专用)
// @Router /account.DeleteEmployeePrivate [post]
// @Param request body DeleteEmployeeRequest true "入参"
// @Success 200 {object} apiobj.BaseResponse
func DeleteEmployeePrivate(ctx *gin.Context, in *DeleteEmployeeRequest, out *apiobj.BaseResponse) error {
	defer func() {
		_ = svccoze.SpaceSync(ctx)
	}()
	if in.Request.EmployeeID == 0 {
		out.Code = errcode.ErrCode_BadRequest
		out.Message = "account_employee_id_empty" // 用户ID不能为空
		return nil
	}

	emp, err := employee.GetEmployeeByID(ctx, in.Request.EmployeeID)
	if err != nil {
		logs.ErrorContextf(ctx, "[account][GetEmployeeByID] get employee error: %v", err)
		out.Code = errcode.ErrCode_InternalError
		out.Message = "account_get_employee_failed" // 获取员工信息失败
		return nil
	}

	if emp.CompanyID != runtime.CompanyID(ctx) {
		logs.ErrorContextf(ctx, "[account][user] invalid company[id:%v][uin:%v] area delete action %v -> delete %v",
			runtime.CompanyID(ctx), runtime.Uin(ctx), emp.CompanyID, in.Request.EmployeeID)
		out.Code = errcode.ErrCode_InternalError
		out.Message = "account_invalid_company_employee" // 非法公司员工ID
		return nil
	}

	if emp.SysRole == accounttype.SysRoleSysAdmin {
		logs.ErrorContextf(ctx, "[account][user] invalid delete action %v -> delete [RoleAdmin]%v",
			runtime.EmployeeID(ctx), in.Request.EmployeeID)
		out.Code = errcode.ErrCode_InternalError
		out.Message = "account_delete_admin_forbidden" // 禁止删除管理员
		return nil
	}

	if err = employee.DeleteEmployeeWithUser(dbutil.Account(), in.Request.EmployeeID); err != nil {
		logs.ErrorContextf(ctx, "[account][DeleteEmployee] delete employee error: %v", err)
		out.Code = errcode.ErrCode_InternalError
		out.Message = "account_delete_employee_failed" // 删除用户失败
		return nil
	}

	return nil
}

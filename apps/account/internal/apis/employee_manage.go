package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/employee"
	"github.com/insmtx/corekg/apps/kecore/services/svccoze"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
)

// ListEmployee list users
// @Tags 员工管理
// @Summary 员工列表
// @Description 员工列表
// @Router /account.ListEmployee [post]
// @Param request body ListEmployeeRequest true "入参"
// @Success 200 {object} ListEmployeeResponse
func ListEmployee(ctx *gin.Context, in *ListEmployeeRequest, out *ListEmployeeResponse) error {
	in.Request.CompanyID = runtime.CompanyID(ctx)
	logs.InfoContextf(ctx, "[account][user] query user list, %+v", in.Request)

	if in.Validity(&out.BaseResponse); out.Code != 0 {
		return nil
	}

	err := employee.QueryEmployeeList(ctx, in.Request, &out.Response)
	if err != nil {
		out.Code = errcode.ErrCode_InternalError
		out.Message = "account_list_employee_failed" // 获取用户列表失败
		return nil
	}

	return nil
}

// GetEmployeeDetail 获取员工详情信息
// @Tags 员工管理
// @Summary 获取员工详情信息
// @Description 获取员工详情信息
// @Router /account.GetEmployeeDetail [post]
// @Param request body GetEmployeeRequest true "入参"
// @Success 200 {object} EmployeeResponse
func GetEmployeeDetail(ctx *gin.Context, in *GetEmployeeRequest, out *EmployeeResponse) error {
	logs.InfoContextf(ctx, "[account][user] query user, %d", in.Request.EmployeeID)

	err := employee.GetEmployeeDetailByID(ctx, in.Request.EmployeeID, &out.Response.EmployeeDetail)
	if err != nil {
		out.Code = errcode.ErrCode_InternalError
		out.Message = "account_get_employee_detail_failed" // 获取用户详情失败
		return nil
	}

	return nil
}

// UpdateEmployee update user
// @Tags 账号管理
// @Summary 更新用户
// @Description 更新用户
// @Router /account.UpdateEmployee [post]
// @Param request body UpdateEmployeeRequest true "入参"
// @Success 200 {object} apiobj.BaseResponse
func UpdateEmployee(ctx *gin.Context, in *UpdateEmployeeRequest, out *apiobj.BaseResponse) error {
	if in.Validity(out); out.Code != 0 {
		return nil
	}
	// 更新员工信息
	err := employee.UpdateEmployee(ctx, in.Request)
	if err != nil {
		out.Code = errcode.ErrCode_InternalError
		out.Message = "account_update_employee_failed" // 更新用户失败
		return nil
	}

	return nil
}

// DeleteEmployee delete user
// @Tags 账号管理
// @Summary 删除用户
// @Description 删除用户
// @Router /account.DeleteEmployee [post]
// @Param request body DeleteEmployeeRequest true "入参"
// @Success 200 {object} apiobj.BaseResponse
func DeleteEmployee(ctx *gin.Context, in *DeleteEmployeeRequest, out *apiobj.BaseResponse) error {
	defer func() {
		_ = svccoze.SpaceSync(ctx)
	}()
	if in.Request.EmployeeID == 0 {
		out.Code = errcode.ErrCode_BadRequest
		out.Message = "account_id_empty" // 用户ID不能为空
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
		out.Code = errcode.ErrCode_BadRequest
		out.Message = "account_invalid_company_employee" // 非法公司员工ID
		return nil
	}

	if emp.SysRole == accounttype.SysRoleSysAdmin {
		logs.ErrorContextf(ctx, "[account][user] invalid delete action %v -> delete [RoleAdmin]%v",
			runtime.EmployeeID(ctx), in.Request.EmployeeID)
		out.Code = errcode.ErrCode_BadRequest
		out.Message = "account_delete_admin_forbidden" // 禁止删除管理员
		return nil
	}

	err = employee.DeleteEmployee(in.Request.EmployeeID)
	if err != nil {
		logs.ErrorContextf(ctx, "[account][DeleteEmployee] delete employee error: %v", err)
		out.Code = errcode.ErrCode_InternalError
		out.Message = "account_delete_employee_failed" // 删除用户失败
		return nil
	}

	return nil
}

// ListEmployeeNickID 员工列表昵称ID
// @Tags 员工管理
// @Summary 员工列表昵称ID
// @Description 员工列表昵称ID
// @Router /account.ListEmployeeNickID [post]
// @Param request body ListEmployeeNickIDRequest true "入参"
// @Success 200 {object} ListEmployeeNickIDRequest
func ListEmployeeNickID(ctx *gin.Context, in *ListEmployeeNickIDRequest, out *ListEmployeeNickIDResponse) {
	in.Request.CompanyID = runtime.CompanyID(ctx)
	logs.DebugContextf(ctx, "[account][user] query user list, %+v", in.Request)

	if in.Validity(&out.BaseResponse); out.Code != 0 {
		return
	}

	err := employee.QueryEmployeeSimpleList(ctx, in.Request, &out.Response)
	if err != nil {
		out.Code = errcode.ErrCode_InternalError
		out.Message = "account_list_employee_failed" // 获取用户列表失败
		return
	}
}

// GetCompanyAdmins 获取管理员列表
// @Tags 员工管理
// @Summary 获取管理员列表
// @Description 获取管理员列表
// @Router /account.GetCompanyAdmins [post]
// @Param request body apiobj.BaseRequest true "入参"
// @Success 200 {object} GetCompanyAdminsResponse
func GetCompanyAdmins(ctx *gin.Context, in *apiobj.BaseRequest, out *GetCompanyAdminsResponse) {
	cmpID := runtime.CompanyID(ctx)
	res, err := employee.GetAdminEmployeeByCompanyUD(cmpID)
	if err != nil {
		logs.ErrorContextf(ctx, "[account][GetCompanyAdmins] GetAdminEmployeeByCompanyUD(%v): %v", cmpID, err)
		out.Code = errcode.ErrCode_InternalError
		out.Message = "account_get_company_admins_failed" // 获取公司管理员记录失败
		return
	}
	out.Response.Data = res
}

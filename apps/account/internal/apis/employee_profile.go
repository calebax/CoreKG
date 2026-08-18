package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/employee"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
)

// GetMyAction 获取action_path列表
// @Tags 运营个人中心
// @Summary 获取action_path列表
// @Description 获取action_path列表
// @Router /account.GetMyAction [post]
// @Param user body GetMyActionRequest true "入参"
// @Success 200 {object} GetMyActionResponse "返回值"
func GetMyAction(ctx *gin.Context, req *GetMyActionRequest, resp *GetMyActionResponse) {
	uin := runtime.Uin(ctx)
	emp, err := employee.GetEmployeeByUIN(ctx, uin)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_query_employee_failed" // 查询员工信息失败
		return
	}
	_, _, actionPaths, err := employee.GetEmployeeRbac(emp)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_get_permissions_failed" // 获取权限信息失败
		return
	}

	resp.Response.ActionPaths = actionPaths
}

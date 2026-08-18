package accountmds

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/employee"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/runtime/auth"
	"github.com/ygpkg/yg-go/logs"
)

// RequireSysAdminRole 需要管理员角色
func RequireSysAdminRole(ctx *gin.Context) {
	ls := ctx.MustGet(global.CtxKeyLoginStatus).(*auth.LoginStatus)

	if ls.State != auth.StateSucc {
		logs.ErrorContextf(ctx, "[RequireAdminRole] unauthorized, %+v", ls)
		msg := &apiobj.BaseResponse{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		}
		ctx.AbortWithStatusJSON(200, msg)
		return
	}
	uin := ls.GetID(global.CtxKeyUin)

	// 否则检查员工权限
	if uin == 0 {
		logs.ErrorContextf(ctx, "[RequireAdminRole] missing UIN")
		ctx.AbortWithStatusJSON(http.StatusOK, &apiobj.BaseResponse{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		})
		return
	}

	emp, err := employee.GetEmployeeByUIN(ctx, uin)
	if err != nil {
		logs.ErrorContextf(ctx, "[RequireAdminRole] get employee by UIN failed: %v", err)
		ctx.AbortWithStatusJSON(http.StatusOK, &apiobj.BaseResponse{
			Code:    http.StatusInternalServerError,
			Message: "Internal Error",
		})
		return
	}

	if emp.SysRole != accounttype.SysRoleSysAdmin {
		logs.ErrorContextf(ctx, "[RequireAdminRole] employee(uin:%v)'s role(%v) is not admin: %v", uin, emp.SysRole, err)
		ctx.AbortWithStatusJSON(http.StatusOK, &apiobj.BaseResponse{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		})
		return
	}

	ctx.Next()
}

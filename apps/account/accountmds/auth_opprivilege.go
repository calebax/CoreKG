package accountmds

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/employee"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/pkgs/utils/apipath"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/runtime/auth"
	"github.com/ygpkg/yg-go/logs"
)

// RequireOpPrivilege 需要运营接口权限
func RequireOpPrivilege(ctx *gin.Context) {
	ls := ctx.MustGet(global.CtxKeyLoginStatus).(*auth.LoginStatus)

	//
	if ls.State != auth.StateSucc {
		logs.ErrorContextf(ctx, "[privilege_auth] unauthorized, %+v", ls)
		msg := &apiobj.BaseResponse{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		}
		ctx.AbortWithStatusJSON(200, msg)
		return
	}

	apiPath := apipath.ExtractAPIFromRequestURI(ctx.Request.RequestURI)
	uin := ls.GetID(global.CtxKeyUin)
	// 否则检查员工权限
	if uin == 0 {
		logs.ErrorContextf(ctx, "[privilege_auth] missing UIN")
		ctx.AbortWithStatusJSON(http.StatusOK, &apiobj.BaseResponse{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		})
		return
	}

	emp, err := employee.GetEmployeeByUIN(ctx, uin)
	if err != nil {
		logs.ErrorContextf(ctx, "[privilege_auth] get employee by UIN failed: %v", err)
		ctx.AbortWithStatusJSON(http.StatusOK, &apiobj.BaseResponse{
			Code:    http.StatusInternalServerError,
			Message: "Internal Error",
		})
		return
	}

	hasPrivilege, err := employee.HasEmployeeApiPrivilege(emp.ID, apiPath)
	if err != nil {
		logs.ErrorContextf(ctx, "[privilege_auth] check employee privilege failed: %v", err)
		ctx.AbortWithStatusJSON(http.StatusOK, &apiobj.BaseResponse{
			Code:    http.StatusInternalServerError,
			Message: "Internal Error",
		})
		return
	}

	if !hasPrivilege {
		logs.WarnContextf(ctx, "[privilege_auth] employee %d no access to %s", emp.ID, apiPath)
		ctx.AbortWithStatusJSON(http.StatusOK, &apiobj.BaseResponse{
			Code:    http.StatusForbidden,
			Message: "Forbidden",
		})
		return
	}

	ctx.Next()
}

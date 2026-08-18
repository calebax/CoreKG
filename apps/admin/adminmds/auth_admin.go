package adminmds

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/admin/models/admintype"
	"github.com/insmtx/corekg/apps/admin/models/employee"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/runtime/auth"
	"github.com/ygpkg/yg-go/config"
	"github.com/ygpkg/yg-go/logs"
)

var managerJwt *config.JwtConfig

// InjectEmployeeLoginStatus 注入管理后台员工登录状态
func InjectEmployeeLoginStatus(ctx *gin.Context, ls *auth.LoginStatus) (err error) {
	if ls.State != auth.StateSucc {
		return
	}

	emp, err := employee.GetEmployeeByUin(ctx, ls.Claim.Uin)
	if err != nil {
		logs.WarnContextf(ctx, "[manager_auth][%s] auth failed, %s", ls.Claim.Uin, err)
		ls.Err = err
		ls.State = auth.StateFailed
		return
	}

	if emp.Status != admintype.UserStatusNormal {
		logs.WarnContextf(ctx, "[manager_auth][%s] user status is not normal, %s",
			emp.ID, emp.Status)
		ls.Err = fmt.Errorf("user status is not normal, %s", emp.Status)
		ls.State = auth.StateFailed
		return
	}

	ls.Role = auth.RoleEmployee
	ls.SetID(global.CtxKeyUin, ls.Claim.Uin)
	ls.SetID(global.CtxKeyCompanyID, 1)
	ls.SetID(global.CtxKeyEmployeeID, emp.ID)

	return
}

// RequireOpPrivilege 需要接口权限
func RequireOpPrivilege(ctx *gin.Context) {
	ls := ctx.MustGet(global.CtxKeyLoginStatus).(*auth.LoginStatus)
	empid := ls.GetID(global.CtxKeyEmployeeID)
	// if ls.State != auth.StateSucc || ls.Role != auth.RoleEmployee || empid == 0 {
	if ls.State != auth.StateSucc || empid == 0 {
		// zsy debug
		// logs.Info("State:", ls.State, "Role:", ls.Role, "empid:", empid)
		logs.ErrorContextf(ctx, "[privilege_auth] unauthorized, %+v", ls)
		msg := &apiobj.BaseResponse{
			Code:    http.StatusUnauthorized,
			Message: "Unauthorized",
		}
		ctx.AbortWithStatusJSON(200, msg)
		return
	}
	emp, err := employee.GetEmployeeByID(ctx, empid)
	if err != nil {
		logs.ErrorContextf(ctx, "[privilege_auth] unauthorized, %+v", ls)
		msg := &apiobj.BaseResponse{
			Code:    http.StatusInternalServerError,
			Message: "Internal Error",
		}
		ctx.AbortWithStatusJSON(200, msg)
		return
	}
	if emp.SysRole == admintype.SysRoleSysAdmin {
		ctx.Next()
		return
	}

	api := strings.TrimPrefix(ctx.Request.RequestURI, global.PrefixAPIV2)
	//zsy debug
	// logs.Info("api: ", api)
	hasPrivilege, err := employee.HasEmployeeApiPrivilege(ctx, empid, api)
	if err != nil {
		logs.ErrorContextf(ctx, "[privilege_auth] unauthorized, %+v", ls)
		msg := &apiobj.BaseResponse{
			Code:    http.StatusInternalServerError,
			Message: "Internal Error",
		}
		ctx.AbortWithStatusJSON(200, msg)
		return
	}
	if !hasPrivilege {
		msg := &apiobj.BaseResponse{
			Code:    http.StatusForbidden,
			Message: "Forbidden",
		}
		ctx.AbortWithStatusJSON(200, msg)
		return
	}

	ctx.Next()
}

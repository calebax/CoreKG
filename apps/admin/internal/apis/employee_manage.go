package apis

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/user"
	"github.com/insmtx/corekg/apps/admin/models/employee"
	"github.com/insmtx/corekg/apps/admin/models/login_setting"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/insmtx/corekg/pkgs/wx"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/apis/runtime/auth"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/types"
)

// ListEmployee list users
// @Tags 账号管理
// @Summary 用户列表
// @Description 用户列表
// @Router /admin.ListEmployee [post]
// @Param request body ListEmployeeRequest true "入参"
// @Success 200 {object} ListEmployeeResponse
func ListEmployee(ctx *gin.Context, in *ListEmployeeRequest, out *ListEmployeeResponse) error {
	in.Request.CompanyID = runtime.CompanyID(ctx)
	logs.InfoContextf(ctx, "[admin][user] query user list, %+v", in.Request)

	if in.Validity(&out.BaseResponse); out.Code != 0 {
		return nil
	}

	err := employee.QueryEmployeeList(ctx, in.Request, &out.Response)
	if err != nil {
		out.Code = errcode.ErrCode_InternalError
		out.Message = fmt.Sprintf("获取用户列表失败: %v", err)
		return nil
	}

	return nil
}

type CreateEmployeeResponse struct {
	apiobj.BaseResponse

	Response struct {
		EmployeeID uint   `json:"employee_id"`
		BindKey    string `json:"bind_key"`
	}
}

// CreateEmployee 添加员工
// @Tags 运营员工
// @Summary 添加员工
// @Description 添加员工
// @Router /admin.CreateEmployee [post]
// @Param user body CreateEmployeeRequest true "入参"
// @Success 200 {object} CreateEmployeeResponse "返回值"
func CreateEmployee(ctx *gin.Context, req *CreateEmployeeRequest, resp *CreateEmployeeResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		logs.WarnContextf(ctx, "CreateEmployee validate params failed")
		return
	}
	emp, err := employee.CreateEmployee(ctx, &req.Request)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = fmt.Sprintf("添加员工失败: %v", err)
		logs.WarnContextf(ctx, "CreateEmployee failed ,err %s", err)
		return
	}

	resp.Response.EmployeeID = emp.ID
	resp.Response.BindKey = types.SafeID(emp.ID).Enc()
}

// UpdateEmployee update user
// @Tags 账号管理
// @Summary 更新用户
// @Description 更新用户
// @Router /admin.UpdateEmployee [post]
// @Param request body UpdateEmployeeRequest true "入参"
// @Success 200 {object} apiobj.BaseResponse
func UpdateEmployee(ctx *gin.Context, in *UpdateEmployeeRequest, out *apiobj.BaseResponse) error {
	if in.Validity(out); out.Code != 0 {
		return nil
	}

	err := employee.UpdateEmployee(ctx, in.Request)
	if err != nil {
		out.Code = errcode.ErrCode_InternalError
		out.Message = fmt.Sprintf("更新用户失败: %v", err)
		return nil
	}

	return nil
}

// UpdateEmployeePassword update user
// @Tags 账号管理
// @Summary 更新用户
// @Description 更新用户
// @Router /admin.UpdateEmployeePassword [post]
// @Param request body UpdateEmployeePasswordRequest true "入参"
// @Success 200 {object} apiobj.BaseResponse
func UpdateEmployeePassword(ctx *gin.Context, in *UpdateEmployeePasswordRequest, out *apiobj.BaseResponse) error {
	if in.Validity(out); out.Code != 0 {
		return nil
	}

	_, err := employee.UpdateEmployeePassword(ctx, in.Request.ID, in.Request.Password)
	if err != nil {
		logs.ErrorContextf(ctx, "[admin] update user password (%d) failed, %s", in.Request.ID, err)
		out.Code = errcode.ErrCode_InternalError
		out.Message = fmt.Sprintf("更新用户失败: %v", err)
		return nil
	}

	return nil
}

// DeleteEmployee delete user
// @Tags 账号管理
// @Summary 删除用户
// @Description 删除用户
// @Router /admin.DeleteEmployee [post]
// @Param request body DeleteEmployeeRequest true "入参"
// @Success 200 {object} apiobj.BaseResponse
func DeleteEmployee(ctx *gin.Context, in *DeleteEmployeeRequest, out *apiobj.BaseResponse) error {
	if in.Request.EmployeeID == 0 {
		out.Code = errcode.ErrCode_BadRequest
		out.Message = "用户ID不能为空"
		return nil
	}

	err := employee.DeleteEmployee(ctx, in.Request.EmployeeID)
	if err != nil {
		out.Code = errcode.ErrCode_InternalError
		out.Message = fmt.Sprintf("删除用户失败: %v", err)
		return nil
	}

	return nil
}

// GetEmployeeBindKey 绑定员工微信
// @Tags 运营员工
// @Summary 获取微信绑定bind_key参数
// @Description 获取微信绑定bind_key参数
// @Router /admin.GetEmployeeBindKey [post]
// @Param user body GetEmployeeBindKeyRequest true "入参"
// @Success 200 {object} GetEmployeeBindKeyResponse "返回值"
func GetEmployeeBindKey(ctx *gin.Context, req *GetEmployeeBindKeyRequest, resp *GetEmployeeBindKeyResponse) {
	if req.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "请选择员工"
		return
	}
	emp, err := employee.GetEmployeeByID(ctx, req.Request.ID)
	if err != nil {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "员工不存在"
		return
	}
	if emp.UnionID != nil {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "员工已经绑定微信"
		return
	}
	resp.Response.EmployeeID = emp.ID
	resp.Response.BindKey = types.SafeID(emp.ID).Enc()
}

// BindEmployeeWechat 绑定员工微信
// @Tags 运营员工
// @Summary 绑定员工微信
// @Description 绑定员工微信
// @Router /admin.BindEmployeeWechat [post]
// @Param request body BindEmployeeWechatRequest true "入参"
// @Success 200 {object} employee.BindEmployeeWechatResponse "返回值"
func BindEmployeeWechat(ctx *gin.Context, req *BindEmployeeWechatRequest, resp *employee.BindEmployeeWechatResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.WarnContextf(ctx, "BindEmployeeWechat validate params failed")
		return
	}

	var empID types.SafeID
	empID.Dec(req.Request.BindKey)
	if empID == 0 {
		logs.ErrorContextf(ctx, "BindEmployeeWechat: decode empID failed, %s", req.Request.BindKey)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "decode empID failed"
		return
	}
	eID := uint(empID)

	emp, err := employee.GetEmployeeByID(ctx, eID)
	if err != nil {
		logs.ErrorContextf(ctx, "BindEmployeeWechat: get employee by id failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = fmt.Sprintf("获取员工失败: %v", err)
		return
	}
	if emp.UnionID != nil {
		logs.WarnContextf(ctx, "BindEmployeeWechat: employee already bind wechat, %s", emp.UnionID)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "员工已经绑定微信"
		return
	}

	login_setting, err := login_setting.GetLoginSettingByPath(req.Request.DomainName, "")
	if err != nil {
		logs.ErrorContextf(ctx, "loginWechatWeb: get login setting failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = fmt.Sprintf("获取登录设置失败: %v", err)
		return
	}
	weapp, err := wx.GetWechatWebOAuth(ctx, "account", login_setting.AuthKey)
	if err != nil {
		logs.ErrorContextf(ctx, "loginWechatWeb: get work wechat oauth config failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = fmt.Sprintf("获取微信OAuth配置失败: %v", err)
		return
	}

	tkn, err := weapp.GetUserAccessToken(req.Request.Code)
	if err != nil {
		logs.ErrorContextf(ctx, "loginWechatWeb: get user access token failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = fmt.Sprintf("获取用户授权信息失败: %v", err)
		return
	}

	ui, err := weapp.GetUserInfo(tkn.AccessToken, tkn.OpenID, "")
	if err != nil {
		logs.ErrorContextf(ctx, "loginWechatWeb: get user info failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = fmt.Sprintf("获取微信用户信息失败: %v", err)
		return
	}

	emp.UnionID = &ui.Unionid
	emp.AvatarURL = ui.HeadImgURL
	if err := dbutil.Account().Save(emp).Error; err != nil {
		logs.ErrorContextf(ctx, "BindEmployeeWechat: save employee failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = fmt.Sprintf("保存员工失败: %v", err)
		return
	}

	resp.Response.JwtToken = user.GenerateJwtToken(ctx, emp.ID, auth.LoginWayWorkWechat, runtime.GetRealIP(ctx.Request), global.IssuerYYGUAdmin)
	if resp.Response.JwtToken == "" {
		logs.ErrorContextf(ctx, "BindEmployeeWechat: generate jwt token failed")
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "生成JWT token失败"
		return
	}

	positions, _, actionPaths, err := employee.GetEmployeeRbac(ctx, emp)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = fmt.Sprintf("获取用户权限失败: %v", err)
		return
	}
	resp.Response.UserInfo = employee.EmployeeLoginInfo{
		EmployeeDetail: employee.EmployeeDetail{
			Employee:    *emp,
			Positions:   positions,
			ActionPaths: actionPaths,
		},
	}
}

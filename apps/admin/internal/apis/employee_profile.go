package apis

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/admin/models/employee"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
)

// GetMyUserInfo 获取个人信息
// @Tags 运营个人中心
// @Summary 获取个人信息
// @Description 获取个人信息
// @Router /op.GetMyUserInfo [post]
// @Param user body GetMyUserInfoRequest true "入参"
// @Success 200 {object} GetMyUserInfoResponse "返回值"
func GetMyUserInfo(ctx *gin.Context, req *GetMyUserInfoRequest, resp *GetMyUserInfoResponse) {
	myID := runtime.EmployeeID(ctx)

	err := employee.GetEmployeeDetailByID(ctx, myID, &resp.Response.EmployeeDetail)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = fmt.Sprintf("获取用户详情失败: %v", err)
		return
	}
}

// GetEmployeeDetail 新增获取用户详情
func GetEmployeeDetail(ctx *gin.Context, in *GetEmployeeRequest, out *EmployeeResponse) error {
	logs.InfoContextf(ctx, "[admin][user] query user, %d", in.Request.EmployeeID)

	err := employee.GetEmployeeDetailByID(ctx, uint(in.Request.EmployeeID), &out.Response.EmployeeDetail)
	if err != nil {
		out.Code = errcode.ErrCode_InternalError
		out.Message = fmt.Sprintf("获取用户详情失败: %v", err)
		return nil
	}
	// }
	// err = employee.GetEmployeePositionsById(in.Request.EmployeeID, out.Response.Positions)
	// if err != nil {
	// 	out.Code = errcode.ErrCode_InternalError
	// 	out.Message = fmt.Sprintf("获取用户职位失败: %v", err)
	// 	return nil
	// }

	return nil
}

// ModifyMyUserInfo 编辑个人信息
// @Tags 运营个人中心
// @Summary 编辑个人信息
// @Description 编辑个人信息
// @Router /op.ModifyMyUserInfo [post]
// @Param user body ModifyMyUserInfoRequest true "入参"
// @Success 200 {object} ModifyMyUserInfoResponse "返回值"
func ModifyMyUserInfo(ctx *gin.Context, req *ModifyMyUserInfoRequest, resp *ModifyMyUserInfoResponse) {
	myID := runtime.EmployeeID(ctx)
	if req.Validity(resp); resp.Code != 0 {
		return
	}

	err := employee.ModifyEmployeeSimple(ctx, myID, req.Request)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = fmt.Sprintf("修改个人信息失败: %v", err)
		return
	}
}

// ChangeMyWechat 换绑微信
// @Tags 运营个人中心
// @Summary 换绑微信
// @Description 换绑微信
// @Router /op.ChangeMyWechat [post]
// @Param user body ChangeMyWechatRequest true "入参"
// @Success 200 {object} ChangeMyWechatResponse "返回值"
func ChangeMyWechat(ctx *gin.Context, req *ChangeMyWechatRequest, resp *ChangeMyWechatResponse) {
	myID := runtime.EmployeeID(ctx)

	userInfo, err := employee.GetOpsWechatWebLoginUserInfo(ctx, req.Request.Code, &resp.BaseResponse)
	if err != nil {
		return
	}

	EmpWxInfo := &employee.EmployeeWechatInfo{
		NickName:  userInfo.Name,
		AvatarURL: userInfo.AvatarURL,
		UnionID:   userInfo.WechatUnionID,
		WebOpenID: userInfo.WechatWebOpenID,
	}

	_, err = employee.UpdateEmployeeWechatInfo(ctx, myID, EmpWxInfo)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = fmt.Sprintf("换绑微信失败: %v", err)
		return
	}
}

// GetMyAction 获取action_path列表
// @Tags 运营个人中心
// @Summary 获取action_path列表
// @Description 获取action_path列表
// @Router /account.GetMyAction [post]
// @Param user body GetMyActionRequest true "入参"
// @Success 200 {object} GetMyActionResponse "返回值"
func GetMyAction(ctx *gin.Context, req *GetMyActionRequest, resp *GetMyActionResponse) {
	myID := runtime.EmployeeID(ctx)
	emp, err := employee.GetEmployeeByID(ctx, myID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = fmt.Sprintf("查询员工信息失败: %v", err)
		return
	}
	_, _, actionPaths, err := employee.GetEmployeeRbac(ctx, emp)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = fmt.Sprintf("获取权限信息失败: %v", err)
		return
	}

	resp.Response.ActionPaths = actionPaths
}

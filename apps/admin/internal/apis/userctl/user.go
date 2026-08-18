package userctl

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/admin/models/company"
	"github.com/insmtx/corekg/apps/admin/models/user"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// CreateUser 创建用户
// @Tags Admin用户管理
// @Summary 创建用户
// @Description 创建用户
// @Router /admin.CreateUser [post]
// @Param user body CreateUserRequest true "入参"
// @Success 200 {object} CreateUserResponse "返回值"
func CreateUser(ctx *gin.Context, req *CreateUserRequest, resp *CreateUserResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		return
	}
	exist, err := req.Request.IsExist()
	if err != nil {
		logs.ErrorContextf(ctx, "[CreateUser]: IsExist failed, %+v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "查询用户信息失败"
		return
	}
	if exist {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "用户已存在"
		return
	}
	u, err := user.CreateUser(ctx, &req.Request)
	if err != nil {
		logs.ErrorContextf(ctx, "[CreateUser]: CreateUser failed, %+v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "创建用户失败"
		return
	}
	resp.Response.User = *u
}

// ListUser 用户列表
// @Tags Admin用户管理
// @Summary 用户列表
// @Description 用户列表
// @Router /admin.ListUser [post]
// @Param user body ListUserRequest true "入参"
// @Success 200 {object} ListUserResponse "返回值"
func ListUser(ctx *gin.Context, req *ListUserRequest, resp *ListUserResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		return
	}
	err := user.QueryUserList(ctx, &req.Request, &resp.Response)
	if err != nil {
		logs.ErrorContextf(ctx, "[ListUser]: QueryUserList failed, %+v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "查询用户列表失败"
		return
	}
}

// ModifyUser 修改用户
// @Tags Admin用户管理
// @Summary 修改用户
// @Description 修改用户
// @Router /admin.ModifyUser [post]
// @Param user body ModifyUserRequest true "入参"
// @Success 200 {object} ModifyUserResponse "返回值"
func ModifyUser(ctx *gin.Context, req *ModifyUserRequest, resp *ModifyUserResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		return
	}
	exist, err := req.Request.IsExist(req.Request.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "[CreateUser]: IsExist failed, %+v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "查询用户信息失败"
		return
	}
	if exist {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "该手机号或者邮箱已经存在"
		return
	}
	u, err := user.ModifyUser(ctx, req.Request.ID, &req.Request.CreateUserOption)
	if err != nil {
		logs.ErrorContextf(ctx, "[ModifyUser]: ModifyUser failed, %+v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "修改用户失败"
		return
	}
	resp.Response.User = *u
}

// ModifyUserPassword 修改用户密码
// @Tags Admin用户管理
// @Summary 修改用户密码
// @Description 修改用户密码
// @Router /admin.ModifyUserPassword [post]
// @Param user body ModifyUserPasswordRequest true "入参"
// @Success 200 {object} ModifyUserPasswordResponse "返回值"
func ModifyUserPassword(ctx *gin.Context, req *ModifyUserPasswordRequest, resp *ModifyUserPasswordResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		return
	}
	u, err := user.ModifyUserPassword(ctx, req.Request.ID, req.Request.Password)
	if err != nil {
		logs.ErrorContextf(ctx, "[ModifyUserPassword]: ModifyUserPassword failed, %+v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "修改用户密码失败"
		return
	}
	resp.Response.User = *u
}

// GetUserDetail 获取用户详情
// @Tags Admin用户管理
// @Summary 获取用户详情
// @Description 获取用户详情
// @Router /admin.GetUserDetail [post]
// @Param user body GetUserDetailRequest true "入参"
// @Success 200 {object} GetUserDetailResponse "返回值"
func GetUserDetail(ctx *gin.Context, req *GetUserDetailRequest, resp *GetUserDetailResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		return
	}
	detail, err := user.GetUserDetail(req.Request.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetUserDetail]: GetUserDetail failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "获取用户详情失败"
		return
	}

	ret := &company.QueryCompanyEmployeeListResponse{}
	err = company.QueryCompanyEmployeeList(ctx, &apiobj.PageQuery{
		ListAll: true,
		Filters: []apiobj.Filter{
			{
				Field: "user_id",
				Value: []string{strconv.Itoa(int(req.Request.ID))},
			},
		},
	}, ret)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetUserDetail]: QueryCompanyEmployeeList failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "获取用户团队信息失败"
		return
	}
	resp.Response.UserDetail = *detail
	resp.Response.EmployeeList = ret.Data
}

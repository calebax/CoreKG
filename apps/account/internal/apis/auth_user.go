package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/user"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// PersonAuth 个人实名认证
// @Tags User
// @Summary 个人实名认证
// @Description 个人实名认证
// @Router /account.PersonAuth [post]
// @Param user body PersonAuthRequest true "入参"
// @Success 200 {object} PersonAuthResponse "返回值"
func PersonAuth(ctx *gin.Context, req *PersonAuthRequest, resp *PersonAuthResponse) {
	if req.Validity(&resp.BaseResponse); resp.BaseResponse.Code != 0 {
		logs.WarnContextf(ctx, "PersonAuth: validity failed,err = %v", resp.BaseResponse.Message)
		return
	}
	individual, err := user.GetIndividual(req.Request.Uin)
	if err != nil {
		logs.ErrorContextf(ctx, "PersonAuth: GetIndividual failed, %+v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_get_user_info_failed" // 获取用户信息失败
		return
	}
	if individual.RealNameStatus == accounttype.IndividualStatuPassed {
		logs.ErrorContextf(ctx, "PersonAuth: user has authed")
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_user_already_authed" // 用户已实名认证
		return
	}
	individual.IDCard = req.Request.IDCard
	individual.RealName = req.Request.RealName
	individual.RealNameStatus = accounttype.IndividualStatuPending
	err = user.SaveIndividual(ctx, individual)
	if err != nil {
		logs.ErrorContextf(ctx, "PersonAuth: SaveIndividual failed, %+v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_save_user_info_failed" // 保存用户信息失败
		return
	}
}

// ListPersonAuth 获取等待认证的列表
// @Tags User
// @Summary 获取等待认证的列表
// @Description 获取等待认证的列表
// @Router /account.ListPersonAuth [post]
// @Param user body ListPersonAuthRequest true "入参"
// @Success 200 {object} ListPersonAuthResponse "返回值"
func ListPersonAuth(ctx *gin.Context, req *ListPersonAuthRequest, resp *ListPersonAuthResponse) {
	if req.Validity(resp); resp.Message != "" {
		logs.WarnContextf(ctx, "ListPersonAuth: validity failed,err = %v", resp.Message)
		return
	}
	err := user.QueryIndividualsList(req.Request, &resp.Response)
	if err != nil {
		logs.ErrorContextf(ctx, "ListPersonAuth: ListIndividuals failed, %+v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_get_individual_list_failed" // 获取信息失败
		return
	}
}

// ReviewPersonAuth 审阅用户信息
// @Tags User
// @Summary 审阅用户信息
// @Description 审阅用户信息
// @Router /account.ReviewPersonAuth [post]
// @Param user body ReviewPersonAuthRequest true "入参"
// @Success 200 {object} ReviewPersonAuthResponse "返回值"
func ReviewPersonAuth(ctx *gin.Context, req *ReviewPersonAuthRequest, resp *ReviewPersonAuthResponse) {
	if req.Request.Uin == 0 {
		logs.WarnContextf(ctx, "ReviewPersonAuth: invalid params")
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_parameters" // 参数错误
		return
	}
	individual, err := user.GetIndividual(req.Request.Uin)
	if err != nil {
		logs.ErrorContextf(ctx, "ReviewPersonAuth: GetIndividual failed, %+v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_get_user_info_failed" // 获取用户信息失败
		return
	}
	if individual.RealNameStatus == accounttype.IndividualStatuPassed {
		logs.ErrorContextf(ctx, "ReviewPersonAuth: user has authed")
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_user_already_authed" // 用户已实名认证
		return
	}
	if req.Request.Review {
		individual.RealNameStatus = accounttype.IndividualStatuPassed
	} else {
		individual.RealNameStatus = accounttype.IndividualStatuFialed
	}
	err = user.SaveIndividual(ctx, individual)
	if err != nil {
		logs.ErrorContextf(ctx, "ReviewPersonAuth: SaveIndividual failed, %+v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_save_user_info_failed" // 保存用户信息失败
		return
	}
}

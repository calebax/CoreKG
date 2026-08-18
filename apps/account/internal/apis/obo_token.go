package apis

import (
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/user"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/apis/runtime/auth"
	"github.com/ygpkg/yg-go/logs"
)

const oboTokenExpire = 15 * time.Minute
const oboAllowedAudience = "kg_open_coze"

// GetOBOToken 内部服务 OBO 获取调用凭证
// @Tags User
// @Summary 内部服务 OBO 获取调用凭证
// @Description 通过 UIN 和 audience 生成 15 分钟有效期 token
// @Router /account.GetOBOToken [post]
// @Param user body GetOBOTokenRequest true "入参"
// @Success 200 {object} GetOBOTokenResponse "返回值"
func GetOBOToken(ctx *gin.Context, req *GetOBOTokenRequest, resp *GetOBOTokenResponse) {
	logs.InfoContextf(
		ctx,
		"GetOBOToken: request params uin=%d, audience=%s, grant_type=%s, scope=%s",
		req.Request.Uin,
		req.Request.Audience,
		req.Request.GrantType,
		req.Request.Scope,
	)

	if req.Validity(resp); resp.Code != 0 {
		return
	}

	if req.Request.Audience != oboAllowedAudience {
		logs.WarnContextf(ctx, "GetOBOToken: invalid audience=%s", req.Request.Audience)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_parameters"
		return
	}

	uinInfo, err := user.GetUserIdentificationByUIN(ctx, req.Request.Uin)
	if err != nil {
		logs.ErrorContextf(ctx, "GetOBOToken: get uin failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_uin_fetch_failed"
		return
	}
	token := user.GenerateJwtToken(
		ctx,
		req.Request.Uin,
		auth.LoginWayUnknown,
		runtime.GetRealIP(ctx.Request),
		uinInfo.Issuer,
		oboTokenExpire,
	)
	if token == "" {
		logs.ErrorContextf(ctx, "GetOBOToken: generate jwt token failed")
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_jwt_token_generation_failed"
		return
	}
	resp.Response.JwtToken = token
	resp.Response.ExpiredAt = jwt.TimeFunc().Add(oboTokenExpire).Unix()
}

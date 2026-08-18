package loginctl

import (
	"errors"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/mozillazg/go-pinyin"
	"github.com/insmtx/corekg/apps/account/models/user"
	"github.com/insmtx/corekg/apps/admin/models/admintype"
	"github.com/insmtx/corekg/apps/admin/models/employee"
	"github.com/insmtx/corekg/apps/admin/models/login_setting"
	user2 "github.com/insmtx/corekg/apps/admin/models/user"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/pkgs/wx"
	"github.com/xen0n/go-workwx/errcodes"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/apis/runtime/auth"
	"github.com/ygpkg/yg-go/logs"
)

// LoginByPassword 密码登录
// @Tags Admin登录
// @Summary 密码登录
// @Description 登录成功code=0
// @Router /admin.LoginByPassword [post]
// @Param user body LoginPasswordRequest true "入参"
// @Success 200 {object} LoginResponse
func LoginByPassword(ctx *gin.Context, req *LoginPasswordRequest, resp *LoginResponse) {
	if req.Validity(&resp.BaseResponse); resp.BaseResponse.Code != 0 {
		logs.WarnContextf(ctx, "LoginByPassword:validity failed,err = %v", resp.BaseResponse.Message)
		return
	}

	emp, err := employee.GetEmployeeByUsername(ctx, req.Request.Username)
	if err != nil {
		logs.WarnContextf(ctx, "get employee by username failed, %s", err)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "用户或密码错误"
		return
	}
	if emp.Status != admintype.UserStatusNormal {
		logs.WarnContextf(ctx, "user status is not normal, %s", emp.Status)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "该账号无法登录"
		return
	}

	ok := user2.VerifyPassword(ctx, req.Request.Password, string(emp.Password))
	if !ok {
		logs.ErrorContextf(ctx, "LoginByPassword: password not match, %s", err)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "用户或密码错误"
		return
	}

	resp.Response.Issuer = global.IssuerYYGUAdmin
	resp.Response.AllowRegister = false
	resp.Response.LoginWay = 8
	loginSuccess(ctx, emp.Uin, resp, 8)
}

// loginSuccess 登陆成功
func loginSuccess(ctx *gin.Context,
	uin uint,
	resp *LoginResponse,
	loginWay auth.LoginWay) {

	resp.Response.LoginStatus = "success"
	token := user.GenerateJwtToken(ctx, uin, loginWay, runtime.GetRealIP(ctx.Request), resp.Response.Issuer)
	if token == "" {
		logs.ErrorContextf(ctx, "loginSuccess: generate jwt token failed")
		resp.Code = errcode.ErrCode_InternalError
		return
	}
	resp.Response.UserInfo = &user.UserInfo{
		Uin: uin,
	}
	resp.Response.JwtToken = token
	return

}

// LoginThird github app登录
// @Tags User
// @Summary 第三方登录
// @Description 第三方登录
// @Router /account.LoginThird [post]
// @Param user body LoginThirdRequest true "入参"
// @Success 200 {object} LoginResponse "返回值"
func LoginThird(ctx *gin.Context, req *LoginThirdRequest, resp *LoginResponse) {
	if req.Request.Code == "" || req.Request.Way == "" || req.Request.DomainName == "" {
		logs.ErrorContextf(ctx, "LoginThird: invalid code or way")
		runtime.BadRequest(ctx, "invalid code ")
		return
	}

	logs.DebugContextf(ctx, "LoginThird: request %+v", req.Request)
	var (
		userInfo *user.UserInfo
		user     *admintype.Employee
		err      error
		way      auth.LoginWay
	)
	loginSetting, err := login_setting.GetLoginSettingByPath(req.Request.DomainName, "")
	if err != nil {
		logs.ErrorContextf(ctx, "LoginThird: get login setting failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		return
	}

	resp.Response.LoginStatus = "failed"
	switch req.Request.Way {
	case "wechat_web":
		way = auth.LoginWayWxWeb
		userInfo, user, err = loginWechatWeb(ctx, req)
		if err != nil {
			if errors.Is(err, ErrOauthCodeInvalid) {
				logs.WarnContextf(ctx, "loginThird: wechat code invalid,req.code=%s has been used or outdated", req.Request.Code)
				resp.Code = errcode.ErrCode_Unauthorized
				resp.Message = "account_oauth_token_invalid"
				return
			}
		}
		logs.InfoContextf(ctx, "LoginThird: wechat web, %+v, %+v, %v", userInfo, user, err)

	default:
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "invalid way, " + req.Request.Way
		return
	}

	resp.Response.LoginWay = way
	resp.Response.Issuer = loginSetting.Issuer
	resp.Response.AllowRegister = loginSetting.AllowRegister

	if err != nil {
		logs.WarnContextf(ctx, "LoginThird: get user failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		return
	}
	resp.Response.UserInfo = userInfo
	resp.Response.LoginWay = way
	resp.Response.Issuer = loginSetting.Issuer
	resp.Response.AllowRegister = loginSetting.AllowRegister

	loginSuccess(ctx, user.Uin, resp, way)
}

var (
	ErrOauthCodeInvalid = errors.New("oauth code invalid")
)

// loginWechatWeb 微信网页登录
func loginWechatWeb(ctx *gin.Context, req *LoginThirdRequest) (*user.UserInfo, *admintype.Employee, error) {
	loginSetting, err := login_setting.GetLoginSettingByPath(req.Request.DomainName, "")
	if err != nil {
		logs.ErrorContextf(ctx, "loginWechatWeb: get login setting failed, %s", err)
		return nil, nil, err
	}
	weapp, err := wx.GetWechatWebOAuth(ctx, "account", loginSetting.AuthKey)
	if err != nil {
		logs.ErrorContextf(ctx, "loginWechatWeb: get work wechat oauth config failed, %s", err)
		return nil, nil, err
	}

	tkn, err := weapp.GetUserAccessToken(req.Request.Code)
	if err != nil {
		logs.ErrorContextf(ctx, "loginWechatWeb: get user access token failed, %s", err)
		if tkn.ErrCode == errcodes.ErrCode40029 {
			logs.ErrorContextf(ctx, "loginWechatWeb: get user access token failed, errcode=%v, errmsg=%v", tkn.ErrCode, tkn.ErrMsg)
			return nil, nil, ErrOauthCodeInvalid
		}
		return nil, nil, err
	}

	ui, err := weapp.GetUserInfo(tkn.AccessToken, tkn.OpenID, "")
	if err != nil {
		logs.ErrorContextf(ctx, "loginWechatWeb: get user info failed, %s", err)
		return nil, nil, err
	}

	logs.InfoContextf(ctx, "loginWechatWeb: GetUserInfoByCode %+v", ui)

	userInfo := &user.UserInfo{
		WechatUnionID:   ui.Unionid,
		WechatWebOpenID: ui.OpenID,
		Identify:        strings.Join(pinyin.LazyPinyin(ui.Nickname, pinyin.NewArgs()), ""),
		AvatarURL:       ui.HeadImgURL,
		Name:            ui.Nickname,
	}

	emp, err := employee.GetEmployeeByUnionID(userInfo.WechatUnionID)
	if err != nil {
		return nil, nil, err
	}

	return userInfo, emp, err
}

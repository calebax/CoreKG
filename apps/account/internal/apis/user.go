package apis

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/internal/dto/dtouser"
	"github.com/insmtx/corekg/apps/account/models/company"
	"github.com/insmtx/corekg/apps/account/models/employee"
	"github.com/insmtx/corekg/apps/account/models/user"
	"github.com/insmtx/corekg/apps/account/services/svcuser"
	"github.com/insmtx/corekg/pkgs/platform/login_setting"
	"github.com/insmtx/corekg/pkgs/wx"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/i18n"
	"github.com/ygpkg/yg-go/logs"
	"golang.org/x/crypto/bcrypt"
)

// DetailPersonalCenter 获取员工详情信息
// @Tags 用户管理
// @Summary 获取员工详情信息
// @Description 获取员工详情信息
// @Router /account.DetailPersonalCenter [post]
// @Param request body DetailPersonalCenterRequest true "入参"
// @Success 200 {object} DetailPersonalCenterResponse "返回值"
func DetailPersonalCenter(ctx *gin.Context, req *DetailPersonalCenterRequest, resp *DetailPersonalCenterResponse) {
	if req.Validity(resp); resp.Message != "" {
		return
	}
	uin := runtime.Uin(ctx)
	if uin == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_get_user_info_failed" // 获取用户信息失败
		return
	}
	companyID := runtime.CompanyID(ctx)
	userinfo, err := user.GetUserInfo(ctx, uin)
	if err != nil {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_get_user_info_failed" // 获取用户信息失败
		return
	}
	resp.Response.UserInfo = userinfo
	resp.Response.UserInfo.Uin = uin
	if companyID != 0 {
		resp.Response.EmployeeDetail = &employee.EmployeeDetail{}
		err = employee.GetEmployeeDetailByID(ctx, runtime.EmployeeID(ctx), resp.Response.EmployeeDetail)
		if err != nil {
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "account_get_employee_detail_failed" // 获取用户详情失败
			return
		}
		companyinfo, err := company.GetCompany(companyID)
		if err != nil {
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "account_get_company_info_failed" // 获取公司信息失败
			return
		}
		resp.Response.CompanyInfo = companyinfo
	}
}

// ListUin 获取用户全部身份
// @Tags 用户管理
// @Summary 获取用户全部身份
// @Description 获取用户全部身份
// @Router /account.ListUin [post]
// @Param request body ListUinRequest true "入参"
// @Success 200 {object} ListUinResponse "返回值"
func ListUin(ctx *gin.Context, req *ListUinRequest, resp *ListUinResponse) {
	if req.Validity(resp); resp.Message != "" {
		return
	}
	uin, err := user.GetUserIdentificationByUIN(ctx, runtime.Uin(ctx))
	if err != nil {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_get_user_info_failed" // 获取用户信息失败
		return
	}
	uins, err := user.GetUserUinsByUserID(ctx, uin.UserID, uin.Issuer)
	if err != nil {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_get_user_list_failed" // 获取用户列表失败
		return
	}
	// 分类uin
	us, err := getUinType(ctx, uins, uin.Issuer)
	if err != nil {
		logs.ErrorContextf(ctx, "ListUin: get uin type failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_internal_server_error" // 内部服务器错误
		return
	}
	resp.Response.Uin = us
}

// SwitchLogin 切换登录
// @Tags 用户管理
// @Summary 切换登录
// @Description 切换登录
// @Router /account.SwitchLogin [post]
// @Param request body SwitchLoginRequest true "入参"
// @Success 200 {object} SwitchLoginResponse "返回值"
func SwitchLogin(ctx *gin.Context, req *SwitchLoginRequest, resp *SwitchLoginResponse) {
	if req.Validity(resp); resp.Message != "" {
		return
	}
	// 获取uin信息
	uin, err := user.GetUserIdentificationByUIN(ctx, req.Request.Uin)
	if err != nil {
		logs.ErrorContextf(ctx, "ChooseUin: get uin failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_get_user_info_failed" // 获取用户信息失败
		return
	}
	token := user.GenerateJwtToken(ctx, uin.ID, req.Request.LoginWay, runtime.GetRealIP(ctx.Request), uin.Issuer)
	if token == "" {
		logs.ErrorContextf(ctx, "loginSuccess: generate jwt token failed")
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_generate_token_failed" // 生成token失败
		return
	}
	resp.Response.JwtToken = token
}

// UpdatePhoneVerifyCode 更新手机号验证码
// @Tags 用户管理
// @Summary 更新手机号验证码
// @Description 更新手机号验证码
// @Router /account.UpdatePhoneVerifyCode [post]
// @Param request body UpdatePhoneVerifyCodeRequest true "入参"
// @Success 200 {object} UpdatePhoneVerifyCodeResponse "返回值"
func UpdatePhoneVerifyCode(ctx *gin.Context, req *UpdatePhoneVerifyCodeRequest, resp *UpdatePhoneVerifyCodeResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		return
	}
	// 校验手机号是否已经存在
	isExists, err := user.ExistsUserByPhone(ctx, req.Request.Phone)
	if err != nil {
		logs.ErrorContextf(ctx, "get account phone exist failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_get_phone_exist_failed" // 获取手机号存在状态失败
		return
	}
	// 手机号已经存在
	if isExists {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_phone_already_exists" // 手机号已经存在，请更换手机号~
		return
	}
	err = user.CustomerVerifySms(runtime.Uin(ctx), req.Request.Phone, user.UpdatePhonePosition, req.Request.PhoneCode, &resp.BaseResponse)
	if err != nil {
		logs.ErrorContextf(ctx, "update phone verify code failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_update_phone_verify_code_failed" // 更新手机号验证码失败
		return
	}
	uin, err := user.GetUserIdentificationByUIN(ctx, runtime.Uin(ctx))
	if err != nil {
		logs.ErrorContextf(ctx, "get uin failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_get_user_info_failed" // 获取用户信息失败
		return
	}
	err = user.UpdateUserPhoneByID(uin.UserID, req.Request.Phone)
	if err != nil {
		logs.ErrorContextf(ctx, "update user phone failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_update_phone_failed" // 更新手机号失败
		return
	}
}

// UpdatePhoneSendCode 更新手机号发送验证码
// @Tags 用户管理
// @Summary 更新手机号发送验证码
// @Description 更新手机号发送验证码
// @Router /account.UpdatePhoneSendCode [post]
// @Param request body UpdatePhoneSendCodeRequest true "入参"
// @Success 200 {object} UpdatePhoneSendCodeResponse "返回值"
func UpdatePhoneSendCode(ctx *gin.Context, req *UpdatePhoneSendCodeRequest, resp *UpdatePhoneSendCodeResponse) {
	// 校验参数
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		return
	}

	// 校验手机号是否已经存在
	isExists, err := user.ExistsUserByPhone(ctx, req.Request.Phone)
	if err != nil {
		logs.ErrorContextf(ctx, "get account phone exist failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_get_phone_exist_failed" // 获取手机号存在状态失败
		return
	}
	// 手机号已经存在
	if isExists {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_phone_already_exists" // 手机号已经存在，请更换手机号~
		return
	}
	if err = user.CustomerSendSms(ctx, runtime.Uin(ctx), req.Request.Phone, user.UpdatePhonePosition, &resp.BaseResponse); err != nil {
		logs.ErrorContextf(ctx, "send sms failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_send_sms_failed" // 发送验证码失败
		return
	}
}

// BindUserWechat 换绑微信
// @Tags 用户管理
// @Summary 换绑微信
// @Description 换绑微信
// @Router /account.BindUserWechat [post]
// @Param request body BindUserWechatRequest true "入参"
// @Success 200 {object} BindUserWechatResponse "返回值"
func BindUserWechat(ctx *gin.Context, req *BindUserWechatRequest, resp *BindUserWechatResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		return
	}
	uin := runtime.Uin(ctx)

	origin := ctx.Request.Header.Get("Origin")
	lSet, err := login_setting.GetLoginSettingByPath(origin, "")
	if err != nil {
		logs.ErrorContextf(ctx, "GetLoginSettingByPath failed, origin=%v, error=%v", origin, err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "account_get_login_setting_failed")) // 获取LoginSetting失败
		return
	}

	// 获取微信 Web OAuth 配置
	weApp, err := wx.GetWechatWebOAuth(ctx, "account", lSet.AuthKey)
	if err != nil {
		logs.ErrorContextf(ctx, "bindCustomerWechat: get wechat oauth config failed, uin=%d, error=%s", uin, err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_wechat_oauth_config_failed" // 获取微信OAuth配置失败
		return
	}

	// 获取用户 AccessToken
	tkn, err := weApp.GetUserAccessToken(req.Request.Code)
	if err != nil {
		logs.ErrorContextf(ctx, "bindCustomerWechat: get user access token failed, uin=%d, code=%s, error=%s", uin, req.Request.Code, err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_wechat_user_info_fetch_failed" // 获取用户信息失败
		return
	}

	// 获取用户信息
	ui, err := weApp.GetUserInfo(tkn.AccessToken, tkn.OpenID, "")
	if err != nil {
		logs.ErrorContextf(ctx, "bindCustomerWechat: get user info failed, uin=%d, openid=%s, error=%s", uin, tkn.OpenID, err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_wechat_user_fetch_failed" // 获取用户信息失败，请重试
		return
	}

	// 查询当前用户信息
	uinByID, err := user.GetUserIdentificationByUIN(ctx, uin)
	if err != nil {
		logs.ErrorContextf(ctx, "bindCustomerWechat: get user identification by uin failed, uin=%d, error=%s", uin, err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_user_info_expired" // 用户信息异常，请联系客服
		return
	}

	// 检查 UnionID 是否已被其他用户绑定
	isOK, err := user.CheckUnionIDExist(ctx, ui.Unionid)
	if err != nil {
		logs.ErrorContextf(ctx, "bindCustomerWechat: check union id exist failed, uin=%d, unionid=%s, error=%s", uin, ui.Unionid, err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_internal_server_error" // 服务器内部错误，请稍后再试
		return
	}
	if isOK {
		logs.WarnContextf(ctx, "bindCustomerWechat: union id exist, uin=%d, unionid=%s", uin, ui.Unionid)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_wechat_user_already_bound" // 该微信已被其他账号绑定
		return
	}

	// 更新用户 UnionID
	if err = user.UpdateUserWcAndNameByID(uinByID.UserID, &ui); err != nil {
		logs.ErrorContextf(ctx, "bindCustomerWechat: update user unionid failed, uin=%d, userid=%d, unionid=%s, error=%s",
			uin, uinByID.UserID, ui.Unionid, err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_wechat_bind_failed" // 绑定微信失败，请稍后再试
		return
	}
}

// UpdateAccountPassword 修改账号密码
// @Tags 用户管理
// @Summary 修改账号密码
// @Description 修改账号密码
// @Router /account.UpdateAccountPassword [post]
// @Param request body UpdateAccountPasswordRequest true "入参"
// @Success 200 {object} UpdateAccountPasswordResponse "返回值"
func UpdateAccountPassword(ctx *gin.Context, req *UpdateAccountPasswordRequest, resp *UpdateAccountPasswordResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		return
	}
	uin := runtime.Uin(ctx)

	u, err := user.GetUserByUin(ctx, uin)
	if err != nil {
		logs.ErrorContextf(ctx, "update account password: get user by uin failed, uin=%d, error=%s", uin, err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_get_user_failed" // 获取信息失败，请稍后再试
		return
	}

	if u.Password != nil {
		if err = bcrypt.CompareHashAndPassword([]byte(*u.Password), []byte(req.Request.OldPassword)); err != nil {
			logs.ErrorContextf(ctx, "LoginByPassword: password not match, %s", err)
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "account_invalid_password" // 用户或密码错误
			return
		}
	}

	if err = user.UpdateAccountPassword(u.ID, req.Request.NewPassword); err != nil {
		logs.ErrorContextf(ctx, "update account password failed, uin=%d, error=%s", uin, err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_update_password_failed" // 修改密码失败，请稍后再试
		return
	}
}

// UpdateUserInfo 修改用户信息
// @Tags 用户管理
// @Summary 修改用户信息
// @Description 修改用户信息
// @Router /account.UpdateUserInfo [post]
// @Param request body UpdateUserInfoRequest true "入参"
// @Success 200 {object} UpdateUserInfoResponse "返回值"
func UpdateUserInfo(ctx *gin.Context, req *UpdateUserInfoRequest, resp *UpdateUserInfoResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		return
	}
	uin := runtime.Uin(ctx)
	cmpID := runtime.CompanyID(ctx)
	userID, err := user.GetUserByUin(ctx, uin)
	if err != nil {
		logs.ErrorContextf(ctx, "update account password: get user by uin failed, uin=%d, error=%s", uin, err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_get_user_failed" // 获取用户失败
		return
	}
	// 更新用户信息
	err = user.UpdateUserInfo(ctx, userID.ID, cmpID, uin, req.Request.Name, req.Request.AvatarURL, req.Request.Email)
	if err != nil {
		if errors.Is(err, user.ErrNameAlreadyExist) {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "account_name_already_exists" // 该昵称在当前团队中已存在，请修改后重试
			return
		}
		if errors.Is(err, user.ErrEmailAlreadyExist) {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "account_email_already_exists" // 该邮箱在当前团队中已存在，请修改后重试
			return
		}
		logs.ErrorContextf(ctx, "update account password failed, uin=%d, error=%s", uin, err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_update_user_info_failed" // 修改信息失败，请稍后再试
		return
	}
}

// RequestPasswordResetCode 发送重置密码验证码
// @Tags 用户管理
// @Summary 发送重置密码验证码
// @Description 发送重置密码验证码
// @Router /account.RequestPasswordResetCode [post]
// @Param request body dtouser.RequestPasswordResetCodeRequest true "入参"
// @Success 200 {object} dtouser.RequestPasswordResetCodeResponse "返回值"
func RequestPasswordResetCode(ctx *gin.Context, req *dtouser.RequestPasswordResetCodeRequest, resp *dtouser.RequestPasswordResetCodeResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		return
	}
	errResponse := svcuser.RequestPasswordResetCode(ctx, req.Request.Phone, req.Request.Key)
	if errResponse == nil {
		return
	}
	resp.Code = errResponse.Code
	resp.Message = errResponse.Message
}

// ForgotPassword 使用验证码重置密码
// @Tags 用户管理
// @Summary 重置密码
// @Description 使用验证码重置密码
// @Router /account.ForgotPassword [post]
// @Param request body dtouser.ForgotPasswordRequest true "入参"
// @Success 200 {object} dtouser.ForgotPasswordResponse "返回值"
func ForgotPassword(ctx *gin.Context, req *dtouser.ForgotPasswordRequest, resp *dtouser.ForgotPasswordResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		return
	}
	errResponse := svcuser.ForgotPassword(ctx, req.Request.Phone, req.Request.Code, req.Request.Password)
	if errResponse == nil {
		return
	}
	resp.Code = errResponse.Code
	resp.Message = errResponse.Message
}

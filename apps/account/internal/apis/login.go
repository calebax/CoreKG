package apis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/company"
	"github.com/insmtx/corekg/apps/account/models/employee"
	"github.com/insmtx/corekg/apps/account/models/user"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/pkgs/platform/login_setting"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/insmtx/corekg/pkgs/utils/validate"
	"github.com/insmtx/corekg/pkgs/wecoms"
	"github.com/insmtx/corekg/pkgs/wx"
	"github.com/mozillazg/go-pinyin"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/apis/runtime/auth"
	"github.com/ygpkg/yg-go/dbtools/redispool"
	"github.com/ygpkg/yg-go/i18n"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/random"
	"github.com/ygpkg/yg-go/settings"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// LoginThird github app登录
// @Tags User
// @Summary 第三方登录
// @Description 第三方登录
// @Router /account.LoginThird [post]
// @Param user body LoginThirdRequest true "入参"
// @Success 200 {object} LoginThirdResponse "返回值"
func LoginThird(ctx *gin.Context, req *LoginThirdRequest, resp *LoginThirdResponse) {
	if req.Request.Code == "" || req.Request.Way == "" || req.Request.DomainName == "" {
		logs.ErrorContextf(ctx, "LoginThird: invalid code or way")
		runtime.BadRequest(ctx, i18n.T(runtime.GetLanguage(ctx), "account_invalid_code_or_way")) // invalid code
		return
	}

	logs.DebugContextf(ctx, "LoginThird: request %+v", req.Request)
	var (
		userInfo *user.UserInfo
		u        *accounttype.User
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
	case "work_wechat":
		way = auth.LoginWayWorkWechat
		userInfo, u, err = loginWorkWechat(ctx, req)
		logs.InfoContextf(ctx, "LoginThird: work wechat, %+v, %+v, %v", userInfo, u, err)
	case "wechat_web":
		way = auth.LoginWayWxWeb
		userInfo, u, err = loginWechatWeb(ctx, req)
		if err != nil {
			if errors.Is(err, ErrOauthCodeInvalid) {
				logs.WarnContextf(ctx, "loginThird: wechat code invalid,req.code=%s has been used or outdated", req.Request.Code)
				resp.Code = errcode.ErrCode_Unauthorized
				resp.Message = "account_oauth_token_invalid"
				return
			}
		}
		logs.InfoContextf(ctx, "LoginThird: wechat web, %+v, %+v, %v", userInfo, u, err)

	default:
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_way_data" // invalid way
		resp.MessageData = map[string]interface{}{
			"way": req.Request.Way,
		}
		return
	}

	resp.Response.LoginWay = way
	resp.Response.Issuer = loginSetting.Issuer
	resp.Response.AllowRegister = loginSetting.AllowRegister

	if errors.Is(err, gorm.ErrRecordNotFound) {
		if !loginSetting.AllowRegister {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "account_registration_not_allowed" // 没有账号无法注册，请联系管理员
			return
		}
		logs.InfoContextf(ctx, "LoginThird: u not found, %+v", userInfo.Identify)
		saveWaitRegisterUserInfo(ctx, way, userInfo)
		resp.Response.UserInfo = userInfo
		resp.Response.LoginStatus = "register"
		return
	} else if err != nil {
		logs.WarnContextf(ctx, "LoginThird: get u failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		return
	}
	resp.Response.UserInfo = userInfo
	resp.Response.LoginWay = way
	resp.Response.Issuer = loginSetting.Issuer
	resp.Response.AllowRegister = loginSetting.AllowRegister

	loginSuccess(ctx, resp, u, way, req.Request.Option)
}

// LoginByEmail 仅仅为OPO注册使用
// @Tags User
// @Summary 仅仅为OPO注册使用
// @Description 仅仅为OPO注册使用
// @Router /account.LoginByEmail [post]
// @Param user body LoginByEmailRequest true "入参"
// @Success 200 {object} LoginThirdResponse "返回值"
func LoginByEmail(ctx *gin.Context, req *LoginByEmailRequest, resp *LoginThirdResponse) {
	if req.Request.DomainName == "" || req.Request.Email == "" {
		logs.ErrorContextf(ctx, "LoginByEmail: invalid domainname or email")
		runtime.BadRequest(ctx, i18n.T(runtime.GetLanguage(ctx), "account_invalid_domain_or_email")) // invalid domainname or email
		return
	}
	loginSettings, err := login_setting.GetLoginSettingByPath(req.Request.DomainName, "")
	if err != nil {
		logs.ErrorContextf(ctx, "LoginThird: get login setting failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		return
	}

	usr, err := user.GetUserByEmail(req.Request.Email)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			resp.Response.UserInfo = &user.UserInfo{
				Identify: req.Request.Email,
				Name:     req.Request.Email,
				Email:    req.Request.Email,
			}
			// 没找到注册一个公司，然后给个管理员
			err = dbutil.Account().Transaction(func(tx *gorm.DB) error {
				// 创建客户
				usr, err = user.CreateUserByBindLogin(ctx, tx, resp.Response.UserInfo)
				if err != nil {
					logs.ErrorContextf(ctx, "LoginByEmail: create user failed, %+v", err)
					return err
				}
				// 创建公司
				comp, err := company.CreateCompany(ctx, tx, &company.CompanyInfo{
					Name:  req.Request.Email,
					Email: req.Request.Email,
				})
				if err != nil {
					logs.ErrorContextf(ctx, "LoginByEmail: create user failed, %+v", err)
					return err
				}
				// 创建公司uin身份
				uin, err := company.CreateEmployeeIdentification(ctx, tx, usr.ID, comp.ID, loginSettings.Issuer, usr.Name)
				if err != nil {
					logs.ErrorContextf(ctx, "LoginByEmail: create user failed, %+v", err)
					return err
				}
				// 创建个人账户实名认证信息等
				if err := user.CreateIndividual(ctx, tx, usr.ID); err != nil {
					logs.ErrorContextf(ctx, "LoginByEmail: create individual failed, %+v", err)
					return err
				}
				emp := &accounttype.Employee{
					CompanyID: comp.ID,
					UserID:    usr.ID,
					Uin:       uin.ID,
					SysRole:   accounttype.SysRoleSysAdmin,
				}
				// 创建员工
				if err := employee.CreateEmployee(ctx, tx, emp); err != nil {
					logs.ErrorContextf(ctx, "CompanyAuth: CreateEmployee failed, %+v", err)
					return err
				}

				// 所有操作成功，返回 nil 以提交事务
				return nil
			})
			if err != nil {
				runtime.InternalError(ctx, err)
				return
			}
			resp.Response.Issuer = loginSettings.Issuer
			resp.Response.AllowRegister = loginSettings.AllowRegister
			resp.Response.LoginWay = login_setting.LoginWayOpoEmail
			loginSuccess(ctx, resp, usr, login_setting.LoginWayOpoEmail, LoginOption{})
			return
		}
		logs.ErrorContextf(ctx, "LoginByEmail: get user failed, %s", err)
		resp.Code = errcode.ErrCode_BadRequest
		return
	}
	resp.Response.UserInfo = &user.UserInfo{
		Identify:  usr.Identify,
		AvatarURL: usr.AvatarURL,
		Bio:       usr.Bio,
		Name:      usr.Name,
	}
	resp.Response.Issuer = loginSettings.Issuer
	resp.Response.AllowRegister = loginSettings.AllowRegister
	resp.Response.LoginWay = login_setting.LoginWayOpoEmail
	loginSuccess(ctx, resp, usr, login_setting.LoginWayOpoEmail, LoginOption{})
}

// LoginByPassword 微信扫码登陆
// @Tags NewAccount
// @Summary 微信扫码登录
// @Description 登录成功code=0
// @Router /account.LoginByPassword [post]
// @Param user body LoginPasswordRequest true "入参"
// @Success 200 {object} LoginThirdResponse
func LoginByPassword(ctx *gin.Context, req *LoginPasswordRequest, resp *LoginThirdResponse) {
	if req.Request.Password == "" || req.Request.Username == "" {
		logs.ErrorContextf(ctx, "LoginByPassword: invalid password or username")
		runtime.BadRequest(ctx, i18n.T(runtime.GetLanguage(ctx), "account_invalid_password_or_username")) // invalid password or username
		return
	}

	if strings.Contains(req.Request.Username, "@") {
		//? is email
		if err := validate.IsEmail(req.Request.Username); err != nil {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "account_invalid_email"
			return
		}
	} else {
		//? is phone number
		if err := validate.IsPhone(req.Request.Username); err != nil {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "account_invalid_phone_format"
			return
		}
	}

	var (
		delay = 5 * time.Minute
		//origin       = ctx.ClientIP()
		accountRdsKey = user.RedisKeyLoginAccount(req.Request.Username)
		failCount     = 0
		errAPI        error
		err           error
	)
	maxLoginTry, errAPI := settings.GetValue("account", "max_login_try")
	if errAPI != nil {
		logs.ErrorContextf(ctx, "LoginByPassword: get max_login_try error, %s", errAPI)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "account_get_login_try_failed")) // 获取登陆尝试配置失败
		return
	}

	maxTry, errAPI := strconv.Atoi(maxLoginTry)
	if errAPI != nil {
		logs.ErrorContextf(ctx, "LoginByPassword: try conv str[%v] to int failed %v ", maxLoginTry, errAPI)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "account_login_try_conversion_failed")) // 尝试配置转换失败
		return
	}

	logs.DebugContextf(ctx, "LoginByPassword:  accountRdsKey: %s", accountRdsKey)

	// check if origin ip has already called this login api multi times
	if exist, ttl := redispool.IsExistKey(accountRdsKey); !exist {
		logs.DebugContextf(ctx, "LoginByPassword: redis key %s not exist, this is first login in last [%v] seconds", accountRdsKey, delay.Seconds())
		if errAPI = redispool.SetString(accountRdsKey, fmt.Sprintf("%v", 0), delay); errAPI != nil {
			logs.ErrorContextf(ctx, "set redis key [%v] failed", accountRdsKey)
			runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "account_origin_cache_validation_failed")) // 登陆Origin缓存验证失败
			return
		}
	} else {
		countStr, errAPI := redispool.GetString(accountRdsKey)
		if errAPI != nil {
			logs.ErrorContextf(ctx, "LoginByPassword: get redis key [%v] failed", accountRdsKey)
			runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "account_get_origin_cache_failed")) // 获取登陆Origin缓存失败
			return
		}

		logs.DebugContextf(ctx, "Get origin redis key [%v] value[%v]", accountRdsKey, countStr)

		count, errAPI := strconv.Atoi(countStr)
		if errAPI != nil {
			logs.ErrorContextf(ctx, "LoginByPassword: convert redis key [%v] value[%v] to int failed %v ", accountRdsKey, countStr, errAPI)
			runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "account_origin_cache_parsing_failed")) // 登陆Origin缓存解析失败
			return
		}

		if count > maxTry {
			logs.ErrorContextf(ctx, "account %s has already up to MaxLoginTry[%v]\nAPI[loginByPassword] has been banned for %v seconds, debug with key:%v",
				req.Request.Username, maxTry, ttl.Seconds(), accountRdsKey)
			runtime.InternalError(ctx, i18n.TWithData(runtime.GetLanguage(ctx), "account_login_attempts_exceeded", map[string]interface{}{
				"max_try": maxTry,
			})) // 登陆失败次数已达上限
			return
		}
		failCount = count
	}

	defer func() {
		if err != nil {
			if err := redispool.SetString(accountRdsKey, fmt.Sprintf("%v", failCount+1), delay); err != nil {
				logs.ErrorContextf(ctx, "LoginByPassword: set redis key [%v] value[%v] failed", accountRdsKey, failCount+1)
				runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "account_update_origin_cache_failed")) // 更新origin登陆缓存失败
				return
			}
		}
	}()

	setting, err := login_setting.GetLoginSettingByPath(req.Request.DomainName, "")
	if err != nil {
		logs.ErrorContextf(ctx, "LoginThird: get login setting failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_get_login_setting_failed" // 获取登录设置失败
		return
	}

	var (
		usr *accounttype.User
	)
	if strings.Contains(req.Request.Username, "@") {
		usr, err = user.GetUserByEmail(req.Request.Username)
	} else {
		usr, err = user.GetUserByPhone(req.Request.Username)
	}
	if err != nil {
		logs.ErrorContextf(ctx, "LoginByPassword: get user failed, %s", err)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_user_or_password" // 用户或密码错误
		return
	}
	if usr == nil || usr.ID == 0 {
		logs.ErrorContextf(ctx, "LoginByPassword: user not found")
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_user_or_password" // 用户或密码错误
		return
	}
	if usr.Password == nil || *usr.Password == "" {
		logs.ErrorContextf(ctx, "LoginByPassword: user password not set")
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_user_or_password" // 用户或密码错误
		return
	}
	err = bcrypt.CompareHashAndPassword([]byte(*usr.Password), []byte(req.Request.Password))
	if err != nil {
		logs.WarnContextf(ctx, "LoginByPassword: password not match, %s", err)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_user_or_password" // 用户或密码错误
		return
	}

	resp.Response.UserInfo = &user.UserInfo{
		Identify:        usr.Identify,
		AvatarURL:       usr.AvatarURL,
		Bio:             usr.Bio,
		Name:            usr.Name,
		PasswordChanged: usr.PasswordChanged,
		CompanyQuota:    usr.CompanyQuota,
	}
	resp.Response.Issuer = setting.Issuer
	resp.Response.AllowRegister = setting.AllowRegister
	resp.Response.LoginWay = 8
	loginSuccess(ctx, resp, usr, 8, LoginOption{})
}

// ChooseUin 选择身份登录
// @Tags user
// @Summary 选择身份登录
// @Description 选择身份登录
// @Router /account.ChooseUin [post]
// @Param user body ChooseUinRequest true "入参"
// @Success 200 {object} ChooseUinResponse "返回值"
func ChooseUin(ctx *gin.Context, req *ChooseUinRequest, resp *ChooseUinResponse) {
	if req.Request.UinID == 0 || req.Request.LoginWay == 0 {
		logs.ErrorContextf(ctx, "ChooseUin: invalid uin id or login way")
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_uin_or_login_way" // 参数错误
		return
	}
	// 获取用户refresh_token
	refresh_token, err := user.GetRefreshToken(req.Request.UserID)
	if err != nil {
		logs.ErrorContextf(ctx, "ChooseUin: get refresh token failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_refresh_token_fetch_failed" // 初始化用户身份失败
		return
	}
	if refresh_token != req.Request.RefreshToken {
		logs.ErrorContextf(ctx, "ChooseUin: refresh token not match, %s", err)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_refresh_token_mismatch" // 身份校验失败
		return
	}

	// 获取uin信息
	uin, err := user.GetUserIdentificationByUIN(ctx, req.Request.UinID)
	if err != nil {
		logs.ErrorContextf(ctx, "ChooseUin: get uin failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_uin_fetch_failed" // 获取用户信息失败
		return
	}

	token := user.GenerateJwtToken(ctx, req.Request.UinID, req.Request.LoginWay, runtime.GetRealIP(ctx.Request), uin.Issuer)
	if token == "" {
		logs.ErrorContextf(ctx, "loginSuccess: generate jwt token failed")
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_jwt_token_generation_failed" // 切换身份失败
		return
	}

	//updateLastLoginAt 更新用户最后登录时间
	if err := company.UpdateUinLoginTime(ctx, dbutil.Account(), req.Request.UinID); err != nil {
		logs.ErrorContextf(ctx, "ChooseUin: update last login at failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_account_login_pre_fail"
		return
	}

	resp.Response.JwtToken = token
}

// loginWorkWechat 企业微信登录
func loginWorkWechat(ctx *gin.Context, req *LoginThirdRequest) (*user.UserInfo, *accounttype.User, error) {
	weapp, err := wecoms.GetWxCliFromSetting("account", "work_wechat_oauth")
	if err != nil {
		logs.ErrorContextf(ctx, "loginWorkWechat: get work wechat oauth config failed, %s", err)
		return nil, nil, fmt.Errorf("account_wechat_oauth_config_failed") // 获取企业微信OAuth配置失败
	}

	ui, err := weapp.GetUserInfoByCode(req.Request.Code)
	if err != nil {
		logs.ErrorContextf(ctx, "loginWorkWechat: get us info failed, %s", err)
		return nil, nil, fmt.Errorf("account_wechat_user_info_fetch_failed") // 获取用户信息失败
	}

	logs.InfoContextf(ctx, "loginWorkWechat: GetUserInfoByCode %+v", ui)
	wxUser, err := weapp.GetUser(ui.UserID)
	if err != nil {
		logs.ErrorContextf(ctx, "loginWorkWechat: get us failed, %s", err)
		return nil, nil, fmt.Errorf("account_wechat_user_fetch_failed") // 获取用户失败
	}
	logs.InfoContextf(ctx, "loginWorkWechat: GetUser %+v", wxUser)

	userInfo := &user.UserInfo{
		WorkWechatUserID: wxUser.UserID,
		Identify:         wxUser.UserID,
		AvatarURL:        wxUser.AvatarURL,
		Email:            wxUser.Email,
		Phone:            wxUser.Mobile,
		Name:             wxUser.Name,
	}

	us, err := user.GetUserByWorkWechatUserID(userInfo.WorkWechatUserID)
	userInfo.CompanyQuota = us.CompanyQuota
	return userInfo, us, err
}

var (
	ErrOauthCodeInvalid = errors.New("oauth code invalid")
)

// loginWechatWeb 微信网页登录
func loginWechatWeb(ctx *gin.Context, req *LoginThirdRequest) (*user.UserInfo, *accounttype.User, error) {
	login_setting, err := login_setting.GetLoginSettingByPath(req.Request.DomainName, "")
	if err != nil {
		logs.ErrorContextf(ctx, "loginWechatWeb: get login setting failed, %s", err)
		return nil, nil, err
	}
	weapp, err := wx.GetWechatWebOAuth(ctx, "account", login_setting.AuthKey)
	if err != nil {
		logs.ErrorContextf(ctx, "loginWechatWeb: get work wechat oauth config failed, %s", err)
		return nil, nil, err
	}

	tkn, err := weapp.GetUserAccessToken(req.Request.Code)
	if err != nil {
		logs.ErrorContextf(ctx, "loginWechatWeb: get user access token failed, %s", err)
		return nil, nil, ErrOauthCodeInvalid
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

	user, err := user.GetUserByWechatUnionID(userInfo.WechatUnionID)

	if userInfo.Identify == "" {
		userInfo.Identify = random.Alphanum(8)
	}

	if user != nil {
		userInfo.AvatarURL = user.AvatarURL
		userInfo.CompanyQuota = user.CompanyQuota
	}

	return userInfo, user, err
}

// saveWaitRegisterUserInfo 保存等待注册的用户信息
func saveWaitRegisterUserInfo(ctx context.Context, way auth.LoginWay, userInfo *user.UserInfo) error {
	data, err := json.Marshal(userInfo)
	if err != nil {
		logs.ErrorContextf(ctx, "getUserInfoFromGithubUser: marshal failed, %s", err)
		return err
	}
	// 打印序列化结果，用于调试
	logs.InfoContextf(ctx, "Serialized user info: %s", string(data))
	var key string
	switch way {
	case auth.LoginWayGithub:
		key = global.RedisKeyGithubOauthUserInfo(userInfo.GithubID)
	case auth.LoginWayWorkWechat:
		key = global.RedisKeyWorkWechatOauthUserInfo(userInfo.WorkWechatUserID)
	case auth.LoginWayWxWeb:
		key = global.RedisKeyWechatWebOauthUserInfo(userInfo.WechatWebOpenID)
	default:
		logs.ErrorContextf(ctx, "saveWaitRegisterUserInfo: invalid way, %s", way)
		return errors.New("invalid way")
	}
	_, err = redispool.Redis().SetEx(context.Background(), key, data, time.Minute*5).Result()
	if err != nil {
		logs.ErrorContextf(ctx, "saveWaitRegisterUserInfo: set failed, %s", err)
	}
	return nil
}

// getWaitRegisterUserInfo  获取等待注册的用户信息
func getWaitRegisterUserInfo(ctx context.Context, way auth.LoginWay, userInfo *user.UserInfo) (*user.UserInfo, error) {
	var key string
	switch way {
	case auth.LoginWayGithub:
		key = global.RedisKeyGithubOauthUserInfo(userInfo.GithubID)
	case auth.LoginWayWorkWechat:
		key = global.RedisKeyWorkWechatOauthUserInfo(userInfo.WorkWechatUserID)
	case auth.LoginWayWxWeb:
		key = global.RedisKeyWechatWebOauthUserInfo(userInfo.WechatWebOpenID)
	default:
		logs.ErrorContextf(ctx, "getWaitRegisterUserInfo: invalid way, %s", way)
		return nil, errors.New("invalid way")
	}
	data, err := redispool.Redis().Get(context.Background(), key).Result()
	if err != nil {
		logs.ErrorContextf(ctx, "getWaitRegisterUserInfo: get failed, %s", err)
		return nil, err
	}
	ret := &user.UserInfo{}
	err = json.Unmarshal([]byte(data), ret)
	if err != nil {
		logs.ErrorContextf(ctx, "getWaitRegisterUserInfo: unmarshal failed, %s", err)
		return nil, err
	}
	logs.InfoContextf(ctx, "getWaitRegisterUserInfo: %+v", ret)

	return ret, nil
}

// loginSuccess 登陆成功
func loginSuccess(ctx *gin.Context, resp *LoginThirdResponse, user1 *accounttype.User, loginWay auth.LoginWay, opt LoginOption) {

	if user1 == nil {
		logs.ErrorContextf(ctx, "loginSuccess: user is nil")
		resp.Code = errcode.ErrCode_InternalError
		return
	}
	resp.Response.UserID = user1.ID
	resp.Response.LoginStatus = "success"

	if user1.Email != nil {
		resp.Response.UserInfo.Email = *user1.Email
	}

	uins, err := user.GetUserUinsByUserID(ctx, user1.ID, resp.Response.Issuer)
	if err != nil {
		logs.ErrorContextf(ctx, "loginSuccess: get user uins failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		return
	}

	if len(uins) == 0 {
		if !resp.Response.AllowRegister {
			resp.Code = errcode.ErrCode_InternalError
			resp.Message = "account_user_not_registered" // 该用户未注册，请联系管理员
			return
		}
		// if opt.SetDefaultUin {
		// 	if err = dbutil.Account().Transaction(func(tx *gorm.DB) error {
		// 		// 创建公司
		// 		comp, err := company.CreateCompany(ctx, tx, &company.CompanyInfo{
		// 			Name: fmt.Sprintf("%v_%v_%v", resp.Response.UserInfo.Name, resp.Response.UserInfo.Identify, random.String(6)),
		// 		})
		// 		if err != nil {
		// 			logs.Errorf("loginSuccess: CreateCompany failed, %+v", err)
		// 			return err
		// 		}
		// 		// 创建公司uin身份
		// 		uin, err := company.CreateEmployeeIdentification(ctx, tx, user1.ID, comp.ID, resp.Response.Issuer, resp.Response.UserInfo.Name)
		// 		if err != nil {
		// 			logs.Errorf("loginSuccess: CreateEmployeeIdentification failed, %+v", err)
		// 			return err
		// 		}
		// 		emp := &accounttype.Employee{
		// 			CompanyID: comp.ID,
		// 			UserID:    resp.Response.UserInfo.ID,
		// 			Uin:       uin.ID,
		// 			SysRole:   accounttype.SysRoleSysAdmin,
		// 		}
		// 		// 创建员工
		// 		if err := employee.CreateEmployee(ctx, tx, emp); err != nil {
		// 			logs.Errorf("CompanyAuth: CreateEmployee failed, %+v", err)
		// 			return err
		// 		}
		// 		uins = append(uins, uin)
		// 		return nil
		// 	}); err != nil {
		// 		logs.ErrorContextf(ctx, "loginSuccess: create user uins failed, %+v", err)
		// 		resp.Code = errcode.ErrCode_InternalError
		// 		return
		// 	}
		// } else {
		// 	resp.Code = errcode.ErrCode_InternalError
		// 	resp.Message = "account_user_not_registered"
		// 	return
		// }
	}

	// 分类uin
	luin, err := getUinType(ctx, uins, resp.Response.Issuer)
	if err != nil {
		logs.ErrorContextf(ctx, "loginSuccess: get uin type failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_get_uin_type_failed" // 获取分类uin失败
		return
	}
	resp.Response.Uin = luin
	if len(luin) == 1 {
		// if len(luin) == 1 && opt.SingleUinDirective {
		token := user.GenerateJwtToken(ctx, uins[0].ID, loginWay, runtime.GetRealIP(ctx.Request), resp.Response.Issuer)
		if token == "" {
			logs.ErrorContextf(ctx, "loginSuccess: generate jwt token failed")
			resp.Code = errcode.ErrCode_InternalError
			return
		}
		resp.Response.UserInfo.Uin = uins[0].ID
		resp.Response.JwtToken = token
		// return
	}

	refresh, err := user.GenerateRefreshToken(ctx, user1.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "loginSuccess: generate refresh token failed, %s", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_generate_refresh_token_failed"
		return
	}
	resp.Response.RefreshToken = refresh
}

// getUinType 登陆用户信息
func getUinType(ctx context.Context, uins []*accounttype.UserIdentification, issuer string) ([]*LoginUin, error) {
	var loginUin []*LoginUin
	if len(uins) == 0 {
		return loginUin, nil
	}
	// 收集需要查询的公司 ID 和用户 ID
	var uinids []uint
	var userID uint
	for _, uin := range uins {
		if uin.SubjectType == accounttype.SubjectTypeCompany && uin.Issuer == issuer {
			uinids = append(uinids, uin.ID)
		} else if uin.SubjectType == accounttype.SubjectTypeIndividual {
			userID = uin.UserID
		}
	}
	emps, err := employee.GetCompanyEmployeeInfo(uinids)
	if err != nil {
		logs.ErrorContextf(ctx, "getUinType: get company employee info failed, %s", err)
		return nil, err
	}
	// 按 CompanyID 分组建立映射
	employeeMap := make(map[uint]*employee.CompanyEmployeeInfo)
	for _, emp := range emps {
		employeeMap[emp.CompanyID] = emp
	}
	for _, uin := range uins {
		if uin.SubjectType == accounttype.SubjectTypeCompany {
			emp, exists := employeeMap[uin.SubjectID]
			if !exists || emp == nil {
				logs.ErrorContextf(ctx, "getUinType: company info not found for SubjectID %d ,%+v", uin.SubjectID, employeeMap)
				continue
			}
			loginUin = append(loginUin, &LoginUin{
				Uin:           *uin,
				CompanyLogo:   emp.CompanyLogo,
				CompanyName:   emp.CompanyName,
				Role:          emp.SysRole,
				CompanyStatus: emp.CompanyStatus,
				LastLoginAt:   uin.LastLoginAt,
				CompanyUserID: emp.CompanyUserID,
			})
		} else if uin.SubjectType == accounttype.SubjectTypeIndividual {
			// ! current branch is not in expected, should be removed in future
			user, err := user.GetUserByID(userID)
			if err != nil {
				return nil, err
			}
			loginUin = append(loginUin, &LoginUin{
				Uin:  *uin,
				Name: user.Name,
			})
		}
	}
	return loginUin, nil
}

// CheckPassword 密码校验
// @Tags user
// @Summary 密码校验
// @Description 密码校验
// @Router /account.CheckPassword [post]
// @Param user body CheckPassWordRequest true "入参"
// @Success 200 {object} apiobj.BaseResponse "返回值"
func CheckPassword(ctx *gin.Context, req *CheckPassWordRequest, resp *apiobj.BaseResponse) {
	uin := runtime.Uin(ctx)
	usr, err := user.GetUserByUin(ctx, uin)
	if err != nil {
		logs.ErrorContextf(ctx, "CheckPassword: GetUserByUin(%v) failed, %v", uin, err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "account_get_user_failed")) // 获取用户失败
		return
	}

	if usr.Password == nil && len(req.Request.Password) == 0 {
		logs.WarnContextf(ctx, "CheckPassword: uin[%v] user_id[%v] has no passwd", uin, usr.ID)
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(*usr.Password), []byte(req.Request.Password)); err != nil {
		logs.WarnContextf(ctx, "CheckPassword: compare password failed, %v", err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "account_invalid_password")) // 密码错误
		return
	}

	return
}

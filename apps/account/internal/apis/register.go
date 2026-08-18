package apis

import (
	"fmt"
	"math/rand"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/user"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/insmtx/corekg/pkgs/utils/validate"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime/auth"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

// RegisterThird 第三方注册请求
// @Tags User
// @Summary 注册
// @Description 注册
// @Router /account.RegisterThird [post]
// @Param user body RegisterThirdRequest true "入参"
// @Success 200 {object} LoginThirdResponse "返回值"
func RegisterThird(ctx *gin.Context, req *RegisterThirdRequest, resp *LoginThirdResponse) {
	if req.Request.Way == "" {
		logs.ErrorContextf(ctx, "LoginThird: invalid way")
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_way" // 无效的登录方式
		return
	}

	var way auth.LoginWay
	switch req.Request.Way {
	case "github":
		way = auth.LoginWayGithub
	case "work_wechat":
		way = auth.LoginWayWorkWechat
	case "wechat_web":
		way = auth.LoginWayWxWeb
	default:
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_way_data" // 无效的登录方式: {{.way}}
		resp.MessageData = map[string]interface{}{
			"way": req.Request.Way,
		}
		return
	}

	oriUserinfo, err := getWaitRegisterUserInfo(ctx, way, req.Request.UserInfo)

	if err := validate.IsUsername(oriUserinfo.Identify); err != nil {
		logs.ErrorContextf(ctx, "LoginThird: invalid identify, %+v", err)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_identify" // 无效的用户标识
		resp.MessageData = map[string]interface{}{
			"error": err.Error(),
		}
		return
	}
	if len(oriUserinfo.Name) < 1 || len([]rune(oriUserinfo.Name)) > 32 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_name" // 无效的用户名
		return
	}

	if err != nil {
		logs.ErrorContextf(ctx, "RegisterThird: getWaitRegisterUserInfo failed, %+v", err)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_user_info_expired" // 用户信息不存在或者授权已经过期，请重新授权
		return
	}
	// 检查用户标识是否存在
	exi, err := user.ExistsUserByIIdentify(ctx, oriUserinfo.Identify)
	if err != nil {
		logs.ErrorContextf(ctx, "ExistsUserByIIdentify: exists user failed, %+v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_internal_error" // 内部服务器错误
		return
	}
	if exi {
		// 生成新的 identify，追加一位随机数
		random := rand.Intn(10) // 生成 0 到 9 的随机数
		oriUserinfo.Identify = fmt.Sprintf("%s%d", oriUserinfo.Identify, random)
	}

	//exist, err := company.ExistCompanyByName(req.Request.CompanyInfo.Name)
	//if err != nil {
	//	logs.Errorf("ExistsCompanyByName: exists company failed, %+v", err)
	//	runtime.InternalError(ctx, err)
	//	return
	//}
	//if exist {
	//	logs.WarnContextf(ctx, "[account][user] exist company named as %+v", req.Request.CompanyInfo.Name)
	//	runtime.BadRequest(ctx, err)
	//	return
	//}

	var cus *accounttype.User
	err = dbutil.Account().Transaction(func(tx *gorm.DB) error {
		// 创建客户
		// var err error
		usr, err := user.CreateUserByBindLogin(ctx, tx, req.Request.UserInfo)
		if err != nil {
			logs.ErrorContextf(ctx, "LoginByEmail: create user failed, %+v", err)
			return err
		}

		cus = &accounttype.User{
			Model: gorm.Model{
				ID: usr.ID,
			},
			Identify:     usr.Identify,
			Name:         usr.Name,
			Bio:          usr.Bio,
			AvatarURL:    usr.AvatarURL,
			Email:        usr.Email,
			CompanyQuota: usr.CompanyQuota,
		}
		return nil
	})

	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_internal_error" // 内部服务器错误
		return
	}
	// 只有允许注册逻辑才能执行到这，结果中直接设置true
	oriUserinfo.CompanyQuota = cus.CompanyQuota
	resp.Response.AllowRegister = true
	resp.Response.Issuer = req.Request.Issuer
	resp.Response.UserInfo = oriUserinfo
	resp.Response.LoginWay = way
	loginSuccess(ctx, resp, cus, way, req.Request.Option)
}

// CheckUserIdentifyExist 检查用户标识是否存在
// @Tags User
// @Summary 检查用户标识是否存在
// @Description 检查用户标识是否存在
// @Router /account.CheckUserIdentifyExist [post]
// @Param user body CheckUserIdentifyExistRequest true "入参"
// @Success 200 {object} CheckUserIdentifyExistResponse "返回值"
func CheckUserIdentifyExist(ctx *gin.Context, req *CheckUserIdentifyExistRequest, resp *CheckUserIdentifyExistResponse) {
	if err := validate.IsUsername(req.Request.Identify); err != nil {
		logs.ErrorContextf(ctx, "CheckUserIdentifyExist: invalid identify, %+v", err)
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "account_invalid_identify" // 无效的用户标识
		resp.MessageData = map[string]interface{}{
			"error": err.Error(),
		}
		return
	}

	exi, err := user.ExistsUserByIIdentify(ctx, req.Request.Identify)
	if err != nil {
		logs.ErrorContextf(ctx, "CheckUserIdentifyExist: exists user failed, %+v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_internal_error" // 内部服务器错误
		return
	}
	resp.Response.Identify = req.Request.Identify
	resp.Response.Exist = exi
}

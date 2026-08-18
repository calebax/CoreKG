package apis

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/admin/models/login_setting"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// GetLoginSetting 获取登录页配置
// @Tags 登录页配置管理
// @Summary 获取登录页配置
// @Description 获取登录页配置
// @Router /account.GetLoginSetting [post]
// @Param user body GetLoginSettingRequest true "入参"
// @Success 200 {object} GetLoginSettingResponse "返回值"
func GetLoginSetting(ctx *gin.Context, req *GetLoginSettingRequest, resp *GetLoginSettingResponse) {
	if req.Validity(ctx, &resp.BaseResponse); resp.Code != 0 {
		return
	}
	req.Request.Path = strings.TrimSuffix(req.Request.Path, "/")
	info, err := login_setting.GetLoginSettingByPath(req.Request.DomainName, req.Request.Path)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_get_login_setting_failed_data" // 获取登录配置失败
		resp.MessageData = map[string]interface{}{
			"error": err.Error(),
		}
		logs.ErrorContextf(ctx, "[account] Failed to GetLoginSettingByPath(domain:%v|path:%v) : %v", req.Request.DomainName, req.Request.Path, err)
		return
	}
	resp.Response = settingData{
		Title:         info.Title,
		ImageUrl:      info.ImageUrl,
		AllowRegister: info.AllowRegister,
		Issuer:        info.Issuer,
		AuthKey:       info.AuthKey,
		LoginURL:      info.LoginURL,
		WeChat: struct {
			Enable bool   `json:"enable"`
			AppID  string `json:"appid"`
		}{
			Enable: info.IsEnableWeChat,
			AppID:  info.AppID,
		},
		WeChatCom: struct {
			Enable   bool   `json:"enable"`
			AppIDCom string `json:"appid_com"`
			AgentID  string `json:"agentid"`
		}{
			Enable:   info.IsEnableWeChatCom,
			AppIDCom: info.AppIDCom,
			AgentID:  info.AgentID,
		},
		Phone: struct {
			Enable bool `json:"enable"`
		}{
			Enable: info.IsEnablePhone,
		},
		Emali: struct {
			Enable bool `json:"enable"`
		}{
			Enable: info.IsEnableEmail,
		},
		Password: struct {
			Enable bool `json:"enable"`
		}{
			Enable: info.IsEnablePassword,
		},
	}
}

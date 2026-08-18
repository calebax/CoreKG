package apis

import (
	"fmt"
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
// @Router /admin.GetLoginSetting [post]
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
		logs.ErrorContextf(ctx, "[admin]Failed to GetLoginSettingByPath(domain:%v|path:%v) : %v", req.Request.DomainName, req.Request.Path, err)
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

// ListLoginSetting 登录页配置列表
// @Tags 登录页配置管理
// @Summary 登录页配置列表
// @Description 登录页配置列表
// @Router /admin.ListLoginSetting [post]
// @Param user body ListLoginSettingRequest true "入参"
// @Success 200 {object} ListLoginSettingResponse "返回值"
func ListLoginSetting(ctx *gin.Context, req *ListLoginSettingRequest, resp *ListLoginSettingResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		return
	}

	err := login_setting.QueryLoginSettingList(ctx, req.Request, &resp.Response)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = fmt.Sprintf("获取配置列表失败: %v", err)
		logs.ErrorContextf(ctx, "Failed to obtain configuration list: %v", err)
		return
	}
}

// CreateLoginSetting 创建登录页配置
// @Tags 登录页配置管理
// @Summary 创建登录页配置
// @Description 创建登录页配置
// @Router /admin.CreateLoginSetting [post]
// @Param user body CreateLoginSettingRequest true "入参"
// @Success 200 {object} CreateLoginSettingResponse "返回值"
func CreateLoginSetting(ctx *gin.Context, req *CreateLoginSettingRequest, resp *CreateLoginSettingResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		return
	}

	err := login_setting.CreateLoginSetting(req.Request)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = fmt.Sprintf("创建登录配置失败: %v", err)
		logs.ErrorContextf(ctx, "Failed to create login configuration: %v", err)
		return
	}
}

// UpdateLoginSetting 修改登录页配置
// @Tags 登录页配置管理
// @Summary 修改登录页配置
// @Description 修改登录页配置
// @Router /admin.UpdateLoginSetting [post]
// @Param user body UpdateLoginSettingRequest true "入参"
// @Success 200 {object} UpdateLoginSettingResponse "返回值"
func UpdateLoginSetting(ctx *gin.Context, req *UpdateLoginSettingRequest, resp *UpdateLoginSettingResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		return
	}

	err := login_setting.UpdateLoginSetting(req.Request)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = fmt.Sprintf("修改登录配置失败: %v", err)
		logs.ErrorContextf(ctx, "Failed to modify login configuration: %v", err)
		return
	}
}

// DeleteLoginSetting 删除登录页配置
// @Tags 登录页配置管理
// @Summary 删除登录页配置
// @Description 删除登录页配置
// @Router /admin.DeleteLoginSetting [post]
// @Param user body DeleteLoginSettingRequest true "入参"
// @Success 200 {object} DeleteLoginSettingResponse "返回值"
func DeleteLoginSetting(ctx *gin.Context, req *DeleteLoginSettingRequest, resp *DeleteLoginSettingResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		return
	}

	err := login_setting.DeleteLoginSetting(req.Request)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = fmt.Sprintf("删除登录配置失败: %v", err)
		logs.ErrorContextf(ctx, "Failed to delete login configuration : %v", err)
		return
	}
}

// GetLoginSettingByID 根据ID获取登录页配置
// @Tags 登录页配置管理
// @Summary 根据ID获取登录页配置
// @Description 根据ID获取登录页配置
// @Router /admin.GetLoginSettingByID [post]
// @Param user body GetLoginSettingByIDRequest true "入参"
// @Success 200 {object} GetLoginSettingByIDResponse "返回值"
func GetLoginSettingByID(ctx *gin.Context, req *GetLoginSettingByIDRequest, resp *GetLoginSettingByIDResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		return
	}

	info, err := login_setting.GetLoginSettingByID(req.Request.ID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "account_get_login_setting_failed_data" // 获取登录配置失败
		resp.MessageData = map[string]interface{}{
			"error": err.Error(),
		}
		logs.ErrorContextf(ctx, "Failed to GetLoginSettingByID(id:%v) : %v", req.Request.ID, err)
		return
	}
	resp.Response.LoginSetting = info
}

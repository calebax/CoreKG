package apis

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/admin/models/employee"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/settings"
)

// ListSetting 配置列表
// @Tags 配置管理
// @Summary 配置列表
// @Description 配置列表
// @Router /admin.ListSetting [post]
// @Param user body ListSettingRequest true "入参"
// @Success 200 {object} ListSettingResponse "返回值"
func ListSetting(ctx *gin.Context, req *ListSettingRequest, resp *ListSettingResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		return
	}

	err := employee.QuerySettingList(ctx, req.Request, &resp.Response)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = fmt.Sprintf("获取配置列表失败: %v", err)
		return
	}
}

// UpdateSetting 修改配置
// @Tags 配置管理
// @Summary 更改配置
// @Description 更改配置
// @Router /admin.UpdateSetting [post]
// @Param user body UpdateSettingRequest true "入参"
// @Success 200 {object} UpdateSettingResponse "返回值"
func UpdateSetting(ctx *gin.Context, req *UpdateSettingRequest, resp *UpdateSettingResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "参数错误"
		return
	}
	set, err := employee.GetSettingByID(req.Request.ID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = fmt.Sprintf("获取配置失败：%v", err)
		return
	}
	set.Name = req.Request.Name
	set.Describe = req.Request.Describe
	set.Value = req.Request.Value
	set.ValueType = req.Request.ValueType
	set.Default = req.Request.Default
	if err := settings.Updates(set); err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = fmt.Sprintf("修改配置失败：%v", err)
		return
	}
}

// CreateSetting 新增配置
// @Tags 配置管理
// @Summary 新增配置
// @Description 新增配置
// @Router /admin.CreateSetting [post]
// @Param user body CreateSettingRequest true "入参"
// @Success 200 {object} CreateSettingResponse "返回值"
func CreateSetting(ctx *gin.Context, req *CreateSettingRequest, resp *CreateSettingResponse) {
	if req.Validity(&resp.BaseResponse); resp.Code != 0 {
		return
	}

	if err := employee.CreateSetting(&req.Request); err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = fmt.Sprintf("创建配置失败: %v", err)
		return
	}
}

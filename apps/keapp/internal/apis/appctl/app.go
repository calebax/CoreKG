package appctl

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/keapp/models/apptype"
	"github.com/insmtx/corekg/apps/keapp/services/svcapp"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
)

func CreateApplication(ctx *gin.Context, req *CreateAppRequest, resp *CreateAppResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}

	config := apptype.AppConfig{
		Type:   req.Request.Type,
		Config: req.Request.Config,
	}
	result, err := svcapp.CreateApplication(ctx, &svcapp.CreateApplicationRequest{
		Uin:       runtime.Uin(ctx),
		CompanyID: runtime.CompanyID(ctx),
		Name:      req.Request.Name,
		Type:      req.Request.Type,
		Desc:      req.Request.Description,
		Color:     req.Request.Color,
		Config:    config,
	})
	if err != nil {
		if errors.Is(err, svcapp.ErrAppNameExists) {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "keapp_name_exists"
			return
		}
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "keapp_create_failed"
		return
	}
	resp.Response.AppID = result.AppID
}

func ListApplications(ctx *gin.Context, req *ListAppRequest, resp *ListAppResponse) {
	result, err := svcapp.ListApplications(ctx, &svcapp.ListApplicationsRequest{
		Uin:       runtime.Uin(ctx),
		CompanyID: runtime.CompanyID(ctx),
		NameLike:  req.Request.NameLike,
		Limit:     req.Request.Limit,
		Offset:    req.Request.Offset,
	})
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "keapp_list_failed"
		return
	}
	resp.Response.Items = result.Items
	resp.Response.Total = result.Total
}

func GetApplication(ctx *gin.Context, req *GetAppRequest, resp *GetAppResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}

	entity, err := svcapp.GetApplication(ctx, &svcapp.GetApplicationRequest{
		AppID:     req.Request.AppID,
		Uin:       runtime.Uin(ctx),
		CompanyID: runtime.CompanyID(ctx),
	})
	if err != nil {
		if errors.Is(err, svcapp.ErrAppNotFound) {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "keapp_not_found"
			return
		}
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "keapp_get_failed"
		return
	}
	resp.Response.App = entity
}

func UpdateApplication(ctx *gin.Context, req *UpdateAppRequest, resp *UpdateAppResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}

	var configPtr *apptype.AppConfig
	if req.Request.Config != nil {
		configPtr = &apptype.AppConfig{Config: *req.Request.Config}
	}
	err := svcapp.UpdateApplication(ctx, &svcapp.UpdateApplicationRequest{
		AppID:     req.Request.AppID,
		Uin:       runtime.Uin(ctx),
		CompanyID: runtime.CompanyID(ctx),
		Name:      req.Request.Name,
		Desc:      req.Request.Description,
		Color:     req.Request.Color,
		Config:    configPtr,
	})
	if err != nil {
		if errors.Is(err, svcapp.ErrAppNameExists) {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "keapp_name_exists"
			return
		}
		if errors.Is(err, svcapp.ErrAppNotFound) {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "keapp_not_found"
			return
		}
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "keapp_update_failed"
		return
	}
}

func DeleteApplication(ctx *gin.Context, req *DeleteAppRequest, resp *DeleteAppResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}

	err := svcapp.DeleteApplication(ctx, req.Request.AppID, runtime.Uin(ctx), runtime.CompanyID(ctx))
	if err != nil {
		if errors.Is(err, svcapp.ErrAppNotFound) {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "keapp_not_found"
			return
		}
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "keapp_delete_failed"
		return
	}
}

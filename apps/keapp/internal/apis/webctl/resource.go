package webctl

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/keapp/services/svcweb"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/apis/runtime"
)

func ListResources(ctx *gin.Context, req *ListResourcesRequest, resp *ListResourcesResponse) {
	items, total, err := svcweb.ListWebResources(ctx, req.Request.AppID, req.Request.Limit, req.Request.Offset)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "keapp_list_resources_failed"
		return
	}
	resp.Response.Items = items
	resp.Response.Total = total
}

func GetResource(ctx *gin.Context, req *GetResourceRequest, resp *GetResourceResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}

	resource, err := svcweb.GetWebResource(ctx, req.Request.ID)
	if err != nil {
		if errors.Is(err, svcweb.ErrResourceNotFound) {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "keapp_resource_not_found"
			return
		}
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "keapp_get_resource_failed"
		return
	}
	resp.Response.Resource = resource
}

func DeleteResource(ctx *gin.Context, req *DeleteResourceRequest, resp *DeleteResourceResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}

	err := svcweb.DeleteWebResource(ctx, req.Request.ID)
	if err != nil {
		if errors.Is(err, svcweb.ErrResourceNotFound) {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "keapp_resource_not_found"
			return
		}
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "keapp_delete_resource_failed"
		return
	}
}

func RecrawlResource(ctx *gin.Context, req *RecrawlResourceRequest, resp *RecrawlResourceResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}

	taskID, err := svcweb.RecrawlResource(ctx, req.Request.ResourceID, runtime.Uin(ctx))
	if err != nil {
		if errors.Is(err, svcweb.ErrResourceNotFound) {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "keapp_resource_not_found"
			return
		}
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "keapp_recrawl_failed"
		return
	}
	resp.Response.TaskID = taskID
}

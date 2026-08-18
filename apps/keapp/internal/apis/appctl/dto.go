package appctl

import (
	"encoding/json"

	"github.com/insmtx/corekg/apps/keapp/models/apptype"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

type CreateAppRequest struct {
	apiobj.BaseRequest
	Request struct {
		Name        string                  `json:"name"`
		Type        apptype.AppTemplateType `json:"type"`
		Description string                  `json:"description"`
		Color       string                  `json:"color"`
		Config      json.RawMessage         `json:"config"`
	} `json:"request"`
}

func (r *CreateAppRequest) Validity(resp *CreateAppResponse) {
	if r.Request.Name == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "keapp_name_required"
		return
	}
	if r.Request.Type == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "keapp_type_required"
		return
	}
}

type CreateAppResponse struct {
	apiobj.BaseResponse
	Response struct {
		AppID uint `json:"app_id"`
	} `json:"response"`
}

type GetAppRequest struct {
	apiobj.BaseRequest
	Request struct {
		AppID uint `json:"app_id"`
	} `json:"request"`
}

func (r *GetAppRequest) Validity(resp *GetAppResponse) {
	if r.Request.AppID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "keapp_id_required"
	}
}

type GetAppResponse struct {
	apiobj.BaseResponse
	Response struct {
		App *apptype.KeApplication `json:"app"`
	} `json:"response"`
}

type ListAppRequest struct {
	apiobj.BaseRequest
	Request struct {
		NameLike string `json:"name_like"`
		Limit    int    `json:"limit"`
		Offset   int    `json:"offset"`
	} `json:"request"`
}

type ListAppResponse struct {
	apiobj.BaseResponse
	Response struct {
		Items apptype.KeApplicationList `json:"items"`
		Total int64                     `json:"total"`
	} `json:"response"`
}

type UpdateAppRequest struct {
	apiobj.BaseRequest
	Request struct {
		AppID       uint             `json:"app_id"`
		Name        *string          `json:"name,omitempty"`
		Description *string          `json:"description,omitempty"`
		Color       *string          `json:"color,omitempty"`
		Config      *json.RawMessage `json:"config,omitempty"`
	} `json:"request"`
}

func (r *UpdateAppRequest) Validity(resp *UpdateAppResponse) {
	if r.Request.AppID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "keapp_id_required"
	}
}

type UpdateAppResponse struct {
	apiobj.BaseResponse
}

type DeleteAppRequest struct {
	apiobj.BaseRequest
	Request struct {
		AppID uint `json:"app_id"`
	} `json:"request"`
}

func (r *DeleteAppRequest) Validity(resp *DeleteAppResponse) {
	if r.Request.AppID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "keapp_id_required"
	}
}

type DeleteAppResponse struct {
	apiobj.BaseResponse
}

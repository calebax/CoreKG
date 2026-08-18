package webctl

import (
	"github.com/insmtx/corekg/apps/keapp/models/web"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/errcode"
)

type ListResourcesRequest struct {
	apiobj.BaseRequest
	Request struct {
		AppID  uint `json:"app_id"`
		Limit  int  `json:"limit"`
		Offset int  `json:"offset"`
	} `json:"request"`
}

type ListResourcesResponse struct {
	apiobj.BaseResponse
	Response struct {
		Items []*web.KeWebResource `json:"items"`
		Total int64                `json:"total"`
	} `json:"response"`
}

type GetResourceRequest struct {
	apiobj.BaseRequest
	Request struct {
		ID uint `json:"id"`
	} `json:"request"`
}

func (r *GetResourceRequest) Validity(resp *GetResourceResponse) {
	if r.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "keapp_resource_id_required"
	}
}

type GetResourceResponse struct {
	apiobj.BaseResponse
	Response struct {
		Resource *web.KeWebResource `json:"resource"`
	} `json:"response"`
}

type DeleteResourceRequest struct {
	apiobj.BaseRequest
	Request struct {
		ID uint `json:"id"`
	} `json:"request"`
}

func (r *DeleteResourceRequest) Validity(resp *DeleteResourceResponse) {
	if r.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "keapp_resource_id_required"
	}
}

type DeleteResourceResponse struct {
	apiobj.BaseResponse
}

type RecrawlResourceRequest struct {
	apiobj.BaseRequest
	Request struct {
		ResourceID uint `json:"resource_id"`
	} `json:"request"`
}

func (r *RecrawlResourceRequest) Validity(resp *RecrawlResourceResponse) {
	if r.Request.ResourceID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "keapp_resource_id_required"
	}
}

type RecrawlResourceResponse struct {
	apiobj.BaseResponse
	Response struct {
		TaskID uint `json:"task_id"`
	} `json:"response"`
}

type AddCrawlRuleRequest struct {
	apiobj.BaseRequest
	Request struct {
		AppID    uint              `json:"app_id"`
		RuleType web.CrawlRuleType `json:"rule_type"`
		Pattern  string            `json:"pattern"`
		Priority int               `json:"priority"`
	} `json:"request"`
}

func (r *AddCrawlRuleRequest) Validity(resp *AddCrawlRuleResponse) {
	if r.Request.AppID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "keapp_app_id_required"
		return
	}
	if r.Request.Pattern == "" {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "keapp_pattern_required"
	}
}

type AddCrawlRuleResponse struct {
	apiobj.BaseResponse
}

type ListCrawlRulesRequest struct {
	apiobj.BaseRequest
	Request struct {
		AppID uint `json:"app_id"`
	} `json:"request"`
}

type ListCrawlRulesResponse struct {
	apiobj.BaseResponse
	Response struct {
		Items []*web.KeWebCrawlRule `json:"items"`
	} `json:"response"`
}

type UpdateCrawlRuleRequest struct {
	apiobj.BaseRequest
	Request struct {
		ID       uint               `json:"id"`
		RuleType *web.CrawlRuleType `json:"rule_type,omitempty"`
		Pattern  *string            `json:"pattern,omitempty"`
		Priority *int               `json:"priority,omitempty"`
	} `json:"request"`
}

func (r *UpdateCrawlRuleRequest) Validity(resp *UpdateCrawlRuleResponse) {
	if r.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "keapp_rule_id_required"
	}
}

type UpdateCrawlRuleResponse struct {
	apiobj.BaseResponse
}

type DeleteCrawlRuleRequest struct {
	apiobj.BaseRequest
	Request struct {
		ID uint `json:"id"`
	} `json:"request"`
}

func (r *DeleteCrawlRuleRequest) Validity(resp *DeleteCrawlRuleResponse) {
	if r.Request.ID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "keapp_rule_id_required"
	}
}

type DeleteCrawlRuleResponse struct {
	apiobj.BaseResponse
}

type TriggerCrawlRequest struct {
	apiobj.BaseRequest
	Request struct {
		AppID    uint              `json:"app_id"`
		TaskType web.CrawlTaskType `json:"task_type"`
	} `json:"request"`
}

func (r *TriggerCrawlRequest) Validity(resp *TriggerCrawlResponse) {
	if r.Request.AppID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "keapp_app_id_required"
	}
}

type TriggerCrawlResponse struct {
	apiobj.BaseResponse
	Response struct {
		TaskID uint `json:"task_id"`
	} `json:"response"`
}

type GetCrawlTaskRequest struct {
	apiobj.BaseRequest
	Request struct {
		TaskID uint `json:"task_id"`
	} `json:"request"`
}

func (r *GetCrawlTaskRequest) Validity(resp *GetCrawlTaskResponse) {
	if r.Request.TaskID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "keapp_task_id_required"
	}
}

type GetCrawlTaskResponse struct {
	apiobj.BaseResponse
	Response struct {
		Task *web.KeCrawlTask `json:"task"`
	} `json:"response"`
}

type ListCrawlTasksRequest struct {
	apiobj.BaseRequest
	Request struct {
		AppID  uint `json:"app_id"`
		Limit  int  `json:"limit"`
		Offset int  `json:"offset"`
	} `json:"request"`
}

type ListCrawlTasksResponse struct {
	apiobj.BaseResponse
	Response struct {
		Items []*web.KeCrawlTask `json:"items"`
	} `json:"response"`
}

type CancelCrawlTaskRequest struct {
	apiobj.BaseRequest
	Request struct {
		TaskID uint `json:"task_id"`
	} `json:"request"`
}

func (r *CancelCrawlTaskRequest) Validity(resp *CancelCrawlTaskResponse) {
	if r.Request.TaskID == 0 {
		resp.Code = errcode.ErrCode_BadRequest
		resp.Message = "keapp_task_id_required"
	}
}

type CancelCrawlTaskResponse struct {
	apiobj.BaseResponse
}

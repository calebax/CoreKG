package webctl

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/keapp/models/web"
	"github.com/insmtx/corekg/apps/keapp/services/svcweb"
	"github.com/ygpkg/yg-go/apis/errcode"
)

func AddCrawlRule(ctx *gin.Context, req *AddCrawlRuleRequest, resp *AddCrawlRuleResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}

	entity := &web.KeWebCrawlRule{
		AppID:    req.Request.AppID,
		RuleType: req.Request.RuleType,
		Pattern:  req.Request.Pattern,
		Priority: req.Request.Priority,
	}
	if err := svcweb.AddCrawlRule(ctx, entity); err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "keapp_add_rule_failed"
		return
	}
}

func ListCrawlRules(ctx *gin.Context, req *ListCrawlRulesRequest, resp *ListCrawlRulesResponse) {
	items, err := svcweb.ListCrawlRules(ctx, req.Request.AppID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "keapp_list_rules_failed"
		return
	}
	resp.Response.Items = items
}

func UpdateCrawlRule(ctx *gin.Context, req *UpdateCrawlRuleRequest, resp *UpdateCrawlRuleResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}

	entity := &web.KeWebCrawlRule{
		ID: req.Request.ID,
	}
	if req.Request.RuleType != nil {
		entity.RuleType = *req.Request.RuleType
	}
	if req.Request.Pattern != nil {
		entity.Pattern = *req.Request.Pattern
	}
	if req.Request.Priority != nil {
		entity.Priority = *req.Request.Priority
	}

	if err := svcweb.UpdateCrawlRule(ctx, entity); err != nil {
		if errors.Is(err, svcweb.ErrRuleNotFound) {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "keapp_rule_not_found"
			return
		}
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "keapp_update_rule_failed"
		return
	}
}

func DeleteCrawlRule(ctx *gin.Context, req *DeleteCrawlRuleRequest, resp *DeleteCrawlRuleResponse) {
	if req.Validity(resp); resp.Code != errcode.CodeOK {
		return
	}

	if err := svcweb.DeleteCrawlRule(ctx, req.Request.ID); err != nil {
		if errors.Is(err, svcweb.ErrRuleNotFound) {
			resp.Code = errcode.ErrCode_BadRequest
			resp.Message = "keapp_rule_not_found"
			return
		}
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "keapp_delete_rule_failed"
		return
	}
}

package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtokeywords"
	"github.com/insmtx/corekg/apps/kecore/services/svckeywords"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// CreateMajorKeyword 创建行业术语
// @Tags 关键词管理
// @Summary 创建行业术语
// @Description 创建行业术语
// @Router /kecore.CreateMajorKeyword [post]
// @Param request body dtokeywords.CreateMajorKeywordRequest true "request"
// @Success 200 {object} dtokeywords.CreateMajorKeywordResponse "response"
func CreateMajorKeyword(ctx *gin.Context, req *dtokeywords.CreateMajorKeywordRequest, resp *dtokeywords.CreateMajorKeywordResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[CreateMajorKeyword] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svckeywords.CreateMajorKeyword(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[CreateMajorKeyword] svckeywords.CreateMajorKeyword failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "创建失败"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// DeleteMajorKeyword 删除行业术语
// @Tags 关键词管理
// @Summary 删除行业术语
// @Description 删除行业术语
// @Router /kecore.DeleteMajorKeyword [post]
// @Param request body dtokeywords.DeleteMajorKeywordRequest true "request"
// @Success 200 {object} dtokeywords.DeleteMajorKeywordResponse "response"
func DeleteMajorKeyword(ctx *gin.Context, req *dtokeywords.DeleteMajorKeywordRequest, resp *dtokeywords.DeleteMajorKeywordResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[DeleteMajorKeyword] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}
	// TODO: 需要手动注册路由和修改 Message 的值
	res, err := svckeywords.DeleteMajorKeyword(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[DeleteMajorKeyword] svckeywords.DeleteMajorKeyword failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = errcode.GetMessage(errcode.ErrCode_InternalError)
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// UpdateMajorKeyword 修改行业术语
// @Tags 关键词管理
// @Summary 修改行业术语
// @Description 修改行业术语
// @Router /kecore.UpdateMajorKeyword [post]
// @Param request body dtokeywords.UpdateMajorKeywordRequest true "request"
// @Success 200 {object} dtokeywords.UpdateMajorKeywordResponse "response"
func UpdateMajorKeyword(ctx *gin.Context, req *dtokeywords.UpdateMajorKeywordRequest, resp *dtokeywords.UpdateMajorKeywordResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[UpdateMajorKeyword] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}
	// TODO: 需要手动注册路由和修改 Message 的值
	res, err := svckeywords.UpdateMajorKeyword(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[UpdateMajorKeyword] svckeywords.UpdateMajorKeyword failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = errcode.GetMessage(errcode.ErrCode_InternalError)
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// ListMajorKeywords 获取行业术语列表
// @Tags 关键词管理
// @Summary 获取行业术语列表
// @Description 获取行业术语列表
// @Router /kecore.ListMajorKeywords [post]
// @Param request body dtokeywords.ListMajorKeywordsRequest true "request"
// @Success 200 {object} dtokeywords.ListMajorKeywordsResponse "response"
func ListMajorKeywords(ctx *gin.Context, req *dtokeywords.ListMajorKeywordsRequest, resp *dtokeywords.ListMajorKeywordsResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[ListMajorKeywords] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}
	// TODO: 需要手动注册路由和修改 Message 的值
	res, err := svckeywords.ListMajorKeywords(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[ListMajorKeywords] svckeywords.ListMajorKeywords failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = errcode.GetMessage(errcode.ErrCode_InternalError)
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// GetMajorKeyword 获取行业术语列表
// @Tags 关键词管理
// @Summary 获取行业术语列表
// @Description 获取行业术语列表
// @Router /kecore.GetMajorKeyword [post]
// @Param request body dtokeywords.GetMajorKeywordRequest true "request"
// @Success 200 {object} dtokeywords.GetMajorKeywordResponse "response"
func GetMajorKeyword(ctx *gin.Context, req *dtokeywords.GetMajorKeywordRequest, resp *dtokeywords.GetMajorKeywordResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[GetMajorKeyword] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svckeywords.GetMajorKeyword(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetMajorKeyword] svckeywords.GetMajorKeyword failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = errcode.GetMessage(errcode.ErrCode_InternalError)
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

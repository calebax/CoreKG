package apis

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtokeywords"
	"github.com/insmtx/corekg/apps/kecore/services/svckeywords"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// ListSynonymKeywords 获取同义词列表
// @Tags 关键词管理
// @Summary 获取同义词列表
// @Description 获取同义词列表
// @Router /kecore.ListSynonymKeywords [post]
// @Param request body dtokeywords.ListSynonymKeywordsRequest true "request"
// @Success 200 {object} dtokeywords.ListSynonymKeywordsResponse "response"
func ListSynonymKeywords(ctx *gin.Context, req *dtokeywords.ListSynonymKeywordsRequest, resp *dtokeywords.ListSynonymKeywordsResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[ListSynonymKeywords] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svckeywords.ListSynonymKeywords(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[ListSynonymKeywords] svckeywords.ListSynonymKeywords failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "获取列表失败"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// GetSynonymKeyword 获取同义词详情
// @Tags 关键词管理
// @Summary 获取同义词详情
// @Description 获取同义词详情
// @Router /kecore.GetSynonymKeyword [post]
// @Param request body dtokeywords.GetSynonymKeywordRequest true "request"
// @Success 200 {object} dtokeywords.GetSynonymKeywordResponse "response"
func GetSynonymKeyword(ctx *gin.Context, req *dtokeywords.GetSynonymKeywordRequest, resp *dtokeywords.GetSynonymKeywordResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[GetSynonymKeyword] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}
	res, err := svckeywords.GetSynonymKeyword(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[GetSynonymKeyword] svckeywords.GetSynonymKeyword failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "获取详情失败"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// CreateSynonymKeyword 创建同义词
// @Tags 关键词管理
// @Summary 创建同义词
// @Description 创建同义词
// @Router /kecore.CreateSynonymKeyword [post]
// @Param request body dtokeywords.CreateSynonymKeywordRequest true "request"
// @Success 200 {object} dtokeywords.CreateSynonymKeywordResponse "response"
func CreateSynonymKeyword(ctx *gin.Context, req *dtokeywords.CreateSynonymKeywordRequest, resp *dtokeywords.CreateSynonymKeywordResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[CreateSynonymKeyword] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svckeywords.CreateSynonymKeyword(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[CreateSynonymKeyword] svckeywords.CreateSynonymKeyword failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "创建失败"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// DeleteSynonymKeyword 删除同义词主词
// @Tags 关键词管理
// @Summary 删除同义词主词
// @Description 删除同义词主词
// @Router /kecore.DeleteSynonymKeyword [post]
// @Param request body dtokeywords.DeleteSynonymKeywordRequest true "request"
// @Success 200 {object} dtokeywords.DeleteSynonymKeywordResponse "response"
func DeleteSynonymKeyword(ctx *gin.Context, req *dtokeywords.DeleteSynonymKeywordRequest, resp *dtokeywords.DeleteSynonymKeywordResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[DeleteSynonymKeyword] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svckeywords.DeleteSynonymKeyword(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[DeleteSynonymKeyword] svckeywords.DeleteSynonymKeyword failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "删除失败"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

// UpdateSynonymKeyword 修改同义词内容
// @Tags 关键词管理
// @Summary 修改同义词内容
// @Description 修改同义词内容
// @Router /kecore.UpdateSynonymKeyword [post]
// @Param request body dtokeywords.UpdateSynonymKeywordRequest true "request"
// @Success 200 {object} dtokeywords.UpdateSynonymKeywordResponse "response"
func UpdateSynonymKeyword(ctx *gin.Context, req *dtokeywords.UpdateSynonymKeywordRequest, resp *dtokeywords.UpdateSynonymKeywordResponse) {
	if req.Validity(resp); resp.Code != 0 {
		logs.ErrorContextf(ctx, "[UpdateSynonymKeyword] request invalid, req: %s, error message: %v", logs.JSON(req), resp.Message)
		return
	}

	res, err := svckeywords.UpdateSynonymKeyword(ctx, req)
	if err != nil {
		logs.ErrorContextf(ctx, "[UpdateSynonymKeyword] svckeywords.UpdateSynonymKeyword failed, err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "修改失败"
		return
	}
	resp.Code = res.Code
	resp.Message = res.Message
	resp.Response = res.Response
}

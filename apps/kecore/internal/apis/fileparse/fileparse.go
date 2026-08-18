package fileparse

import (
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/fs"
	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
	"github.com/ygpkg/yg-go/apis/errcode"
	"github.com/ygpkg/yg-go/logs"
)

// GetContent 获取文件解析内容
// @Tags 文件解析
// @Summary 获取文件解析内容
// @Description 获取文件解析内容
// @Router /forest.GetContent [post]
// @Param user body GetContentRequest true "入参"
// @Success 200 {object} GetContentResponse "返回值"
func GetContent(ctx *gin.Context, req *GetContentRequest, resp *GetContentResponse) {
	if req.Validity(resp); resp.Code != 0 {
		return
	}

	// 1.判断uin的forest_id是否存在
	f, err := forest.GetForestFileByID(req.Request.FileID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_query_file_failed" // 查询文件失败
		return
	}
	if f.ParseStatus == foresttype.TaskStatusPending || f.ParseStatus == foresttype.TaskStatusRunning {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_file_parsing" // 文件正在解析
		return
	}
	if f.ParseStatus != foresttype.TaskStatusSuccess {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_parse_not_generated" // 解析未生成
		return
	}
	content, err := fs.GetFileContent(f)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_content_failed" // 获取文件解析内容失败
		return
	}
	resp.Response.Content = string(content)
	resp.Response.Status = f.ParseStatus
}

// GetMindMap 获取思维导图
// @Tags 文件解析
// @Summary 获取思维导图
// @Description 获取思维导图
// @Router /forest.GetMindMap [post]
// @Param user body GetMindMapRequest true "入参"
// @Success 200 {object} GetMindMapResponse "返回值"
func GetMindMap(ctx *gin.Context, req *GetMindMapRequest, resp *GetMindMapResponse) {
	if req.Validity(resp); resp.Code != 0 {
		return
	}

	f, err := forest.GetForestFileByID(req.Request.FileID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_query_file_failed" // 查询文件失败
		return
	}
	frs, err := forest.GetForestByID(ctx, f.ForestID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetMindMap GetForestByID err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_forest_failed" // 获取知识库失败
		return
	}

	resp.Response.Status = f.DescStatus

	if f.DescStatus == foresttype.TaskStatusPending || f.DescStatus == foresttype.TaskStatusRunning {
		//resp.Code = errcode.ErrCode_InternalError
		//resp.Message = "kecore_description_generating" // 文件描述正在生成中
		return
	}
	if f.DescStatus != "" && f.DescStatus != foresttype.TaskStatusSuccess {
		//resp.Code = errcode.ErrCode_InternalError
		//resp.Message = "kecore_description_not_generated" // 文件描述未生成
		return
	}

	w := essearch.NewPureWrapper(ctx, frs.EsIndex(), []uint{frs.ID}, []uint{f.ID}, nil)
	desc, err := w.GetFileDesc()
	if err != nil || desc == nil {
		logs.ErrorContextf(ctx, "GetMindMap GetFileDesc err: %v, desc[%v]", err, desc)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_description_failed" // 获取文件描述记录失败
		return
	}

	resp.Response.MindMap = desc.MindMap
	resp.Response.Status = f.DescStatus
}

// GetAnalysis 获取智能分析
// @Tags 文件解析
// @Summary 获取智能分析
// @Description 获取智能分析
// @Router /forest.GetAnalysis [post]
// @Param user body GetAnalysisRequest true "入参"
// @Success 200 {object} GetAnalysisResponse "返回值"
func GetAnalysis(ctx *gin.Context, req *GetAnalysisRequest, resp *GetAnalysisResponse) {
	if req.Validity(resp); resp.Code != 0 {
		return
	}

	f, err := forest.GetForestFileByID(req.Request.FileID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_query_file_failed" // 查询文件失败
		return
	}
	frs, err := forest.GetForestByID(ctx, f.ForestID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetAnalysis GetForestByID err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_forest_failed" // 获取知识库失败
		return
	}

	resp.Response.Status = f.DescStatus

	if f.DescStatus == foresttype.TaskStatusPending || f.DescStatus == foresttype.TaskStatusRunning {
		//resp.Code = errcode.ErrCode_InternalError
		//resp.Message = "kecore_description_generating" // 文件描述正在生成中
		return
	}
	if f.DescStatus != "" && f.DescStatus != foresttype.TaskStatusSuccess {
		//resp.Code = errcode.ErrCode_InternalError
		//resp.Message = "kecore_description_not_generated" // 文件描述未生成
		return
	}

	w := essearch.NewPureWrapper(ctx, frs.EsIndex(), []uint{frs.ID}, []uint{f.ID}, nil)
	desc, err := w.GetFileDesc()
	if err != nil || desc == nil {
		logs.ErrorContextf(ctx, "GetAnalysis GetAnalysis err: %v,desc[%v]", err, desc)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_description_failed" // 获取文件描述记录失败
		return
	}

	resp.Response.Analysis = desc.Abstract
	resp.Response.Status = f.DescStatus
}

// GetFileDescription 获取文件描述
// @Tags 文件解析
// @Summary 获取文件描述
// @Description 获取文件描述
// @Router /forest.GetFileDescription [post]
// @Param user body GetFileDescriptionRequest true "入参"
// @Success 200 {object} GetFileDescriptionResponse "返回值"
func GetFileDescription(ctx *gin.Context, req *GetFileDescriptionRequest, resp *GetFileDescriptionResponse) {
	if req.Validity(resp); resp.Code != 0 {
		return
	}

	f, err := forest.GetForestFileByID(req.Request.FileID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_query_file_failed" // 查询文件失败
		return
	}
	frs, err := forest.GetForestByID(ctx, f.ForestID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetFileDescription GetForestByID err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_forest_failed" // 获取知识库失败
		return
	}

	if f.DescStatus == foresttype.TaskStatusPending || f.DescStatus == foresttype.TaskStatusRunning {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_description_generating" // 文件描述正在生成中
		return
	}
	if f.DescStatus != "" && f.DescStatus != foresttype.TaskStatusSuccess {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_description_not_generated" // 文件描述未生成
		return
	}

	w := essearch.NewPureWrapper(ctx, frs.EsIndex(), []uint{frs.ID}, []uint{f.ID}, nil)
	desc, err := w.GetFileDesc()
	if err != nil || desc == nil {
		logs.ErrorContextf(ctx, "GetFileDescription GetFileDesc err: %v, desc[%v]", err, desc)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_description_failed" // 获取文件描述记录失败
		return
	}
	resp.Response.Description = desc.Description
	resp.Response.Abstract = desc.Abstract
	resp.Response.MindMap = desc.MindMap
}

// GetRecommendQuestions 获取智能推荐问题
// @Tags 文件解析
// @Summary 获取智能推荐问题
// @Description 获取智能推荐问题
// @Router /forest.GetRecommendQuestions [post]
// @Param user body GetRecommendQuestionsRequest true "入参"
// @Success 200 {object} GetRecommendQuestionsResponse "返回值"
func GetRecommendQuestions(ctx *gin.Context, req *GetRecommendQuestionsRequest, resp *GetRecommendQuestionsResponse) {
	if req.Validity(resp); resp.Code != 0 {
		return
	}

	f, err := forest.GetForestFileByID(req.Request.FileID)
	if err != nil {
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_query_file_failed" // 查询文件失败
		return
	}
	frs, err := forest.GetForestByID(ctx, f.ForestID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetRecommendQuestions GetForestByID err: %v", err)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_forest_failed" // 获取知识库失败
		return
	}

	resp.Response.Status = f.DescStatus

	if f.DescStatus == foresttype.TaskStatusPending || f.DescStatus == foresttype.TaskStatusRunning {
		//resp.Code = errcode.ErrCode_InternalError
		//resp.Message = "kecore_description_generating" // 文件描述正在生成中
		return
	}
	if f.DescStatus != "" && f.DescStatus != foresttype.TaskStatusSuccess {
		//resp.Code = errcode.ErrCode_InternalError
		//resp.Message = "kecore_description_not_generated" // 文件描述未生成
		return
	}

	w := essearch.NewPureWrapper(ctx, frs.EsIndex(), []uint{frs.ID}, []uint{f.ID}, nil)
	desc, err := w.GetFileDesc()
	if err != nil || desc == nil {
		logs.ErrorContextf(ctx, "GetRecommendQuestions GetFileDesc err: %v, desc[%v]", err, desc)
		resp.Code = errcode.ErrCode_InternalError
		resp.Message = "kecore_get_description_failed" // 获取文件描述记录失败
		return
	}

	resp.Response.RecommendQuestions = desc.Questions
	resp.Response.Status = f.DescStatus
}

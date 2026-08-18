package mds

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/company"
	"github.com/insmtx/corekg/apps/kechat/models/chatquestion"
	"github.com/insmtx/corekg/apps/kecore/internal/apis/graphctl"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtoforestfile"

	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/apps/kecore/models/graph"
	"github.com/insmtx/corekg/apps/kecore/services/membership"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/insmtx/corekg/version"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/i18n"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

// GraphQuotaMD graph quota middleware
func GraphQuotaMD(ctx *gin.Context) {
	//if private deployed
	if version.DeployMode() != "" {
		ctx.Next()
		return
	}
	cmp, err := company.GetCompany(runtime.CompanyID(ctx))
	if err != nil {
		logs.ErrorContextf(ctx, "company.GetCompany(%d) error: %v", runtime.CompanyID(ctx), err)
		ctx.AbortWithStatusJSON(http.StatusOK, apiobj.BaseResponse{
			Code:    http.StatusInternalServerError,
			Message: i18n.T(runtime.GetLanguage(ctx), "account_get_company_info_failed")})
		return
	}

	ct, err := graph.GetGraphByCompanyID(ctx, cmp.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "[GraphQuotaMD] GetGraphByCompanyID(%d) error: %v", cmp.ID, err)
		ctx.AbortWithStatusJSON(http.StatusOK, apiobj.BaseResponse{
			Code:    http.StatusForbidden,
			Message: i18n.T(runtime.GetLanguage(ctx), "kecore_get_company_graph_failed")})
		return
	}
	lct := len(ct)

	if lct >= cmp.Quota.GraphQuota {
		logs.WarnContextf(ctx, "GraphLimited company[%v] has disk quota[%v], now used[%v],desire[%v]", cmp.ID, cmp.Quota.GraphQuota, lct, 1)
		ctx.AbortWithStatusJSON(http.StatusOK, apiobj.BaseResponse{
			Code: http.StatusForbidden,

			Message: i18n.T(runtime.GetLanguage(ctx), "kecore_quota_graph_limited")})
		return
	}

	//pass
	ctx.Next()
}

// ArticleQuotaMD article quota middleware
func ArticleQuotaMD(ctx *gin.Context) {
	if version.DeployMode() != "" {
		ctx.Next()
		return
	}

	valid, left, err := membership.NewQuotaManager().Check(ctx, &membership.QuotaCheckReq{
		CompanyID:    runtime.CompanyID(ctx),
		ResourceType: membership.QuotaResourceTypeArticle,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "quotaManager.Check(%d) error: %v", runtime.CompanyID(ctx), err)
		ctx.AbortWithStatusJSON(http.StatusOK, apiobj.BaseResponse{
			Code:    http.StatusInternalServerError,
			Message: i18n.T(runtime.GetLanguage(ctx), "account_get_company_info_failed")})
		return
	}

	if !valid || left <= 0 {
		logs.WarnContextf(ctx, "ArticleLimited companyID: %d, resourceType: %s, valid: %v, left: %d",
			runtime.CompanyID(ctx), membership.QuotaResourceTypeArticle, valid, left)
		ctx.AbortWithStatusJSON(http.StatusOK, apiobj.BaseResponse{
			Code:    http.StatusForbidden,
			Message: i18n.T(runtime.GetLanguage(ctx), "kecore_quota_article_limited")})
		return
	}

	ctx.Next()
}

// Deprecated: 历史版本弃用v2.6废弃
func DiskQuotaMD(ctx *gin.Context) {
	//if private deployed
	if version.DeployMode() != "" {
		ctx.Next()
		return
	}

	file, err := ctx.FormFile("file")
	if err != nil {
		logs.ErrorContextf(ctx, "FormFile() error: %v", err)
		ctx.AbortWithStatusJSON(http.StatusOK, apiobj.BaseResponse{
			Code:    http.StatusInternalServerError,
			Message: i18n.T(runtime.GetLanguage(ctx), "kecore_get_form_file_failed")})
		return
	}

	valid, left, err := membership.NewQuotaManager().Check(ctx, &membership.QuotaCheckReq{
		CompanyID:    runtime.CompanyID(ctx),
		ResourceType: membership.QuotaResourceTypeDisk,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "quotaManager.Check(%d) error: %v", runtime.CompanyID(ctx), err)
		ctx.AbortWithStatusJSON(http.StatusOK, apiobj.BaseResponse{
			Code:    http.StatusInternalServerError,
			Message: i18n.T(runtime.GetLanguage(ctx), "account_get_company_info_failed")})
		return
	}

	if !valid || left < file.Size {
		logs.WarnContextf(ctx, "DiskLimited companyID: %d, resourceType: %s, valid: %v, left: %d, desire: %d",
			runtime.CompanyID(ctx), membership.QuotaResourceTypeDisk, valid, left, file.Size)
		ctx.AbortWithStatusJSON(http.StatusOK, apiobj.BaseResponse{
			Code:    http.StatusForbidden,
			Message: i18n.T(runtime.GetLanguage(ctx), "kecore_quota_disk_limited")})
		return
	}

	//pass
	ctx.Next()
}

// Deprecated: 历史版本弃用
func DiskPreQuotaMD(ctx *gin.Context) {
	//if private deployed
	if version.DeployMode() != "" {
		ctx.Next()
		return
	}
	if ctx.Request.Body != nil {
		//read request body and after check the request body would be repair
		bodyBytes, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			logs.ErrorContextf(ctx, "io.ReadAll() error: %v", err)
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		var req dtoforestfile.PreUploadFileRequest
		if err = json.Unmarshal(bodyBytes, &req); err != nil {
			logs.ErrorContextf(ctx, "json.Unmarshal() error: %v", err)
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}
		//get company
		cmp, err := company.GetCompany(runtime.CompanyID(ctx))
		if err != nil {
			logs.ErrorContextf(ctx, "company.GetCompany(%d) error: %v", runtime.CompanyID(ctx), err)
			ctx.AbortWithStatusJSON(http.StatusOK, apiobj.BaseResponse{
				Code:    http.StatusInternalServerError,
				Message: i18n.T(runtime.GetLanguage(ctx), "account_get_company_info_failed")})
			return
		}

		diskSize, err := forest.GetFilesSizeByCompanyID(ctx, cmp.ID)
		if err != nil {
			logs.ErrorContextf(ctx, "forest.GetFilesSizeByCompanyID(%d) error: %v", cmp.ID, err)
			ctx.AbortWithStatusJSON(http.StatusOK, apiobj.BaseResponse{
				Code:    http.StatusInternalServerError,
				Message: i18n.T(runtime.GetLanguage(ctx), "kecore_get_company_file_size_failed"),
			})
			return
		}

		var requestFileSize int64
		for _, file := range req.Request.Files {
			requestFileSize += file.Size
		}

		if diskSize+requestFileSize > cmp.Quota.DiskQuota {
			logs.WarnContextf(ctx, "DiskLimited company[%v] has disk quota[%v], now used[%v],desire[%v]", cmp.ID, cmp.Quota.DiskQuota, diskSize, requestFileSize)
			ctx.AbortWithStatusJSON(http.StatusOK, apiobj.BaseResponse{
				Code:    http.StatusForbidden,
				Message: i18n.T(runtime.GetLanguage(ctx), "kecore_quota_disk_limited")})
			return
		}

		ctx.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	} else {
		logs.ErrorContextf(ctx.Request.Context(), "Request.Body is nil")
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}
	ctx.Next()
}

func DiskPreQuotaMDV2(ctx *gin.Context) {
	// 如果是私有部署，直接跳过
	if version.DeployMode() != "" {
		ctx.Next()
		return
	}

	if ctx.Request.Body == nil {
		logs.ErrorContextf(ctx.Request.Context(), "Request.Body is nil")
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// 读取请求体
	bodyBytes, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		logs.ErrorContextf(ctx, "io.ReadAll() error: %v", err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	var req dtoforestfile.PreUploadFileRequest
	if err = json.Unmarshal(bodyBytes, &req); err != nil {
		logs.ErrorContextf(ctx, "json.Unmarshal() error: %v", err)
		ctx.AbortWithStatus(http.StatusBadRequest)
		return
	}

	// 计算请求中所有文件总大小
	var totalRequestSize int64
	for _, file := range req.Request.Files {
		totalRequestSize += file.Size
	}

	valid, left, err := membership.NewQuotaManager().Check(ctx, &membership.QuotaCheckReq{
		CompanyID:    runtime.CompanyID(ctx),
		ResourceType: membership.QuotaResourceTypeDisk,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "quotaManager.Check(%d) error: %v", runtime.CompanyID(ctx), err)
		ctx.AbortWithStatusJSON(http.StatusOK, apiobj.BaseResponse{
			Code:    http.StatusInternalServerError,
			Message: i18n.T(runtime.GetLanguage(ctx), "account_get_company_info_failed"),
		})
		return
	}

	// 检查磁盘配额
	if !valid || left <= 0 {
		logs.WarnContextf(ctx, "DiskLimited companyID: %d, resourceType: %s, valid: %v, left: %d, requested: %d",
			runtime.CompanyID(ctx), membership.QuotaResourceTypeDisk, valid, left, totalRequestSize)
		ctx.AbortWithStatusJSON(http.StatusOK, apiobj.BaseResponse{
			Code:    global.ErrCodeRequireDiskQuota,
			Message: i18n.T(runtime.GetLanguage(ctx), "kecore_quota_disk_limited"),
		})
		return
	}

	// 重新设置请求体，供后续处理
	ctx.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	ctx.Next()
}

func QAChatQuotaMD(ctx *gin.Context) {
	//if private deployed
	if version.DeployMode() != "" {
		ctx.Next()
		return
	}

	valid, left, err := membership.NewQuotaManager().Check(ctx, &membership.QuotaCheckReq{
		CompanyID:    runtime.CompanyID(ctx),
		ResourceType: membership.QuotaResourceTypeQA,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "quotaManager.Check(%d) error: %v", runtime.CompanyID(ctx), err)
		ctx.AbortWithStatusJSON(http.StatusOK, apiobj.BaseResponse{
			Code:    http.StatusInternalServerError,
			Message: i18n.T(runtime.GetLanguage(ctx), "account_get_company_info_failed")})
		return
	}

	if !valid || left <= 0 {
		logs.WarnContextf(ctx, "QALimited companyID: %d, resourceType: %s, valid: %v, left: %d",
			runtime.CompanyID(ctx), membership.QuotaResourceTypeQA, valid, left)
		ctx.AbortWithStatusJSON(http.StatusOK, CodeResp{
			Response: struct {
				Code int `json:"code"`
			}{Code: http.StatusBadRequest},
		})
		return
	}

	//pass
	ctx.Next()
}

func QAChatQuotaWithErrMD(ctx *gin.Context) {
	//if private deployed
	if version.DeployMode() != "" {
		ctx.Next()
		return
	}
	cmp, err := company.GetCompany(runtime.CompanyID(ctx))
	if err != nil {
		logs.ErrorContextf(ctx, "company.GetCompany(%d) error: %v", runtime.CompanyID(ctx), err)
		ctx.AbortWithStatusJSON(http.StatusOK, apiobj.BaseResponse{
			Code:    http.StatusInternalServerError,
			Message: i18n.T(runtime.GetLanguage(ctx), "account_get_company_info_failed")})
		return
	}

	qas, err := chatquestion.GetUnscopedQAByCompanyID(ctx, cmp.ID)
	if err != nil {
		logs.ErrorContextf(ctx, "keqa.GetUnscopedQAByCompanyID(%d) error: %v", cmp.ID, err)
		ctx.AbortWithStatusJSON(http.StatusOK, apiobj.BaseResponse{
			Code:    http.StatusInternalServerError,
			Message: i18n.T(runtime.GetLanguage(ctx), "kecore_get_company_agent_failed")})
		return
	}

	if qas >= int64(cmp.Quota.QAQuota) {
		logs.WarnContextf(ctx, "QALimited company[%v] has qa quota[%v], now used[%v],desire[%v]", cmp.ID, cmp.Quota.QAQuota, qas, 1)
		ctx.AbortWithStatusJSON(http.StatusOK, apiobj.BaseResponse{
			Code:    http.StatusBadRequest,
			Message: i18n.T(runtime.GetLanguage(ctx), "kecore_quota_qa_limited"),
		})
		return
	}

	//pass
	ctx.Next()
}

type CodeResp struct {
	apiobj.BaseResponse
	Response struct {
		Code int `json:"code"`
	}
}

func HasForestUsePerm(ctx *gin.Context) {
	uin := runtime.Uin(ctx)
	cmpID := runtime.CompanyID(ctx)

	if ctx.Request.Body != nil {
		bodyBytes, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			logs.ErrorContextf(ctx, "HasForestViewPerm read request body err:%v", err)
			//runtime.InternalError(ctx, "获取请求体失败")
			runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_get_request_body_failed")) // 获取请求体失败
			return
		}

		var req *apiobj.DetailIdRequest
		if err = json.Unmarshal(bodyBytes, &req); err != nil {
			logs.ErrorContextf(ctx, "HasForestViewPerm Unmarshal request body err:%v", err)
			//runtime.InternalError(ctx, "解析请求体失败")
			runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_parse_request_body_failed")) // 解析请求体失败
			return
		}

		if !CanViewForest(ctx, req.Request.ID, uin, cmpID) {
			logs.WarnContextf(ctx, "uin[%v] desire to use forest[%v] with comapnyID[%v] but doesn't have perm", uin, req.Request.ID, cmpID)
			runtime.BadRequest(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_no_forest_use_permission")) // 无知识库使用权限
			return
		}

		ctx.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	} else {
		logs.ErrorContextf(ctx, "HasForestViewPerm get nil request body")
		runtime.BadRequest(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_request_body_empty")) // 请求体为空
		return
	}
	ctx.Next()
}

func CanViewForest(ctx context.Context, frID, uin, companyID uint) bool {
	var c int64
	ag, err := forest.GetForestByID(ctx, frID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logs.WarnContextf(ctx, "[CanViewForest] forest not found, forest id: %d", frID)
		} else {
			logs.ErrorContextf(ctx, "[CanViewForest] get forest failed, forest id: %d, err: %v", frID, err)
		}
		return false
	}

	if ag.PublicScope == foresttype.PublicScopePublic ||
		(ag.PublicScope == foresttype.PublicScopeCompany && ag.CompanyID == companyID) ||
		(ag.PublicScope == foresttype.PublicScopePrivate && ag.Uin == uin) {
		return true
	}

	if err := dbutil.Knownow().Table(foresttype.TableNameKeResourceScope).
		Where("deleted_at IS NULL").
		Where("resource_type = ?", foresttype.ResourceTypeForest).
		Where("resource_id = ?", frID).
		Where("scope_type = ?", foresttype.ScopeTypeUser).
		Where("scope_id = ?", uin).
		Count(&c).Error; err != nil {
		logs.ErrorContextf(ctx, "get resource_scope faild %v", err)
		return false
	}

	return c > 0
}

func HasGraphUsePerm(ctx *gin.Context) {
	uin := runtime.Uin(ctx)
	cmpID := runtime.CompanyID(ctx)

	if ctx.Request.Body != nil {
		bodyBytes, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			logs.ErrorContextf(ctx, "HasGraphUsePerm read request body err:%v", err)
			runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_get_request_body_failed")) // 获取请求体失败
			return
		}

		var req *graphctl.GetGraphInfoRequest
		if err = json.Unmarshal(bodyBytes, &req); err != nil {
			logs.ErrorContextf(ctx, "HasGraphUsePerm Unmarshal request body err:%v", err)
			runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_parse_request_body_failed")) // 解析请求体失败
			return
		}

		if !CanViewGraph(ctx, req.Request.GraphID, uin, cmpID) {
			logs.WarnContextf(ctx, "uin[%v] desire to use Graph[%v] with comapnyID[%v] but doesn't have perm", uin, req.Request.GraphID, cmpID)
			runtime.BadRequest(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_no_forest_use_permission")) // 无知识库使用权限
			return
		}

		ctx.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	} else {
		logs.ErrorContextf(ctx, "HasGraphUsePerm get nil request body")
		runtime.BadRequest(ctx, i18n.T(runtime.GetLanguage(ctx), "kecore_request_body_empty")) // 请求体为空
		return
	}
	ctx.Next()
}

func CanViewGraph(ctx context.Context, gID, uin, companyID uint) bool {
	var c int64
	g, err := graph.GetGraph(ctx, gID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetGraph(%d) failed: %v", gID, err)
		return false
	}

	if g.PublicScope == foresttype.PublicScopePublic ||
		(g.PublicScope == foresttype.PublicScopeCompany && g.CompanyID == companyID) ||
		(g.PublicScope == foresttype.PublicScopePrivate && g.Uin == uin) {
		return true
	}

	if err := dbutil.Knownow().Table(foresttype.TableNameKeResourceScope).
		Where("deleted_at IS NULL").
		Where("resource_type = ?", foresttype.ResourceTypeGraph).
		Where("resource_id = ?", gID).
		Where("scope_type = ?", foresttype.ScopeTypeUser).
		Where("scope_id = ?", uin).
		Count(&c).Error; err != nil {
		logs.ErrorContextf(ctx, "get resource_scope faild %v", err)
		return false
	}

	return c > 0
}

const (
	AuthHeaderKey              = "authorization"
	AuthValueGetOriginResource = "wdxiRMkJoCS02Ysb"
	AuthValueMigrateInterface  = "zL+C06KZLjBqyyM0"
)

func HasHeaderStr(key, value string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.GetHeader(key) != value {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, apiobj.BaseResponse{
				Code:    http.StatusUnauthorized,
				Message: i18n.T(runtime.GetLanguage(ctx), "kecore_mds_no_permission"),
			})
			return
		}
		ctx.Next()
	}
}

func IsBanResource(ctx context.Context, uin, resourceID uint, resourceType foresttype.ResourceType) bool {
	var c int64
	if err := dbutil.Knownow().Table(foresttype.TableNameKeResourceScope).
		Where("deleted_at IS NULL").
		Where("resource_type = ?", resourceType).
		Where("resource_id = ?", resourceID).
		Where("scope_type = ?", foresttype.ScopeTypeUser).
		Where("scope_id = ?", uin).
		Where("action = ?", foresttype.ActionBan).
		Count(&c).Error; err != nil {
		logs.ErrorContextf(ctx, "get resource_scope faild %v", err)
		return false
	}
	return c > 0
}

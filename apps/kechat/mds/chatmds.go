package mds

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/employee"
	"github.com/insmtx/corekg/apps/kechat/models/chatagent"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kechat/models/llmchat"
	"github.com/insmtx/corekg/apps/kecore/services/membership"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/version"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/apis/runtime/auth"
	"github.com/ygpkg/yg-go/i18n"
	"github.com/ygpkg/yg-go/logs"
)

// QAQuotaForRoleORForestMD 检查用量
func QAQuotaForRoleORForestMD(ctx *gin.Context) {
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

	if !valid || left < 1 {
		logs.WarnContextf(ctx, "QALimited companyID: %d, resourceType: %s, valid: %v, left: %d, desire: %d",
			runtime.CompanyID(ctx), membership.QuotaResourceTypeQA, valid, left, 1)
		ctx.AbortWithStatusJSON(http.StatusOK, apiobj.BaseResponse{
			Code:    global.ErrCodeRequireQaQuota,
			Message: i18n.T(runtime.GetLanguage(ctx), "kecore_quota_qa_limited")})
		return
	}

	ctx.Next()
}

// QAQuotaForRoleORForestMD 检查用量
func AgentQuotaMD(ctx *gin.Context) {
	//if private deployed
	if version.DeployMode() != "" {
		ctx.Next()
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

	if !valid || left < 1 {
		logs.WarnContextf(ctx, "AgentLimited companyID: %d, resourceType: %s, valid: %v, left: %d, desire: %d",
			runtime.CompanyID(ctx), membership.QuotaResourceTypeAgent, valid, left, 1)
		ctx.AbortWithStatusJSON(http.StatusOK, apiobj.BaseResponse{
			Code:    global.ErrCodeRequireAgentQuota,
			Message: i18n.T(runtime.GetLanguage(ctx), "kecore_quota_agent_limited")})
		return
	}

	ctx.Next()
}

func HasAgentUsePerm(ctx *gin.Context) {
	uin := runtime.Uin(ctx)
	cmpID := runtime.CompanyID(ctx)

	if ctx.Request.Body != nil {
		bodyBytes, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			logs.ErrorContextf(ctx, "HasForestViewPerm read request body err:%v", err)
			runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kechat_get_request_body_failed")) // 获取请求体失败
			return
		}

		var req *apiobj.DetailIdRequest
		if err = json.Unmarshal(bodyBytes, &req); err != nil {
			logs.ErrorContextf(ctx, "HasForestViewPerm Unmarshal request body err:%v", err)
			runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kechat_parse_request_body_failed")) // 解析请求体失败
			return
		}

		if !chatagent.CanViewAgent(ctx, req.Request.ID, uin, cmpID) {
			logs.WarnContextf(ctx, "uin[%v] desire to use agent[%v] with comapnyID[%v] but doesn't have perm", uin, req.Request.ID, cmpID)
			runtime.BadRequest(ctx, i18n.T(runtime.GetLanguage(ctx), "kechat_no_agent_use_perm")) // 无轻应用使用权限
			return
		}

		ctx.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	} else {
		logs.ErrorContextf(ctx, "HasForestViewPerm get nil request body")
		runtime.BadRequest(ctx, i18n.T(runtime.GetLanguage(ctx), "kechat_request_body_empty")) // 请求体为空
		return
	}
	ctx.Next()
}

func QAQuotaMD(ctx *gin.Context) {
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
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kechat_server_error")) // 服务器内部错误
		return
	}
	if !valid || left <= 0 {
		logs.WarnContextf(ctx, "QALimited companyID: %d, resourceType: %s, valid: %v, left: %d", runtime.CompanyID(ctx), membership.QuotaResourceTypeQA, valid, left)
		bt, err := json.Marshal(struct {
			Code int `json:"code"`
		}{
			Code: int(global.ErrCodeRequireQaQuota),
		})
		if err != nil {
			logs.ErrorContextf(ctx, "QAQuotaMD marshal struct faild: %v", err)
			runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kechat_server_error")) // 服务器内部错误
			return
		}
		ctx.Writer.Write(bt)
		ctx.Writer.Flush()
		ctx.AbortWithStatus(http.StatusOK)
		return
	}

	//pass
	ctx.Next()
}

func HasAgentManagePerm(ctx *gin.Context) {
	uin := runtime.Uin(ctx)
	cmpID := runtime.CompanyID(ctx)

	emp, err := employee.GetEmployeeByUin(uin)
	if err != nil {
		logs.ErrorContextf(ctx, "employee.GetEmployeeByUin(%v) error: %v", uin, err)
		runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kechat_get_employee_info_failed")) // 获取员工信息失败
	}
	if emp.SysRole == accounttype.SysRoleSysAdmin {
		ctx.Next()
		return
	}

	if ctx.Request.Body != nil {
		bodyBytes, err := io.ReadAll(ctx.Request.Body)
		if err != nil {
			logs.ErrorContextf(ctx, "HasForestViewPerm read request body err:%v", err)
			runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kechat_get_request_body_failed")) // 获取请求体失败
			return
		}

		var req *apiobj.DetailIdRequest
		if err = json.Unmarshal(bodyBytes, &req); err != nil {
			logs.ErrorContextf(ctx, "HasForestViewPerm Unmarshal request body err:%v", err)
			runtime.InternalError(ctx, i18n.T(runtime.GetLanguage(ctx), "kechat_parse_request_body_failed")) // 解析请求体失败
			return
		}

		if !chatagent.CanManageAgent(ctx, req.Request.ID, uin) {
			logs.WarnContextf(ctx, "uin[%v] desire to use agent[%v] with comapnyID[%v] but doesn't have perm", uin, req.Request.ID, cmpID)
			runtime.BadRequest(ctx, i18n.T(runtime.GetLanguage(ctx), "kechat_no_agent_manage_perm")) // 无轻应用管理权限
			return
		}

		ctx.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
	} else {
		logs.ErrorContextf(ctx, "HasForestViewPerm get nil request body")
		runtime.BadRequest(ctx, i18n.T(runtime.GetLanguage(ctx), "kechat_request_body_empty")) // 请求体为空
		return
	}
	ctx.Next()
}

const (
	KeyExtID   = "ext_id"
	KeyExtType = "ext_type"
)

func ExtAuthMD(ctx *gin.Context) {
	tok := ctx.GetHeader("Authorization")
	authStr := strings.TrimPrefix(tok, auth.AuthBearer)

	claims := new(chatagent.ExternalClaims)
	_, err := jwt.ParseWithClaims(authStr, claims, func(token *jwt.Token) (interface{}, error) {
		if token.Claims == nil {
			return nil, fmt.Errorf("token claims is nil")
		}
		c, ok := token.Claims.(*chatagent.ExternalClaims)
		if !ok {
			return nil, fmt.Errorf("token claims is not ExternalClaims")
		}

		return auth.GetJwtSecret(c.Issuer)
	})
	if err != nil {
		logs.ErrorContextf(ctx, "jwt.ParseWithClaims(%s) error: %v", authStr, err)
		ctx.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	ctx.Set(KeyExtID, claims.ExtID)
	ctx.Set(KeyExtType, claims.ExtType)

	ctx.Next()
}

func ExtAgentValidMD(ctx *gin.Context) {
	if ctx.MustGet(KeyExtType).(chatagent.ExternalType) != chatagent.ExternalAgent {
		logs.ErrorContextf(ctx, "ext_agent.ExternalAgent invalid")
		ctx.AbortWithStatus(http.StatusForbidden)
		return
	}
	agID := ctx.MustGet(KeyExtID).(uint)
	ag, err := chatagent.GetChatAgentByID(ctx, agID)
	if err != nil {
		logs.ErrorContextf(ctx, "chat_agent.GetChatAgentByID(%d) error: %v", agID, err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if ag.ExternalStatus == chattype.ExternalStatusDisable {
		ctx.AbortWithStatusJSON(http.StatusForbidden, apiobj.BaseResponse{Code: http.StatusForbidden, Message: i18n.T(runtime.GetLanguage(ctx), DisableResponse)})
		return
	}

	ctx.Set(global.CtxKeyCompanyID, ag.CompanyID)
	ctx.Next()
}

var (
	DisableResponse = "kechat_disable_response"
)

const AgentVersion = "agent_version"

func ExtAgentChatValidMD(ctx *gin.Context) {
	if ctx.MustGet(KeyExtType).(chatagent.ExternalType) != chatagent.ExternalAgent {
		logs.ErrorContextf(ctx, "ext_agent.ExternalAgent invalid")
		ctx.AbortWithStatus(http.StatusForbidden)
		return
	}
	agID := ctx.MustGet(KeyExtID).(uint)
	ag, err := chatagent.GetChatAgentByID(ctx, agID)
	if err != nil {
		logs.ErrorContextf(ctx, "chat_agent.GetChatAgentByID(%d) error: %v", agID, err)
		ctx.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	if ag.ExternalStatus == chattype.ExternalStatusDisable {
		content := llmchat.WriteResult{
			Content: i18n.T(runtime.GetLanguage(ctx), DisableResponse),
		}
		if err := json.NewEncoder(ctx.Writer).Encode(content); err != nil {
			logs.ErrorContextf(ctx, "ExtAgentChatValidMD Encode(%v) error: %v", content, err)
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		ctx.Writer.Flush()
		ctx.AbortWithStatus(http.StatusForbidden)
		return
	}
	ctx.Set(AgentVersion, ag.Version)
	ctx.Set(global.CtxKeyCompanyID, ag.CompanyID)
	ctx.Next()
}

func QAQuotaExtAgentMD(ctx *gin.Context) {
	//if private deployed
	if version.DeployMode() != "" {
		ctx.Next()
		return
	}
	cmpID := ctx.MustGet(global.CtxKeyCompanyID).(uint)

	valid, left, err := membership.NewQuotaManager().Check(ctx, &membership.QuotaCheckReq{
		CompanyID:    cmpID,
		ResourceType: membership.QuotaResourceTypeQA,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "quotaManager.Check(%d) error: %v", cmpID, err)
		ctx.AbortWithStatusJSON(http.StatusOK, apiobj.BaseResponse{
			Code:    http.StatusInternalServerError,
			Message: i18n.T(runtime.GetLanguage(ctx), "kechat_server_error"), // 服务器内部错误
		})
		return
	}
	if !valid || left <= 0 {
		logs.WarnContextf(ctx, "QALimited companyID: %d, resourceType: %s, valid: %v, left: %d", cmpID, membership.QuotaResourceTypeQA, valid, left)
		ctx.AbortWithStatusJSON(http.StatusForbidden, apiobj.BaseResponse{
			Code:    global.ErrCodeRequireQaQuota,
			Message: i18n.T(runtime.GetLanguage(ctx), "kechat_quota_limited"), // 额度限制
		})
		return
	}

	ctx.Next()
}

package accountmds

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/account/models/company"
	"github.com/insmtx/corekg/apps/kecore/services/membership"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/version"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/logs"
)

func EmployeeQuotaBindMD(ctx *gin.Context) {
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
		var req *struct {
			apiobj.BaseRequest
			Request struct {
				Key        string `json:"key"`
				Code       string `json:"code"`
				Way        string `json:"way"`
				DomainName string `json:"domain_name"`
			}
		}
		if err = json.Unmarshal(bodyBytes, &req); err != nil {
			logs.ErrorContextf(ctx, "json.Unmarshal() error: %v", err)
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}
		//get company
		k, err := company.GetInviteByKey(req.Request.Key)
		if err != nil {
			logs.ErrorContextf(ctx, "company.GetInviteByKey(%v) error: %v", req.Request.Key, err)
			ctx.AbortWithStatus(http.StatusBadRequest)
			return
		}

		valid, left, err := membership.NewQuotaManager().Check(ctx, &membership.QuotaCheckReq{
			CompanyID:    k.CompanyID,
			ResourceType: membership.QuotaResourceTypeEmployee,
		})
		if err != nil {
			logs.ErrorContextf(ctx, "quotaManager.Check(%d) error: %v", k.CompanyID, err)
			runtime.InternalError(ctx, "检查额度失败")
			return
		}
		if !valid || left <= 0 {
			logs.WarnContextf(ctx.Request.Context(), "检查额度失败, companyID: %d, resourceType: %s, valid: %v, left: %d", k.CompanyID, membership.QuotaResourceTypeEmployee, valid, left)
			ctx.AbortWithStatusJSON(http.StatusOK, apiobj.BaseResponse{
				Code:    global.ErrCodeRequireEmployeeQuota,
				Message: "成员额度不足",
			})
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

func EmployeeQuotaMD(ctx *gin.Context) {
	//if private deployed
	if version.DeployMode() != "" {
		ctx.Next()
		return
	}

	companyID := runtime.CompanyID(ctx)

	valid, left, err := membership.NewQuotaManager().Check(ctx, &membership.QuotaCheckReq{
		CompanyID:    companyID,
		ResourceType: membership.QuotaResourceTypeEmployee,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "quotaManager.Check(%d) error: %v", companyID, err)
		runtime.InternalError(ctx, "检查额度失败")
		return
	}
	if !valid || left <= 0 {
		logs.WarnContextf(ctx.Request.Context(), "检查额度失败, companyID: %d, resourceType: %s, valid: %v, left: %d", companyID, membership.QuotaResourceTypeEmployee, valid, left)
		ctx.AbortWithStatusJSON(http.StatusOK, apiobj.BaseResponse{
			Code:    global.ErrCodeRequireEmployeeQuota,
			Message: "成员额度不足",
		})
		return
	}

	ctx.Next()
}

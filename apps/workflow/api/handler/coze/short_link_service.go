package coze

import (
	"context"
	"strconv"
	"time"

	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"

	"github.com/insmtx/corekg/apps/workflow/application/openauth"
	application "github.com/insmtx/corekg/apps/workflow/application/singleagent"
	"github.com/insmtx/corekg/apps/workflow/domain/openauth/openapiauth/entity"
	"github.com/ygpkg/yg-go/logs"
)

type getAgentShortLinkCodeRequest struct {
	BotID   int64 `json:"bot_id,string" form:"bot_id" query:"bot_id"`
	SpaceID int64 `json:"space_id,string" form:"space_id" query:"space_id"`
}

type getAgentShortLinkCodeResponse struct {
	Code int32  `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		ShortLinkCode string `json:"short_link_code"`
		Status        string `json:"status"`
	} `json:"data"`
}

type setAgentExternalStatusRequest struct {
	BotID   int64  `json:"bot_id,string" form:"bot_id" query:"bot_id"`
	SpaceID int64  `json:"space_id,string" form:"space_id" query:"space_id"`
	Status  string `json:"status" form:"status" query:"status"`
}

type getAgentUserIDByShortCodeRequest struct {
	ShortCode string `json:"short_code" form:"short_code" query:"short_code"`
}

type getAgentAuthTokenResponse struct {
	Code int32  `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		AgentID   string `json:"agent_id"`
		AuthToken string `json:"auth_token"`
		ExpireAt  int64  `json:"expire_at"`
	} `json:"data"`
}

// GetAgentShortLinkCode 获取 agent 的短链 code
// @router /api/internal/agent/short_link_code [POST]
func GetAgentShortLinkCode(ctx context.Context, c *app.RequestContext) {
	var req getAgentShortLinkCodeRequest
	if err := c.BindAndValidate(&req); err != nil {
		logs.WarnContextf(ctx, "GetAgentShortLinkCode bind/validate failed: %v", err)
		invalidParamRequestResponse(c, err.Error())
		return
	}

	if req.BotID <= 0 || req.SpaceID <= 0 {
		logs.WarnContextf(ctx, "GetAgentShortLinkCode invalid params bot_id=%d space_id=%d", req.BotID, req.SpaceID)
		invalidParamRequestResponse(c, "bot_id and space_id are required")
		return
	}

	code, status, err := application.SingleAgentSVC.GetAgentShortLinkCode(ctx, req.BotID, req.SpaceID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetAgentShortLinkCode failed bot_id=%d space_id=%d: %v", req.BotID, req.SpaceID, err)
		internalServerErrorResponse(ctx, c, err)
		return
	}

	resp := getAgentShortLinkCodeResponse{
		Code: 0,
		Msg:  "success",
	}
	resp.Data.ShortLinkCode = code
	resp.Data.Status = status

	c.JSON(consts.StatusOK, resp)
}

// SetAgentExternalStatus 设置 agent 外部访问状态
// @router /api/internal/agent/set_external_status [POST]
func SetAgentExternalStatus(ctx context.Context, c *app.RequestContext) {
	var req setAgentExternalStatusRequest
	if err := c.BindAndValidate(&req); err != nil {
		logs.WarnContextf(ctx, "SetAgentExternalStatus bind/validate failed: %v", err)
		invalidParamRequestResponse(c, err.Error())
		return
	}

	if req.BotID <= 0 || req.SpaceID <= 0 || req.Status == "" {
		logs.WarnContextf(ctx, "SetAgentExternalStatus invalid params bot_id=%d space_id=%d status=%s", req.BotID, req.SpaceID, req.Status)
		invalidParamRequestResponse(c, "bot_id, space_id and status are required")
		return
	}

	if err := application.SingleAgentSVC.SetAgentExternalStatus(ctx, req.BotID, req.SpaceID, req.Status); err != nil {
		logs.ErrorContextf(ctx, "SetAgentExternalStatus failed bot_id=%d space_id=%d status=%s: %v", req.BotID, req.SpaceID, req.Status, err)
		internalServerErrorResponse(ctx, c, err)
		return
	}

	c.JSON(consts.StatusOK, map[string]any{
		"code": 0,
		"msg":  "success",
	})
}

// PublicGetAgentUserIDByShortCode 根据短链查询 agent 的授权 token（公开接口）
// @router /api/public/agent/external_token [GET]
func PublicGetAgentUserIDByShortCode(ctx context.Context, c *app.RequestContext) {
	var req getAgentUserIDByShortCodeRequest
	if err := c.BindAndValidate(&req); err != nil {
		logs.WarnContextf(ctx, "PublicGetAgentUserIDByShortCode bind/validate failed: %v", err)
		invalidParamRequestResponse(c, err.Error())
		return
	}
	if req.ShortCode == "" {
		logs.WarnContextf(ctx, "PublicGetAgentUserIDByShortCode invalid params short_code is empty")
		invalidParamRequestResponse(c, "short_code is required")
		return
	}

	agentID, userID, err := application.SingleAgentSVC.GetAgentByShortCode(ctx, req.ShortCode)
	if err != nil {
		logs.ErrorContextf(ctx, "PublicGetAgentUserIDByShortCode failed short_code=%s: %v", req.ShortCode, err)
		internalServerErrorResponse(ctx, c, err)
		return
	}

	token, expireAt, err := getValidOrCreateAuthToken(ctx, userID)
	if err != nil {
		logs.ErrorContextf(ctx, "PublicGetAgentUserIDByShortCode generate token failed agent_id=%d user_id=%d short_code=%s: %v", agentID, userID, req.ShortCode, err)
		internalServerErrorResponse(ctx, c, err)
		return
	}

	resp := getAgentAuthTokenResponse{
		Code: 0,
		Msg:  "success",
	}
	resp.Data.AgentID = strconv.FormatInt(agentID, 10)
	resp.Data.AuthToken = token
	resp.Data.ExpireAt = expireAt

	c.JSON(consts.StatusOK, resp)
}

func getValidOrCreateAuthToken(ctx context.Context, userID int64) (string, int64, error) {
	const (
		minValidDuration   = 24 * time.Hour
		newTokenValidFor   = 72 * time.Hour
		tokenNameForPublic = "public_external_token"
	)

	return getValidOrCreateApiKey(ctx, userID, minValidDuration, newTokenValidFor, tokenNameForPublic)
}

func getValidOrCreateApiKey(ctx context.Context, userID int64, minValidDuration, newTokenValidFor time.Duration, tokenName string) (string, int64, error) {
	const (
		defaultListLimit = 50
		defaultListPage  = 1
	)

	now := time.Now()
	page := int64(defaultListPage)
	limit := int64(defaultListLimit)
	for {
		listResp, err := openauth.OpenAuthApplication.OpenAPIDomainSVC.List(ctx, &entity.ListApiKey{
			UserID: userID,
			Limit:  limit,
			Page:   page,
		})
		if err != nil {
			return "", 0, err
		}
		for _, key := range listResp.ApiKeys {
			if key == nil {
				continue
			}
			if key.ExpiredAt == 0 || key.ExpiredAt-now.Unix() >= int64(minValidDuration.Seconds()) {
				token := key.ApiKey
				return token, key.ExpiredAt, nil
			}
		}
		if !listResp.HasMore {
			break
		}
		page++
	}

	expireAt := now.Add(newTokenValidFor).Unix()
	newKey, err := openauth.OpenAuthApplication.OpenAPIDomainSVC.Create(ctx, &entity.CreateApiKey{
		Name:   tokenName,
		Expire: expireAt,
		UserID: userID,
		AkType: entity.AkTypeCustomer,
	})
	if err != nil {
		return "", 0, err
	}

	return newKey.ApiKey, expireAt, nil
}

package singleagent

import (
	"context"
	"crypto/rand"
	"math/big"
	"strings"

	"github.com/insmtx/corekg/apps/workflow/application/base/ctxutil"
	"github.com/insmtx/corekg/apps/workflow/domain/agent/singleagent/entity"
	"github.com/insmtx/corekg/apps/workflow/domain/agent/singleagent/repository"
	"github.com/insmtx/corekg/apps/workflow/domain/permission"
	"github.com/insmtx/corekg/apps/workflow/pkg/errorx"
	"github.com/ygpkg/yg-go/logs"
	"github.com/insmtx/corekg/apps/workflow/types/consts"
	"github.com/insmtx/corekg/apps/workflow/types/errno"
	"github.com/insmtx/corekg/apps/workflow/utils/requestyygu"
)

const (
	shortCodeMinLen = 6
	shortCodeMaxLen = 8
)

// CreateShortLink generates a unique short code and upserts it for the bot.
// It retries on conflict a few times with fresh codes.
func CreateShortLink(ctx context.Context, appCtx *ServiceComponents, botID, userID, spaceID int64) (string, error) {
	repo := repository.NewShortLinkRepo(appCtx.DB, appCtx.IDGen)

	existing, err := repo.GetActiveByBotUserSpace(ctx, botID, userID, spaceID)
	if err != nil {
		return "", err
	}
	if existing != nil {
		return existing.ShortCode, nil
	}

	var lastErr error
	for i := 0; i < 3; i++ {
		code := genShortCode()
		err := repo.CreateOrUpdate(ctx, &entity.ShortLink{
			BotID:     botID,
			ShortCode: code,
			UserID:    userID,
			SpaceID:   spaceID,
			Status:    entity.ShortLinkStatusActive,
			UserToken: "",
		})
		if err == nil {
			return code, nil
		}
		lastErr = err
		logs.WarnContextf(ctx, "CreateShortLink attempt %d failed, bot: %d, code: %s, err: %v", i+1, botID, code, err)
	}

	return "", lastErr
}

// GetAgentShortLinkCode 获取 agent 已存在的短链 code
func (s *SingleAgentApplicationService) GetAgentShortLinkCode(
	ctx context.Context, botID, spaceID int64,
) (string, string, error) {
	if botID <= 0 || spaceID <= 0 {
		return "", "", errorx.New(errno.ErrAgentInvalidParamCode, errorx.KV("msg", "bot_id or space_id is invalid"))
	}

	userID := resolveUserID(ctx)

	if err := s.checkAgentAccess(ctx, botID, spaceID, userID); err != nil {
		logs.WarnContextf(ctx, "GetAgentShortLinkCode access check failed bot_id=%d space_id=%d user_id=%d: %v", botID, spaceID, userID, err)
		return "", "", err
	}

	repo := repository.NewShortLinkRepo(s.appContext.DB, s.appContext.IDGen)
	link, err := repo.GetActiveByBotUserSpace(ctx, botID, userID, spaceID)
	if err != nil {
		logs.ErrorContextf(ctx, "GetAgentShortLinkCode query failed bot_id=%d space_id=%d user_id=%d: %v", botID, spaceID, userID, err)
		return "", "", err
	}
	if link == nil {
		logs.InfoContextf(ctx, "GetAgentShortLinkCode no existing link, creating bot_id=%d space_id=%d user_id=%d", botID, spaceID, userID)
		code, err := CreateShortLink(ctx, s.appContext, botID, userID, spaceID)
		if err != nil {
			logs.ErrorContextf(ctx, "GetAgentShortLinkCode create short link failed bot_id=%d space_id=%d user_id=%d: %v", botID, spaceID, userID, err)
			return "", "", err
		}
		logs.InfoContextf(ctx, "GetAgentShortLinkCode created short link bot_id=%d space_id=%d user_id=%d code=%s", botID, spaceID, userID, code)
		return code, fromShortLinkStatus(entity.ShortLinkStatusPublicDisabled), nil
	}

	return link.ShortCode, fromShortLinkStatus(link.Status), nil
}

// SetAgentExternalStatus 设置 agent 的外部访问状态
func (s *SingleAgentApplicationService) SetAgentExternalStatus(ctx context.Context, botID, spaceID int64, status string) error {
	if botID <= 0 || spaceID <= 0 {
		return errorx.New(errno.ErrAgentInvalidParamCode, errorx.KV("msg", "bot_id or space_id is invalid"))
	}

	statusVal, err := toShortLinkStatus(status)
	if err != nil {
		return err
	}

	userID := resolveUserID(ctx)
	if err := s.checkAgentAccess(ctx, botID, spaceID, userID); err != nil {
		logs.WarnContextf(ctx, "SetAgentExternalStatus access check failed bot_id=%d space_id=%d user_id=%d: %v", botID, spaceID, userID, err)
		return err
	}

	repo := repository.NewShortLinkRepo(s.appContext.DB, s.appContext.IDGen)
	link, err := repo.GetActiveByBotUserSpace(ctx, botID, userID, spaceID)
	if err != nil {
		logs.ErrorContextf(ctx, "SetAgentExternalStatus query failed bot_id=%d space_id=%d user_id=%d: %v", botID, spaceID, userID, err)
		return err
	}
	if link == nil {
		logs.WarnContextf(ctx, "SetAgentExternalStatus short link not found bot_id=%d space_id=%d user_id=%d", botID, spaceID, userID)
		return errorx.New(errno.ErrAgentInvalidParamCode, errorx.KV("msg", "short link not exist"))
	}

	link.Status = statusVal
	if err := repo.CreateOrUpdate(ctx, link); err != nil {
		logs.ErrorContextf(ctx, "SetAgentExternalStatus update failed bot_id=%d space_id=%d user_id=%d status=%d: %v", botID, spaceID, userID, statusVal, err)
		return err
	}
	logs.InfoContextf(ctx, "SetAgentExternalStatus updated bot_id=%d space_id=%d user_id=%d status=%s", botID, spaceID, userID, fromShortLinkStatus(statusVal))
	return nil
}

// GetAgentByShortCode 返回公开短链对应的 agent 及用户 ID，用于无需鉴权的查询场景
func (s *SingleAgentApplicationService) GetAgentByShortCode(ctx context.Context, shortCode string) (int64, int64, error) {
	if strings.TrimSpace(shortCode) == "" {
		return 0, 0, errorx.New(errno.ErrAgentInvalidParamCode, errorx.KV("msg", "short_code is required"))
	}

	repo := repository.NewShortLinkRepo(s.appContext.DB, s.appContext.IDGen)
	link, err := repo.GetByShortCode(ctx, shortCode)
	if err != nil {
		logs.ErrorContextf(ctx, "GetAgentByShortCode query failed short_code=%s: %v", shortCode, err)
		return 0, 0, err
	}
	if link == nil {
		return 0, 0, errorx.New(errno.ErrAgentInvalidParamCode, errorx.KV("msg", "short link not exist"))
	}
	if link.Status != entity.ShortLinkStatusNormal {
		return 0, 0, errorx.New(errno.ErrAgentPermissionCode, errorx.KV("msg", "short link is not public"))
	}

	return link.BotID, link.UserID, nil
}

func toShortLinkStatus(status string) (int32, error) {
	switch strings.ToLower(status) {
	case "disabled":
		return entity.ShortLinkStatusPublicDisabled, nil
	case "normal":
		return entity.ShortLinkStatusNormal, nil
	default:
		return 0, errorx.New(errno.ErrAgentInvalidParamCode, errorx.KV("msg", "status invalid"))
	}
}

func fromShortLinkStatus(status int32) string {
	switch status {
	case entity.ShortLinkStatusNormal:
		return "normal"
	default:
		return "disabled"
	}
}

func resolveUserID(ctx context.Context) int64 {
	if uid := ctxutil.GetUIDFromCtx(ctx); uid != nil {
		return *uid
	}
	return ctxutil.MustGetUIDFromApiAuthCtx(ctx)
}

func (s *SingleAgentApplicationService) checkAgentAccess(ctx context.Context, botID, spaceID, userID int64) error {
	if err := checkUserSpace(ctx, userID, spaceID); err != nil {
		return err
	}

	agentDraft, err := s.DomainSVC.GetSingleAgentDraft(ctx, botID)
	if err != nil {
		return err
	}

	if agentDraft == nil {
		return errorx.New(errno.ErrAgentPermissionCode, errorx.KV("msg", "agent not exist"))
	}

	if agentDraft.SpaceID != consts.TemplateSpaceID && agentDraft.SpaceID != spaceID {
		return errorx.New(errno.ErrAgentPermissionCode, errorx.KV("msg", "agent not in the given space"))
	}

	if agentDraft.SpaceID != consts.TemplateSpaceID {
		allowedIDs, err := requestyygu.FilterCoreKGResourceIDsByScopePermission(
			ctx,
			userID,
			int64(permission.ResourceTypeAgent),
			[]int64{agentDraft.AgentID},
			string(permission.ActionRead),
		)
		if err != nil {
			logs.ErrorContextf(ctx, "check agent draft permission failed: user=%d agent=%d err=%v", userID, agentDraft.AgentID, err)
			return err
		}
		if _, ok := allowedIDs[agentDraft.AgentID]; !ok {
			logs.ErrorContextf(ctx, "user(%d) has no permission for agent draft(%d)", userID, agentDraft.AgentID)
			return errorx.New(errno.ErrAgentPermissionCode, errorx.KV("msg", "you are not the agent owner"))
		}
	}

	return nil
}

func genShortCode() string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789"

	length, err := rand.Int(rand.Reader, big.NewInt(int64(shortCodeMaxLen-shortCodeMinLen+1)))
	if err != nil {
		return ""
	}
	codeLen := int(length.Int64()) + shortCodeMinLen

	buf := make([]byte, codeLen)
	if _, err := rand.Read(buf); err != nil {
		return ""
	}
	for i := range buf {
		buf[i] = charset[int(buf[i])%len(charset)]
	}
	return string(buf)
}

package chatagent

import (
	"context"

	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/constants"
	"github.com/ygpkg/yg-go/i18n"
	"github.com/ygpkg/yg-go/logs"
)

// GetAgentI18nName 通过 agent name 获取多语言名称
func GetAgentI18nName(ctx context.Context, acceptLang string, agentName global.AgentName) string {
	if acceptLang == "" {
		ctxLang := ctx.Value(constants.CtxKeyLang)
		if ctxLang != nil {
			if lang, ok := ctxLang.(string); ok {
				acceptLang = lang
			}
		}
	}
	//acceptLang = locales.I18nConfig.DefaultLanguage.String()  // 默认语言
	lang := i18n.MatchLanguage(acceptLang).String()
	// 如果是简体中文，直接返回
	if lang == "zh-Hans" {
		return agentName.String()
	}
	// 如果不支持多语言，直接返回
	if !agentName.I18nSupported() {
		return agentName.String()
	}
	// 查询数据库中是否存在该名称的 agent
	i18nName := agentName.I18nName(lang)
	var count int64
	err := dbutil.Chat().Table(chattype.TableNameAgent).
		Where(chattype.TableNameAgent+".name = ?", i18nName).
		Count(&count).Error
	if err != nil {
		logs.ErrorContextf(ctx, "[GetAgentI18nName] check agent i18n name exist failed, err: %v, agentName: %s, lang: %s", err, agentName.String(), lang)
		return agentName.String()
	}
	if count == 0 {
		// 如果不存在，返回原始名称
		logs.ErrorContextf(ctx, "[GetAgentI18nName] agent i18n name not exist, use default name, agentName: %s, lang: %s", agentName.String(), lang)
		return agentName.String()
	}
	return i18nName
}

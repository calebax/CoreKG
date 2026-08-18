package dal

import (
	"context"
	"errors"

	"gorm.io/gorm"

	"github.com/insmtx/corekg/apps/workflow/domain/agent/singleagent/entity"
	"github.com/insmtx/corekg/apps/workflow/domain/agent/singleagent/internal/dal/model"
	"github.com/insmtx/corekg/apps/workflow/infra/idgen"
	"github.com/insmtx/corekg/apps/workflow/pkg/errorx"
	"github.com/insmtx/corekg/apps/workflow/types/errno"
)

type ShortLinkDAO struct {
	db    *gorm.DB
	idGen idgen.IDGenerator
}

func NewShortLinkDAO(db *gorm.DB, idGen idgen.IDGenerator) *ShortLinkDAO {
	return &ShortLinkDAO{
		db:    db,
		idGen: idGen,
	}
}

func (dao *ShortLinkDAO) CreateOrUpdate(ctx context.Context, link *entity.ShortLink) error {
	if link == nil {
		return errorx.New(errno.ErrAgentInvalidParamCode, errorx.KV("msg", "short link is nil"))
	}

	po := dao.shortLinkDo2Po(link)
	table := dao.db

	existing := &model.SingleAgentShortLink{}
	err := table.WithContext(ctx).Where(&model.SingleAgentShortLink{ShortCode: po.ShortCode}).First(existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return errorx.WrapByCode(err, errno.ErrAgentExecuteErrCode)
	}

	if err == nil {
		if existing.BotID != po.BotID {
			return errorx.New(errno.ErrAgentInvalidParamCode, errorx.KV("msg", "short code collision"))
		}
		po.ID = existing.ID
		po.CreatedAt = existing.CreatedAt
		if err := table.WithContext(ctx).Save(po).Error; err != nil {
			return errorx.WrapByCode(err, errno.ErrAgentUpdateCode)
		}
		return nil
	}

	if po.ID == 0 {
		id, err := dao.idGen.GenID(ctx)
		if err != nil {
			return errorx.WrapByCode(err, errno.ErrAgentIDGenFailCode)
		}
		po.ID = id
	}

	if err := table.WithContext(ctx).Create(po).Error; err != nil {
		return errorx.WrapByCode(err, errno.ErrAgentExecuteErrCode)
	}

	return nil
}

func (dao *ShortLinkDAO) GetByBotID(ctx context.Context, botID int64) (*entity.ShortLink, error) {
	po := &model.SingleAgentShortLink{}
	err := dao.db.WithContext(ctx).Where(&model.SingleAgentShortLink{BotID: botID}).First(po).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, errorx.WrapByCode(err, errno.ErrAgentGetCode)
	}

	return dao.shortLinkPo2Do(po), nil
}

func (dao *ShortLinkDAO) GetActiveByBotUserSpace(ctx context.Context, botID, userID, spaceID int64) (*entity.ShortLink, error) {
	po := &model.SingleAgentShortLink{}
	statusList := []int32{entity.ShortLinkStatusPublicDisabled, entity.ShortLinkStatusNormal}
	err := dao.db.WithContext(ctx).
		Where(&model.SingleAgentShortLink{
			BotID:   botID,
			UserID:  userID,
			SpaceID: spaceID,
		}).
		Where(map[string]any{"status": statusList}).
		First(po).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, errorx.WrapByCode(err, errno.ErrAgentGetCode)
	}

	return dao.shortLinkPo2Do(po), nil
}

func (dao *ShortLinkDAO) GetActiveByBotSpace(ctx context.Context, botID, spaceID int64) (*entity.ShortLink, error) {
	po := &model.SingleAgentShortLink{}
	statusList := []int32{entity.ShortLinkStatusPublicDisabled, entity.ShortLinkStatusNormal}
	err := dao.db.WithContext(ctx).
		Where(&model.SingleAgentShortLink{BotID: botID, SpaceID: spaceID}).
		Where(map[string]any{"status": statusList}).
		First(po).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, errorx.WrapByCode(err, errno.ErrAgentGetCode)
	}

	return dao.shortLinkPo2Do(po), nil
}

func (dao *ShortLinkDAO) GetByShortCode(ctx context.Context, code string) (*entity.ShortLink, error) {
	po := &model.SingleAgentShortLink{}
	err := dao.db.WithContext(ctx).Where(&model.SingleAgentShortLink{ShortCode: code}).First(po).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, errorx.WrapByCode(err, errno.ErrAgentGetCode)
	}

	return dao.shortLinkPo2Do(po), nil
}

func (dao *ShortLinkDAO) shortLinkDo2Po(e *entity.ShortLink) *model.SingleAgentShortLink {
	if e == nil {
		return nil
	}

	return &model.SingleAgentShortLink{
		ID:         e.ID,
		BotID:      e.BotID,
		ShortCode:  e.ShortCode,
		UserID:     e.UserID,
		SpaceID:    e.SpaceID,
		Status:     e.Status,
		LastUsedAt: e.LastUsedAt,
		UserToken:  e.UserToken,
		CreatedAt:  e.CreatedAt,
		UpdatedAt:  e.UpdatedAt,
	}
}

func (dao *ShortLinkDAO) shortLinkPo2Do(po *model.SingleAgentShortLink) *entity.ShortLink {
	if po == nil {
		return nil
	}

	return &entity.ShortLink{
		ID:         po.ID,
		BotID:      po.BotID,
		ShortCode:  po.ShortCode,
		UserID:     po.UserID,
		SpaceID:    po.SpaceID,
		Status:     po.Status,
		LastUsedAt: po.LastUsedAt,
		UserToken:  po.UserToken,
		CreatedAt:  po.CreatedAt,
		UpdatedAt:  po.UpdatedAt,
	}
}

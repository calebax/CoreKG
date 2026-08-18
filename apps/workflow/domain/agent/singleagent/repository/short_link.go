package repository

import (
	"context"

	"gorm.io/gorm"

	"github.com/insmtx/corekg/apps/workflow/domain/agent/singleagent/entity"
	"github.com/insmtx/corekg/apps/workflow/domain/agent/singleagent/internal/dal"
	"github.com/insmtx/corekg/apps/workflow/infra/idgen"
)

type ShortLinkRepo interface {
	CreateOrUpdate(ctx context.Context, link *entity.ShortLink) error
	GetByBotID(ctx context.Context, botID int64) (*entity.ShortLink, error)
	GetActiveByBotUserSpace(ctx context.Context, botID, userID, spaceID int64) (*entity.ShortLink, error)
	GetByShortCode(ctx context.Context, code string) (*entity.ShortLink, error)
	GetActiveByBotSpace(ctx context.Context, botID, spaceID int64) (*entity.ShortLink, error)
}

func NewShortLinkRepo(db *gorm.DB, idGen idgen.IDGenerator) ShortLinkRepo {
	return dal.NewShortLinkDAO(db, idGen)
}

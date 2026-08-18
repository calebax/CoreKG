package migrator

import (
	"context"

	"github.com/insmtx/corekg/apps/kechat/models/chat"
	"github.com/insmtx/corekg/apps/kechat/services/svcmodel"
	"github.com/ygpkg/yg-go/logs"
)

const (
	cozeModelIDStart = uint(66010)
)

type CozeModelSyncMigrator struct {
}

func (m *CozeModelSyncMigrator) Migrator(ctx context.Context) error {
	logs.InfoContextf(ctx, "CozeModelSyncMigrator start")

	chatModelDao := chat.NewChatModelDao()
	chatModels, err := chatModelDao.GetListByCond(ctx, &chat.ChatModelCond{})
	if err != nil {
		logs.ErrorContextf(ctx, "CozeModelSyncMigrator failed to query chat_model: %v", err)
		return err
	}

	usedCozeModelIDs := make(map[uint]struct{}, len(chatModels))
	for _, model := range chatModels {
		if model.CozeModelID == 0 {
			continue
		}
		usedCozeModelIDs[model.CozeModelID] = struct{}{}
	}

	nextCozeModelID := cozeModelIDStart
	nextAvailableCozeModelID := func() uint {
		for {
			if _, used := usedCozeModelIDs[nextCozeModelID]; !used {
				id := nextCozeModelID
				usedCozeModelIDs[id] = struct{}{}
				nextCozeModelID++
				return id
			}
			nextCozeModelID++
		}
	}

	for _, model := range chatModels {
		targetCozeModelID := model.CozeModelID
		if targetCozeModelID == 0 {
			targetCozeModelID = nextAvailableCozeModelID()
		}
		if _, err := svcmodel.SyncCozeModelInstance(ctx, model.ID, targetCozeModelID); err != nil {
			logs.ErrorContextf(ctx, "CozeModelSyncMigrator failed to sync chatModelID:%d to cozeModelID:%d, err:%v", model.ID, targetCozeModelID, err)
			return err
		}
	}

	logs.InfoContextf(ctx, "CozeModelSyncMigrator success, synced %d chat models", len(chatModels))
	return nil
}

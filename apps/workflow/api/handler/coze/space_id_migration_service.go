package coze

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/protocol/consts"
	"gorm.io/gorm"

	searchEntity "github.com/insmtx/corekg/apps/workflow/domain/search/entity"
	"github.com/insmtx/corekg/apps/workflow/infra/es"
	esimpl "github.com/insmtx/corekg/apps/workflow/infra/es/impl/es"
	ormmysql "github.com/insmtx/corekg/apps/workflow/infra/orm/impl/mysql"
	"github.com/insmtx/corekg/apps/workflow/pkg/lang/conv"
	"github.com/ygpkg/yg-go/logs"
)

const (
	resourceIndexName = "coze_resource"
	fieldSpaceID      = "space_id"
	fieldResID        = "res_id"
	updateBatchSize   = 200
	migrationKey      = "f5d8b9e1c2a74b9a8e6f0a1c3d2e7b4f"
)

type spaceIDMapping struct {
	OldSpaceID int64 `gorm:"column:old_space_id"`
	SpaceID    int64 `gorm:"column:space_id"`
}

type spaceIDMigrationResponse struct {
	Code int32  `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Mappings    int   `json:"mappings"`
		UpdatedDocs int   `json:"updated_docs"`
		ElapsedMs   int64 `json:"elapsed_ms"`
	} `json:"data"`
}

var (
	migrationOnce     sync.Once
	migrationDB       *gorm.DB
	migrationESClient es.Client
	migrationErr      error
)

// MigrateResourceSpaceID migrates space_id in ES based on history_data_migration_sync_record.
// @router /api/internal/space_id_migration [POST]
func MigrateResourceSpaceID(ctx context.Context, c *app.RequestContext) {
	if !validateMigrationKey(ctx, c) {
		return
	}

	start := time.Now()

	db, esClient, err := getMigrationClients(ctx)
	if err != nil {
		internalServerErrorResponse(ctx, c, err)
		return
	}

	mappings, err := loadSpaceIDMappings(ctx, db)
	if err != nil {
		internalServerErrorResponse(ctx, c, err)
		return
	}

	updatedTotal := 0
	for _, m := range mappings {
		if m.OldSpaceID <= 0 || m.SpaceID <= 0 || m.OldSpaceID == m.SpaceID {
			continue
		}
		updated, err := updateSpaceIDInES(ctx, esClient, m.OldSpaceID, m.SpaceID)
		if err != nil {
			internalServerErrorResponse(ctx, c, err)
			return
		}
		updatedTotal += updated
	}

	resp := spaceIDMigrationResponse{
		Code: 0,
		Msg:  "success",
	}
	resp.Data.Mappings = len(mappings)
	resp.Data.UpdatedDocs = updatedTotal
	resp.Data.ElapsedMs = time.Since(start).Milliseconds()

	c.JSON(consts.StatusOK, resp)
}

func validateMigrationKey(ctx context.Context, c *app.RequestContext) bool {
	provided := strings.TrimSpace(string(c.Query("key")))

	if provided == "" || provided != migrationKey {
		invalidParamRequestResponse(c, "invalid migration key")
		return false
	}

	return true
}

func getMigrationClients(ctx context.Context) (*gorm.DB, es.Client, error) {
	migrationOnce.Do(func() {
		migrationDB, migrationErr = ormmysql.New()
		if migrationErr != nil {
			return
		}
		migrationESClient, migrationErr = esimpl.New()
	})
	if migrationErr != nil {
		logs.ErrorContextf(ctx, "init migration clients failed: %v", migrationErr)
		return nil, nil, migrationErr
	}
	return migrationDB, migrationESClient, nil
}

func loadSpaceIDMappings(ctx context.Context, db *gorm.DB) ([]spaceIDMapping, error) {
	var mappings []spaceIDMapping
	err := db.WithContext(ctx).
		Raw("SELECT old_space_id, space_id FROM history_data_migration_sync_record GROUP BY old_space_id, space_id").
		Scan(&mappings).Error
	if err != nil {
		logs.ErrorContextf(ctx, "load space id mappings failed: %v", err)
		return nil, err
	}
	return mappings, nil
}

func updateSpaceIDInES(ctx context.Context, esClient es.Client, oldID, newID int64) (updated int, err error) {
	bi, err := esClient.NewBulkIndexer(resourceIndexName)
	if err != nil {
		return 0, err
	}
	defer func() {
		closeErr := bi.Close(ctx)
		if err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	var searchAfter []any
	for {
		size := updateBatchSize
		req := &es.Request{
			Query: &es.Query{
				Bool: &es.BoolQuery{
					Must: []es.Query{
						es.NewEqualQuery(fieldSpaceID, conv.Int64ToStr(oldID)),
					},
				},
			},
			Size: &size,
			Sort: []es.SortFiled{{Field: fieldResID, Asc: true}},
		}
		if len(searchAfter) > 0 {
			req.SearchAfter = searchAfter
		}

		result, searchErr := esClient.Search(ctx, resourceIndexName, req)
		if searchErr != nil {
			return updated, searchErr
		}
		if len(result.Hits.Hits) == 0 {
			break
		}

		lastResID := int64(0)
		for _, hit := range result.Hits.Hits {
			doc := &searchEntity.ResourceDocument{}
			if unmarshalErr := sonic.Unmarshal(hit.Source_, doc); unmarshalErr != nil {
				return updated, unmarshalErr
			}

			resID := doc.ResID
			docID := ""
			if resID > 0 {
				docID = conv.Int64ToStr(resID)
			} else if hit.Id_ != nil && *hit.Id_ != "" {
				docID = *hit.Id_
				if parsed, parseErr := strconv.ParseInt(docID, 10, 64); parseErr == nil {
					resID = parsed
				}
			}
			if docID == "" {
				return updated, fmt.Errorf("missing document id for space_id=%d", oldID)
			}

			body, marshalErr := json.Marshal(map[string]any{
				"doc": map[string]any{fieldSpaceID: newID},
			})
			if marshalErr != nil {
				return updated, marshalErr
			}

			if addErr := bi.Add(ctx, es.BulkIndexerItem{
				Index:      resourceIndexName,
				Action:     "update",
				DocumentID: docID,
				Body:       bytes.NewReader(body),
			}); addErr != nil {
				return updated, addErr
			}

			updated++
			if resID > 0 {
				lastResID = resID
			}
		}

		if lastResID == 0 {
			return updated, fmt.Errorf("failed to resolve res_id for space_id=%d", oldID)
		}
		searchAfter = []any{lastResID}

		if len(result.Hits.Hits) < size {
			break
		}
	}

	return updated, err
}

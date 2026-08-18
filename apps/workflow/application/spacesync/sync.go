/*
 * Copyright 2025 coze-dev Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package spacesync

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/insmtx/corekg/apps/workflow/application/base/appinfra"
	"github.com/insmtx/corekg/apps/workflow/conf"
	"github.com/ygpkg/yg-go/logs"
	"github.com/insmtx/corekg/apps/workflow/types/consts"
	"github.com/insmtx/corekg/apps/workflow/utils/requestyygu"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// SyncResult contains the result of a space sync operation
type SyncResult struct {
	SpaceID     int64
	MemberCount int
}

var (
	syncInfraOnce sync.Once
	syncInfra     *appinfra.AppDependencies
	syncInfraErr  error
)

func getSyncDB(ctx context.Context) (*gorm.DB, error) {
	syncInfraOnce.Do(func() {
		syncInfra, syncInfraErr = appinfra.Init(ctx, conf.GetAppConfig())
	})
	if syncInfraErr != nil {
		return nil, syncInfraErr
	}
	return syncInfra.DB, nil
}

// Sync synchronizes space and member information from ROC
func Sync(ctx context.Context) (*SyncResult, error) {
	company, err := requestyygu.GetCompanyInfo(ctx)
	if err != nil {
		logs.ErrorContextf(ctx, "sync space: get company info failed: %v", err)
		return nil, fmt.Errorf("get company info failed: %w", err)
	}

	deptData, err := requestyygu.GetDepartmentTree(ctx, true)
	if err != nil {
		logs.ErrorContextf(ctx, "sync space: get department tree failed: %v", err)
		return nil, fmt.Errorf("get department tree failed: %w", err)
	}

	db, err := getSyncDB(ctx)
	if err != nil {
		logs.ErrorContextf(ctx, "get sync db failed: %v", err)
		return nil, fmt.Errorf("get sync db failed: %w", err)
	}

	var result *SyncResult
	if err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res, txErr := syncSpacesAndMembers(ctx, tx, company, deptData.Employees)
		if txErr != nil {
			return txErr
		}
		result = res
		return nil
	}); err != nil {
		logs.ErrorContextf(ctx, "sync spaces/members failed: %v", err)
		return nil, fmt.Errorf("sync spaces/members failed: %w", err)
	}

	return result, nil
}

func syncSpacesAndMembers(ctx context.Context, db *gorm.DB, company *requestyygu.GetCompanyInfoResponse, employees []requestyygu.EmployeeInfo) (*SyncResult, error) {
	now := time.Now().UnixMilli()

	spaceID := int64(company.ID)
	spaceName := company.Name
	if spaceName == "" {
		spaceName = fmt.Sprintf("space_%d", spaceID)
	}

	// Map existing spaces
	var existingSpace spaceModel
	if err := db.WithContext(ctx).Where("id = ?", spaceID).First(&existingSpace).Error; err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			logs.ErrorContextf(ctx, "query existing space failed: %v", err)
			return nil, err
		}
	}
	exist := existingSpace.ID != 0

	// Determine owner from admin members if available
	var ownerID int64
	for _, emp := range employees {
		if emp.Role == "sys_admin" {
			ownerID = int64(emp.Uin)
			break
		}
	}

	// logoURI := strings.TrimSpace(company.Logo)
	logoURI := consts.DefaultTeamIcon

	// Upsert spaces
	if exist {
		if err := db.WithContext(ctx).
			Model(&spaceModel{}).
			Where("id = ?", spaceID).
			Updates(map[string]any{
				"owner_id":    ownerID,
				"name":        spaceName,
				"description": company.Description,
				"icon_uri":    logoURI,
				"updated_at":  now,
			}).Error; err != nil {
			logs.ErrorContextf(ctx, "update space %d failed: %v", spaceID, err)
			return nil, err
		}
	} else {
		if res := db.WithContext(ctx).Create(&spaceModel{
			ID:          spaceID,
			Name:        spaceName,
			Description: company.Description,
			IconURI:     logoURI,
			OwnerID:     ownerID,
			CreatorID:   ownerID,
			CreatedAt:   now,
			UpdatedAt:   now,
		}); res.Error != nil {
			logs.ErrorContextf(ctx, "create space %d failed: %v", spaceID, res.Error)
			return nil, res.Error
		}
	}

	// Upsert users and memberships
	currentUserIDs := make([]int64, 0, len(employees))
	for _, emp := range employees {
		if err := upsertUser(ctx, db, now, emp); err != nil {
			logs.ErrorContextf(ctx, "upsert user %d failed: %v", emp.Uin, err)
			return nil, err
		}

		roleType := int32(3) // member
		if emp.Role == "sys_admin" {
			roleType = 2 // admin
		}

		if res := db.WithContext(ctx).
			Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "space_id"}, {Name: "user_id"}},
				DoUpdates: clause.Assignments(map[string]interface{}{"role_type": roleType, "updated_at": now}),
			}).
			Create(&spaceUserModel{
				SpaceID:   spaceID,
				UserID:    int64(emp.Uin),
				RoleType:  roleType,
				CreatedAt: now,
				UpdatedAt: now,
			}); res.Error != nil {
			logs.ErrorContextf(ctx, "upsert membership space=%d user=%d failed: %v", spaceID, emp.Uin, res.Error)
			return nil, res.Error
		}

		currentUserIDs = append(currentUserIDs, int64(emp.Uin))
	}

	// Remove memberships no longer present
	if len(currentUserIDs) == 0 {
		if res := db.WithContext(ctx).Where("space_id = ?", spaceID).Delete(&spaceUserModel{}); res.Error != nil {
			logs.ErrorContextf(ctx, "delete all memberships for space=%d failed: %v", spaceID, res.Error)
			return nil, res.Error
		}
	} else {
		if res := db.WithContext(ctx).
			Where("space_id = ?", spaceID).
			Where("user_id NOT IN ?", currentUserIDs).
			Delete(&spaceUserModel{}); res.Error != nil {
			logs.ErrorContextf(ctx, "delete stale memberships for space=%d failed: %v", spaceID, res.Error)
			return nil, res.Error
		}
	}

	return &SyncResult{
		SpaceID:     spaceID,
		MemberCount: len(currentUserIDs),
	}, nil
}

func upsertUser(ctx context.Context, db *gorm.DB, now int64, emp requestyygu.EmployeeInfo) error {
	// 优先使用传入字段，缺失时用 uin 兜底，并做唯一前缀避免冲突
	baseName := strings.TrimSpace(emp.UserName)
	if baseName == "" {
		baseName = strings.TrimSpace(emp.Name)
	}
	if baseName == "" {
		baseName = fmt.Sprintf("user_%d", emp.Uin)
	}
	uniqueName := fmt.Sprintf("%d_%s", emp.Uin, baseName)
	name := baseName

	email := strings.TrimSpace(emp.Email)
	if email == "" {
		if emp.Phone != "" {
			email = fmt.Sprintf("%s@yygu.com", emp.Phone)
		} else {
			email = fmt.Sprintf("%d@yygu.com", emp.Uin)
		}
	}
	// 加上uin前缀，防止与已有账号邮件冲突
	email = fmt.Sprintf("%d_%s", emp.Uin, email)

	return db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{
				"name":        name,
				"unique_name": uniqueName,
				"email":       email,
				"updated_at":  now,
			}),
		}).
		Create(&userModel{
			ID:           int64(emp.Uin),
			Name:         name,
			UniqueName:   uniqueName,
			Email:        email,
			Description:  "synced from corekg",
			IconURI:      consts.DefaultUserIcon,
			UserVerified: true,
			Locale:       "zh-CN",
			CreatedAt:    now,
			UpdatedAt:    now,
		}).Error
}

// local gorm models (avoid importing internal packages)

type spaceModel struct {
	ID          int64          `gorm:"column:id;primaryKey;autoIncrement:false"`
	OwnerID     int64          `gorm:"column:owner_id"`
	Name        string         `gorm:"column:name"`
	Description string         `gorm:"column:description"`
	IconURI     string         `gorm:"column:icon_uri"`
	CreatorID   int64          `gorm:"column:creator_id"`
	CreatedAt   int64          `gorm:"column:created_at"`
	UpdatedAt   int64          `gorm:"column:updated_at"`
	DeletedAt   gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (spaceModel) TableName() string { return "space" }

type spaceUserModel struct {
	ID        int64 `gorm:"column:id;primaryKey;autoIncrement:true"`
	SpaceID   int64 `gorm:"column:space_id"`
	UserID    int64 `gorm:"column:user_id"`
	RoleType  int32 `gorm:"column:role_type"`
	CreatedAt int64 `gorm:"column:created_at"`
	UpdatedAt int64 `gorm:"column:updated_at"`
}

func (spaceUserModel) TableName() string { return "space_user" }

type userModel struct {
	ID           int64          `gorm:"column:id;primaryKey;autoIncrement:false"`
	Name         string         `gorm:"column:name"`
	UniqueName   string         `gorm:"column:unique_name"`
	Email        string         `gorm:"column:email"`
	Description  string         `gorm:"column:description"`
	IconURI      string         `gorm:"column:icon_uri"`
	UserVerified bool           `gorm:"column:user_verified"`
	Locale       string         `gorm:"column:locale"`
	CreatedAt    int64          `gorm:"column:created_at"`
	UpdatedAt    int64          `gorm:"column:updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"column:deleted_at"`
}

func (userModel) TableName() string { return "user" }

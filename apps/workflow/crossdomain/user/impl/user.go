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

package impl

import (
	"context"
	"fmt"

	"github.com/insmtx/corekg/apps/workflow/application/spacesync"
	crossuser "github.com/insmtx/corekg/apps/workflow/crossdomain/user"
	"github.com/insmtx/corekg/apps/workflow/domain/user/entity"
	"github.com/insmtx/corekg/apps/workflow/domain/user/service"
	"github.com/insmtx/corekg/apps/workflow/utils/requestyygu"
)

var defaultSVC crossuser.User

type impl struct {
	DomainSVC service.User
}

func InitDomainService(u service.User) crossuser.User {
	defaultSVC = &impl{
		DomainSVC: u,
	}
	return defaultSVC
}

func (u *impl) GetUserSpaceList(ctx context.Context, userID int64) (spaces []*entity.Space, err error) {
	spaces, err = u.DomainSVC.GetUserSpaceList(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user %d space list: %w", userID, err)
	}

	return u.lazySyncCompanySpace(ctx, spaces, func() ([]*entity.Space, error) {
		return u.DomainSVC.GetUserSpaceList(ctx, userID)
	}, func(int64) bool {
		return true
	}, fmt.Sprintf("failed to get user %d space list", userID))
}

func (u *impl) GetUserSpaceBySpaceID(ctx context.Context, spaceID []int64) (space []*entity.Space, err error) {
	spaces, err := u.DomainSVC.GetUserSpaceBySpaceID(ctx, spaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user space list for space %v: %w", spaceID, err)
	}
	if len(spaces) == len(spaceID) {
		return spaces, nil
	}

	return u.lazySyncCompanySpace(ctx, spaces, func() ([]*entity.Space, error) {
		return u.DomainSVC.GetUserSpaceBySpaceID(ctx, spaceID)
	}, func(companyID int64) bool {
		return containsResourceID(spaceID, companyID)
	}, fmt.Sprintf("failed to get user space list for space %v", spaceID))
}

func (u *impl) GetUserSpaceRoles(ctx context.Context, userID int64, spaceID []int64) (map[int64]int32, error) {
	roles, err := u.DomainSVC.GetUserSpaceRoles(ctx, userID, spaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user %d space roles: %w", userID, err)
	}
	if len(spaceID) == 0 || len(roles) == len(spaceID) {
		return roles, nil
	}

	companyID, hasCompany, err := fetchCompanyID(ctx)
	if err != nil {
		return nil, err
	}
	if !hasCompany || !containsResourceID(spaceID, companyID) || containsRoleID(roles, companyID) {
		return roles, nil
	}

	if _, err := spacesync.Sync(ctx); err != nil {
		return nil, fmt.Errorf("sync space info failed: %w", err)
	}

	roles, err = u.DomainSVC.GetUserSpaceRoles(ctx, userID, spaceID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user %d space roles after sync: %w", userID, err)
	}
	return roles, nil
}

func containsSpaceID(spaces []*entity.Space, id int64) bool {
	for _, space := range spaces {
		if space != nil && space.ID == id {
			return true
		}
	}
	return false
}

func containsResourceID(ids []int64, id int64) bool {
	for _, resourceID := range ids {
		if resourceID == id {
			return true
		}
	}
	return false
}

func containsRoleID(roles map[int64]int32, id int64) bool {
	_, ok := roles[id]
	return ok
}

func fetchCompanyID(ctx context.Context) (int64, bool, error) {
	company, err := requestyygu.GetCompanyInfo(ctx)
	if err != nil {
		return 0, false, fmt.Errorf("failed to get company info: %w", err)
	}
	if company == nil {
		return 0, false, nil
	}
	return int64(company.ID), true, nil
}

func (u *impl) lazySyncCompanySpace(ctx context.Context, spaces []*entity.Space, reload func() ([]*entity.Space, error), needCompany func(companyID int64) bool, errCtx string) ([]*entity.Space, error) {
	companyID, hasCompany, err := fetchCompanyID(ctx)
	if err != nil {
		return nil, err
	}
	if !hasCompany || !needCompany(companyID) || containsSpaceID(spaces, companyID) {
		return spaces, nil
	}

	if _, err := spacesync.Sync(ctx); err != nil {
		return nil, fmt.Errorf("sync space info failed: %w", err)
	}

	spaces, err = reload()
	if err != nil {
		return nil, fmt.Errorf("%s after sync: %w", errCtx, err)
	}
	return spaces, nil
}

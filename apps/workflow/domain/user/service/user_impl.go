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

package service

import (
	"context"
	"unicode/utf8"

	userEntity "github.com/insmtx/corekg/apps/workflow/domain/user/entity"
	"github.com/insmtx/corekg/apps/workflow/domain/user/internal/dal/model"
	"github.com/insmtx/corekg/apps/workflow/domain/user/repository"
	"github.com/insmtx/corekg/apps/workflow/infra/idgen"
	"github.com/insmtx/corekg/apps/workflow/infra/storage"
	"github.com/insmtx/corekg/apps/workflow/pkg/errorx"
	"github.com/insmtx/corekg/apps/workflow/pkg/lang/conv"
	"github.com/insmtx/corekg/apps/workflow/pkg/lang/ptr"
	"github.com/insmtx/corekg/apps/workflow/pkg/lang/slices"
	"github.com/insmtx/corekg/apps/workflow/types/errno"
)

type Components struct {
	IconOSS   storage.Storage
	IDGen     idgen.IDGenerator
	UserRepo  repository.UserRepository
	SpaceRepo repository.SpaceRepository
}

func NewUserDomain(ctx context.Context, c *Components) User {
	return &userImpl{
		Components: c,
	}
}

type userImpl struct {
	*Components
}

func (u *userImpl) ValidateProfileUpdate(ctx context.Context, req *ValidateProfileUpdateRequest) (
	resp *ValidateProfileUpdateResponse, err error,
) {
	if req.UniqueName == nil && req.Email == nil {
		return nil, errorx.New(errno.ErrUserInvalidParamCode, errorx.KV("msg", "missing parameter"))
	}

	if req.UniqueName != nil {
		uniqueName := ptr.From(req.UniqueName)
		charNum := utf8.RuneCountInString(uniqueName)

		if charNum < 4 || charNum > 20 {
			return &ValidateProfileUpdateResponse{
				Code: UniqueNameTooShortOrTooLong,
				Msg:  "unique name length should be between 4 and 20",
			}, nil
		}

		exist, err := u.UserRepo.CheckUniqueNameExist(ctx, uniqueName)
		if err != nil {
			return nil, err
		}

		if exist {
			return &ValidateProfileUpdateResponse{
				Code: UniqueNameExist,
				Msg:  "unique name existed",
			}, nil
		}
	}

	return &ValidateProfileUpdateResponse{
		Code: ValidateSuccess,
		Msg:  "success",
	}, nil
}

func (u *userImpl) MGetUserProfiles(ctx context.Context, userIDs []int64) (users []*userEntity.User, err error) {
	userModels, err := u.UserRepo.GetUsersByIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}

	users = make([]*userEntity.User, 0, len(userModels))
	for _, um := range userModels {
		// Get image URL
		resURL, err := u.IconOSS.GetObjectUrl(ctx, um.IconURI)
		if err != nil {
			continue // If getting the image URL fails, skip the user
		}

		users = append(users, userPo2Do(um, resURL))
	}

	return users, nil
}

func (u *userImpl) GetUserProfiles(ctx context.Context, userID int64) (user *userEntity.User, err error) {
	userInfos, err := u.MGetUserProfiles(ctx, []int64{userID})
	if err != nil {
		return nil, err
	}

	if len(userInfos) == 0 {
		return nil, errorx.New(errno.ErrUserResourceNotFound, errorx.KV("type", "user"),
			errorx.KV("id", conv.Int64ToStr(userID)))
	}

	return userInfos[0], nil
}

func (u *userImpl) GetUserSpaceList(ctx context.Context, userID int64) (spaces []*userEntity.Space, err error) {
	userSpaces, err := u.SpaceRepo.GetSpaceList(ctx, userID)
	if err != nil {
		return nil, err
	}
	spaceIDs := slices.Transform(userSpaces, func(us *model.SpaceUser) int64 {
		return us.SpaceID
	})

	spaceModels, err := u.SpaceRepo.GetSpaceByIDs(ctx, spaceIDs)
	if err != nil {
		return nil, err
	}
	uris := slices.ToMap(spaceModels, func(sm *model.Space) (string, bool) {
		return sm.IconURI, false
	})

	urls := make(map[string]string, len(uris))
	for uri := range uris {
		url, err := u.IconOSS.GetObjectUrl(ctx, uri)
		if err != nil {
			return nil, err
		}
		urls[uri] = url
	}
	return slices.Transform(spaceModels, func(sm *model.Space) *userEntity.Space {
		return spacePo2Do(sm, urls[sm.IconURI])
	}), nil
}

func (u *userImpl) GetUserSpaceBySpaceID(ctx context.Context, spaceID []int64) (spaces []*userEntity.Space, err error) {
	if len(spaceID) == 0 {
		return nil, errorx.New(errno.ErrUserInvalidParamCode, errorx.KV("msg", "spaceID cannot be empty"))
	}

	spaceModels, err := u.SpaceRepo.GetSpaceByIDs(ctx, spaceID)
	if err != nil {
		return nil, err
	}

	if len(spaceModels) == 0 {
		return []*userEntity.Space{}, nil
	}

	uris := slices.ToMap(spaceModels, func(sm *model.Space) (string, bool) {
		return sm.IconURI, false
	})

	urls := make(map[string]string, len(uris))
	for uri := range uris {
		if uri != "" {
			url, err := u.IconOSS.GetObjectUrl(ctx, uri)
			if err != nil {
				return nil, err
			}
			urls[uri] = url
		}
	}

	return slices.Transform(spaceModels, func(sm *model.Space) *userEntity.Space {
		return spacePo2Do(sm, urls[sm.IconURI])
	}), nil
}

func (u *userImpl) GetUserSpaceRoles(ctx context.Context, userID int64, spaceID []int64) (map[int64]int32, error) {
	userSpaces, err := u.SpaceRepo.GetSpaceList(ctx, userID)
	if err != nil {
		return nil, err
	}

	if len(spaceID) == 0 {
		roles := make(map[int64]int32, len(userSpaces))
		for _, us := range userSpaces {
			roles[us.SpaceID] = us.RoleType
		}
		return roles, nil
	}

	targets := make(map[int64]struct{}, len(spaceID))
	for _, id := range spaceID {
		targets[id] = struct{}{}
	}

	roles := make(map[int64]int32, len(spaceID))
	for _, us := range userSpaces {
		if _, ok := targets[us.SpaceID]; ok {
			roles[us.SpaceID] = us.RoleType
		}
	}

	return roles, nil
}

func spacePo2Do(space *model.Space, iconUrl string) *userEntity.Space {
	return &userEntity.Space{
		ID:          space.ID,
		Name:        space.Name,
		Description: space.Description,
		IconURL:     iconUrl,
		SpaceType:   userEntity.SpaceTypePersonal,
		OwnerID:     space.OwnerID,
		CreatorID:   space.CreatorID,
		CreatedAt:   space.CreatedAt,
		UpdatedAt:   space.UpdatedAt,
	}
}

func userPo2Do(model *model.User, iconURL string) *userEntity.User {
	return &userEntity.User{
		UserID:       model.ID,
		Name:         model.Name,
		UniqueName:   model.UniqueName,
		Email:        model.Email,
		Description:  model.Description,
		IconURI:      model.IconURI,
		IconURL:      iconURL,
		UserVerified: model.UserVerified,
		Locale:       model.Locale,
		CreatedAt:    model.CreatedAt,
		UpdatedAt:    model.UpdatedAt,
	}
}

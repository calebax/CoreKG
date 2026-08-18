package svcapp

import (
	"context"
	"errors"

	"github.com/insmtx/corekg/apps/keapp/models/app"
	"github.com/insmtx/corekg/apps/keapp/models/apptype"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

var (
	ErrAppNameExists   = errors.New("application name exists")
	ErrAppNotFound     = errors.New("application not found")
	ErrCreateAppFailed = errors.New("create application failed")
	ErrUpdateAppFailed = errors.New("update application failed")
	ErrDeleteAppFailed = errors.New("delete application failed")
	ErrNoPermission    = errors.New("no permission")
)

type CreateApplicationRequest struct {
	Uin       uint
	CompanyID uint
	Name      string
	Type      apptype.AppTemplateType
	Desc      string
	Color     string
	Config    apptype.AppConfig
}

type CreateApplicationResponse struct {
	AppID uint
}

func CreateApplication(ctx context.Context, req *CreateApplicationRequest) (*CreateApplicationResponse, error) {
	dao := app.NewApplicationDao()
	exists, err := dao.CheckNameExists(ctx, 0, req.Name, req.CompanyID)
	if err != nil {
		return nil, ErrCreateAppFailed
	}
	if exists {
		return nil, ErrAppNameExists
	}

	entity := &apptype.KeApplication{
		Uin:         req.Uin,
		CompanyID:   req.CompanyID,
		Name:        req.Name,
		Type:        req.Type,
		Status:      apptype.AppStatusDraft,
		Description: req.Desc,
		Color:       req.Color,
		SyncStatus:  apptype.SyncStatusSuccess,
		Config:      req.Config,
	}

	txErr := dbutil.Knownow().WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := dao.WithTx(tx).Insert(ctx, entity); err != nil {
			logs.ErrorContextf(ctx, "[CreateApplication] insert failed: %v", err)
			return err
		}
		scope := &foresttype.KeResourceScope{
			ResourceType: foresttype.ResourceTypeApp,
			ResourceID:   entity.ID,
			ScopeType:    foresttype.ScopeTypeCompany,
			ScopeID:      req.CompanyID,
			Action:       foresttype.ActionManage,
		}
		if err := forest.NewKeResourceScopeDao().WithTx(tx).Insert(ctx, scope); err != nil {
			logs.ErrorContextf(ctx, "[CreateApplication] insert scope failed: %v", err)
			return err
		}
		return nil
	})
	if txErr != nil {
		return nil, ErrCreateAppFailed
	}
	return &CreateApplicationResponse{AppID: entity.ID}, nil
}

type GetApplicationRequest struct {
	AppID     uint
	Uin       uint
	CompanyID uint
}

func GetApplication(ctx context.Context, req *GetApplicationRequest) (*apptype.KeApplication, error) {
	dao := app.NewApplicationDao()
	entity, err := dao.GetByID(ctx, req.AppID)
	if err != nil {
		return nil, ErrAppNotFound
	}
	if entity == nil {
		return nil, ErrAppNotFound
	}
	return entity, nil
}

type ListApplicationsRequest struct {
	Uin       uint
	CompanyID uint
	NameLike  string
	Limit     int
	Offset    int
}

type ListApplicationsResponse struct {
	Items apptype.KeApplicationList
	Total int64
}

func ListApplications(ctx context.Context, req *ListApplicationsRequest) (*ListApplicationsResponse, error) {
	dao := app.NewApplicationDao()
	cond := &app.ApplicationCond{}
	cond.CompanyID = req.CompanyID
	cond.NameLike = req.NameLike
	cond.Limit = req.Limit
	cond.Offset = req.Offset
	items, total, err := dao.GetPageListByCond(ctx, cond)
	if err != nil {
		return nil, errors.New("list applications failed")
	}
	return &ListApplicationsResponse{Items: items, Total: total}, nil
}

type UpdateApplicationRequest struct {
	AppID     uint
	Uin       uint
	CompanyID uint
	Name      *string
	Desc      *string
	Color     *string
	Config    *apptype.AppConfig
}

func UpdateApplication(ctx context.Context, req *UpdateApplicationRequest) error {
	dao := app.NewApplicationDao()
	entity, err := dao.GetByID(ctx, req.AppID)
	if err != nil || entity == nil {
		return ErrAppNotFound
	}
	if req.Name != nil && *req.Name != entity.Name {
		exists, err := dao.CheckNameExists(ctx, req.AppID, *req.Name, req.CompanyID)
		if err != nil {
			return ErrUpdateAppFailed
		}
		if exists {
			return ErrAppNameExists
		}
	}
	updateMap := make(map[string]interface{})
	if req.Name != nil {
		updateMap["name"] = *req.Name
	}
	if req.Desc != nil {
		updateMap["description"] = *req.Desc
	}
	if req.Color != nil {
		updateMap["color"] = *req.Color
	}
	if req.Config != nil {
		updateMap["config"] = req.Config
	}
	if len(updateMap) == 0 {
		return nil
	}
	if err := dao.UpdateMap(ctx, req.AppID, updateMap); err != nil {
		return ErrUpdateAppFailed
	}
	return nil
}

func DeleteApplication(ctx context.Context, appID, uin, companyID uint) error {
	dao := app.NewApplicationDao()
	entity, err := dao.GetByID(ctx, appID)
	if err != nil || entity == nil {
		return ErrAppNotFound
	}
	if err := dao.SoftDelete(ctx, appID); err != nil {
		return ErrDeleteAppFailed
	}
	return nil
}

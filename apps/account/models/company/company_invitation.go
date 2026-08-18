package company

import (
	"context"
	"time"

	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/apps/account/models/perm"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/random"
	"github.com/ygpkg/yg-go/types"
	"gorm.io/gorm"
)

// CreateInvitation 创建邀请
func CreateInvitation(compID, count uint, issuer string, invitationRole accounttype.SysRole, expired time.Time) (*accounttype.CompanyInvitation, error) {
	db := dbutil.Account().Table(accounttype.TableNameCompanyInvitation)
	invitation := accounttype.CompanyInvitation{
		Issuer:         issuer,
		Count:          count,
		CompanyID:      compID,
		Key:            random.Alphabet(16),
		Expired:        expired,
		InvitationRole: invitationRole,
	}
	result := db.Create(&invitation)
	if result.Error != nil {
		return nil, result.Error
	}
	return &invitation, nil
}

// GetInvitationByKey 根据邀请码查询邀请
func GetInvitationByKey(key string) (*accounttype.CompanyInvitation, error) {
	var invitation accounttype.CompanyInvitation
	result := dbutil.Account().
		Table(accounttype.TableNameCompanyInvitation).
		Where("`key` = ?", key).
		First(&invitation)
	if result.Error != nil {
		return nil, result.Error
	}
	return &invitation, nil
}

// UpdateInvitation 更新邀请
func UpdateInvitation(tx *gorm.DB, invitation *accounttype.CompanyInvitation) error {
	result := tx.
		Table(accounttype.TableNameCompanyInvitation).
		Model(&accounttype.CompanyInvitation{}).
		Where("id = ?", invitation.ID).
		Save(invitation)
	return result.Error
}

// CreateInvitationWithPermSet will create an invitation code with pre-set permset,
// this invitation code will grant same perms for users all validly used
// if nil permset is reasonable, try to use CreateInvitation
func CreateInvitationWithPermSet(ctx context.Context,
	compID, uin, count uint,
	issuer string, invitationRole accounttype.SysRole,
	expired time.Time, ps *perm.Set,
	departmentIDs types.UintArray) (*accounttype.CompanyInvitation, error) {
	if len(ps.ForestPs)+len(ps.ChatPs) == 0 {
		logs.WarnContextf(ctx, ""+
			"CreateInvitationWithPermSet try to create an invitation without any permset")
	}

	db := dbutil.Account().Table(accounttype.TableNameCompanyInvitation)
	invitation := accounttype.CompanyInvitation{
		Issuer:         issuer,
		Count:          count,
		CompanyID:      compID,
		Key:            random.Alphabet(16),
		Expired:        expired,
		InvitationRole: invitationRole,
		PermSet:        ps,
		Uin:            uin,
		DepartmentIDs:  departmentIDs,
	}
	result := db.Create(&invitation)
	if result.Error != nil {
		return nil, result.Error
	}
	return &invitation, nil
}

func GetInviteByKey(key string) (*accounttype.CompanyInvitation, error) {
	var inv *accounttype.CompanyInvitation
	if err := dbutil.Account().
		Where("`key` = ?", key).
		Find(&inv).Error; err != nil {
		return nil, err
	}
	return inv, nil
}

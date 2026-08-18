package user

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/mozillazg/go-pinyin"
	"github.com/insmtx/corekg/apps/account/models/accounttype"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/silenceper/wechat/v2/officialaccount/oauth"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/random"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// ExistsUserByPhone 判断是否存在个人账号
func ExistsUserByPhone(ctx context.Context, phone string) (bool, error) {
	var count int64
	err := dbutil.Account().
		Table(accounttype.TableNameUser).
		WithContext(ctx).
		Joins("JOIN user_identification uin ON uin.user_id = user.id").
		Where("user.phone = ?", phone).
		Where("uin.uin_status = ?", accounttype.UinStatusNormal).
		Count(&count).Error
	if err != nil {
		logs.ErrorContextf(ctx, "GetPhoneByUin: get user failed, %v", err)
	}
	return count > 0, nil
}

// UpdateUserPhoneByID 更新用户手机号
func UpdateUserPhoneByID(id uint, phone string) error {
	return dbutil.Account().
		Table(accounttype.TableNameUser).
		Where("id = ?", id).
		Update("phone", phone).Error
}

// CheckUnionIDExist 检查Unionid是否存在
func CheckUnionIDExist(ctx context.Context, unionID string) (bool, error) {
	var count int64
	err := dbutil.Account().
		Table(accounttype.TableNameUser+" u ").
		WithContext(ctx).
		Joins("JOIN user_identification uin ON uin.user_id = u.id").
		Where("u.wechat_union_id = ?", unionID).
		Where("uin.uin_status = ?", accounttype.UinStatusNormal).
		Count(&count).Error
	if err != nil {
		logs.ErrorContextf(ctx, "CheckUnionIDExist: get user failed, %v", err)
	}
	return count > 0, nil
}

// UpdateUserWcAndNameByID 更新用户unionid
func UpdateUserWcAndNameByID(id uint, uinfo *oauth.UserInfo) error {
	return dbutil.Account().
		Table(accounttype.TableNameUser).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"name":               uinfo.Nickname,
			"avatar_url":         uinfo.HeadImgURL,
			"identify":           strings.Join(pinyin.LazyPinyin(uinfo.Nickname, pinyin.NewArgs()), "") + random.String(3),
			"wechat_union_id":    uinfo.Unionid,
			"wechat_web_open_id": uinfo.OpenID,
		}).Error
}

func UpdateAccountPassword(userID uint, password string) error {
	db := dbutil.Account()
	var user struct {
		ID uint `json:"id"`
	}
	// 先查找用户是否存在
	if err := db.Table(accounttype.TableNameUser).
		Where("id = ?", userID).
		First(&user).Error; err != nil {

		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("user not found with ID: %d", userID)
		}
		return fmt.Errorf("failed to query user: %w", err)
	}

	crypedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// 更新密码
	if err := db.Table(accounttype.TableNameUser).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"password":         crypedPassword,
			"password_changed": 1,
		}).Error; err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	return nil
}

func RedisKeyLoginAccount(accountName string) string {
	return fmt.Sprintf("account:loginByPasswordAccount:%s", accountName)
}

func RedisKeyLoginByUsername(username string) string {
	return fmt.Sprintf("account:loginByPasswordUsername:%s", username)
}

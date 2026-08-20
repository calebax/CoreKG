package license

import (
	"context"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	license2 "github.com/insmtx/corekg/apps/corekg/models/license"
	"github.com/insmtx/corekg/pkgs/platform/admintype"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/ygpkg/yg-go/apis/apiobj"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

// GenerateLicense creates a signed license
func GenerateLicense(ctx context.Context, license *admintype.License) error {
	privateKey, err := license2.ParsePrivateKey(license.PrivateKey)
	if err != nil {
		return err
	}

	meta := &admintype.Meta{
		CreatedAt:  time.Now(),
		Serial:     license.Serial,
		Subject:    license.Subject,
		Env:        license.Env,
		UID:        license.UID,
		Issuer:     license.Issuer,
		ExpiredAt:  *license.ExpiredAt,
		Seed:       uuid.NewString(),
		VersionKey: license.VersionKey,
	}
	jsonData, err := json.Marshal(meta)
	if err != nil {
		logs.ErrorContextf(ctx, "GenerateLicense marshal[%v] failed: %v", meta, err)
		return err
	}

	hashed := sha256.Sum256(jsonData)
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return err
	}

	encodedSignature := base64.StdEncoding.EncodeToString(signature)
	encodedData := base64.StdEncoding.EncodeToString(jsonData)
	licenseString := fmt.Sprintf("%s.%s", encodedSignature, encodedData)

	license.Meta = *meta
	license.Raw = licenseString

	return nil
}

// Item license信息结构体
type Item struct {
	//license id
	ID uint `json:"id"`
	//创建时间
	CreatedAt time.Time `json:"created_at"`
	//序列号
	Serial string `json:"serial"`
	//环境类型
	Env license2.EnvType `json:"env"`
	//环境唯一标识
	UID string `json:"uid"`
	//主体
	Subject string `json:"subject"`
	//签发人
	Issuer string `json:"issuer"`
	//到期时间
	ExpiredAt time.Time `json:"expired_at"`
	//备注
	Note string `json:"note"`
	//版本类型
	VersionKey string `json:"version_key"`
}

// CreateLicense 创建License
func CreateLicense(tx *gorm.DB, license *admintype.License) error {
	return tx.Create(license).Error
}

type ListLicenseResponse struct {
	apiobj.QueryResponse
	Data []*Item `json:"data"`
}

// QueryLicenseList 查询License列表
func QueryLicenseList(ctx context.Context, opt apiobj.PageQuery, resp *ListLicenseResponse) error {
	query := dbutil.Account().
		Table(admintype.TableNameLicense).
		Select("id, created_at, serial,env, uid, subject,issuer, expired_at, note, version_key").
		Where("deleted_at is null")

	// 添加过滤器
	for _, filter := range opt.Filters {
		f := filter.Field
		switch f {
		case "serial", "env", "uid":
			query = query.Where(fmt.Sprintf("%s = ?", f), filter.Value[0])
		case "subject", "issuer":
			query = query.Where(fmt.Sprintf("%s like ?", f), fmt.Sprintf("%%%s%%", filter.Value[0]))
		case "expired_at":
			query = query.Where("expired_at > ?", filter.Value[0])
		case "version_key":
			query = query.Where("version_key like ?", fmt.Sprintf("%%%s%%", filter.Value[0]))
		default:
			logs.ErrorContextf(ctx, "[admin][QueryLicenseList] invalid filter field: %s", filter.Field)
			return fmt.Errorf("invalid filter field: %s", filter.Field)
		}
	}

	if !opt.BeginTime.IsZero() {
		query = query.Where("created_at >= ?", opt.BeginTime)
	}
	if !opt.EndTime.IsZero() {
		query = query.Where("created_at <= ?", opt.EndTime)
	}

	if err := query.Count(&resp.Total).Error; err != nil {
		return err
	}
	if resp.Total == 0 {
		return nil
	}
	resp.Limit = opt.Limit
	resp.Offset = opt.Offset

	if len(opt.OrderBy) > 0 {
		query = query.Order(strings.Join(opt.OrderBy, ","))
	}

	// 分页逻辑
	query = query.Offset(opt.Offset)
	if !opt.ListAll && opt.Limit > 0 {
		query = query.Limit(opt.Limit)
	}
	err := query.Find(&resp.Data).Error
	if err != nil {
		return err
	}
	return nil
}

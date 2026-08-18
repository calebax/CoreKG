package wecoms

import (
	"context"
	"strconv"

	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/xen0n/go-workwx"
	"github.com/ygpkg/yg-go/config"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
	"gorm.io/gorm"
)

type App struct {
	CompanyID string `gorm:"column:company_id;type:varchar(18);index:idx_app,priority:1,unique" json:"company_id" yaml:"company_id"`
	AppID     string `gorm:"column:app_id;index:idx_app,priority:2,unique" json:"app_id" yaml:"app_id"`

	Name           string `gorm:"column:name;type:varchar(18)" json:"name" yaml:"name"`
	Secret         string `gorm:"column:secret;type:varchar(64)" json:"secret" yaml:"secret"`
	Token          string `gorm:"column:token;type:varchar(32)" json:"token" yaml:"token"`
	EncodingAESKey string `gorm:"column:encoding_aes_key;type:varchar(48)" json:"encoding_aes_key" yaml:"encoding_aes_key"`
}

func (*App) TableName() string { return TableNameApp }

func (cs App) WxCli() *workwx.WorkwxApp {
	appid, _ := strconv.ParseInt(cs.AppID, 10, 64)
	return workwx.New(cs.CompanyID).WithApp(cs.Secret, appid)
}

func (cs App) ToConfig() config.WecomApp {
	agentId, err := strconv.ParseInt(cs.AppID, 10, 64)
	if err != nil {
		return config.WecomApp{}
	}
	return config.WecomApp{
		Name:           cs.Name,
		CompanyID:      cs.CompanyID,
		AgentID:        agentId,
		Secret:         cs.Secret,
		Token:          cs.Token,
		EncodingAESKey: cs.EncodingAESKey,
	}
}

func GetApp(companyid, appid string) (*App, error) {
	sec := &App{}
	err := dbutil.Account().Table(TableNameApp).
		Where("company_id = ? AND app_id = ?", companyid, appid).
		Find(sec).Error
	if err != nil {
		return nil, err
	}
	if sec.CompanyID == "" {
		return nil, gorm.ErrRecordNotFound
	}
	return sec, nil
}

func GetWxCli(app config.WecomApp) *workwx.WorkwxApp {
	return workwx.New(app.CompanyID).WithApp(app.Secret, app.AgentID)
}

func GetWxCliFromSetting(group, key string) (*workwx.WorkwxApp, error) {
	app := config.WecomApp{}
	ctx := context.TODO()
	err := settings.GetYaml(group, key, &app)
	if err != nil {
		logs.ErrorContextf(ctx, "failed to get chatgpt config: %s, %s", group, key)
		return nil, err
	}

	return workwx.New(app.CompanyID).WithApp(app.Secret, app.AgentID), nil
}

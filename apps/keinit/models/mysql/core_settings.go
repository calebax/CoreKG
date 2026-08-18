package mysql

import (
	"context"
	"os"

	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
	"gopkg.in/yaml.v3"
	"gorm.io/gorm"
)

func UpsertCoreSettings(ctx context.Context, db *gorm.DB, settingFile string) error {
	f, err := os.Open(settingFile)
	if err != nil {
		logs.ErrorContextf(ctx, "open setting file (%s) error: %v", settingFile, err)
		return err
	}
	defer f.Close()
	settingList := make([]*settings.SettingItem, 0)
	err = yaml.NewDecoder(f).Decode(&settingList)
	if err != nil {
		logs.ErrorContextf(ctx, "decode setting file (%s) error: %v", settingFile, err)
		return err
	}
	for _, v := range settingList {
		err = settings.UpsertSetting(v)
		if err != nil {
			logs.ErrorContextf(ctx, "UpsertSetting setting file (%s) error: %v", settingFile, err)
			return err
		}
	}
	// err = settings.Updates(settingList...)
	// if err != nil {
	// 	logs.ErrorContextf(ctx, "update setting file (%s) error: %v", settingFile, err)
	// 	return err
	// }
	return nil
}

func ListCoreSettings(ctx context.Context, db *gorm.DB) ([]*settings.SettingItem, error) {
	list := make([]*settings.SettingItem, 0)
	db.Find(&list)
	return list, nil
}

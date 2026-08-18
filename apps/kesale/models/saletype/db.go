package saletype

import (
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

const (
	// TableNamePrefix 表名前缀
	TableNamePrefix            = "sale_"
	TableNameSaleOrder         = TableNamePrefix + "order"
	TableNameSaleOrderItem     = TableNamePrefix + "order_item"
	TableNameSalePaymentRecord = TableNamePrefix + "payment_record"
)

func InitDB(db *gorm.DB) error {
	err := dbtools.InitModel(db,
		&SaleOrder{},
		&SaleOrderItem{},
		&SalePaymentRecord{},
	)
	if err != nil {
		logs.Errorf("[main] init sale database failed, %s", err)
		return err
	}
	logs.Info("[main] init sale database success")
	{
		// db_pre
		if err := presetDatabase(); err != nil {
			return err
		}
	}
	return nil
}

// 数据库初始化准备
func presetDatabase() error {
	return nil
}

package mutex

import "gorm.io/gorm"

const tableNameLockInfo = "sys_cluster_lock"

type lockInfo struct {
	gorm.Model

	Namespace string `gorm:"column:namespace;type:varchar(15);uniqueIndex:uni_lockname;"`
	Name      string `gorm:"column:name;type:varchar(15);uniqueIndex:uni_lockname;"`
	Master    string `gorm:"column:master;type:varchar(32)"`

	Version   int64 `gorm:"column:version"`
	ExpiredAt int64 `gorm:"column:expired_at"`
}

// TableName .
func (lockInfo) TableName() string {
	return tableNameLockInfo
}

// getLockInfo .
func getLockInfo(db *gorm.DB, namespace, name string) (*lockInfo, error) {
	li := &lockInfo{}
	err := db.Table(tableNameLockInfo).
		Where("namespace = ? AND name = ?", namespace, name).
		Find(li).Error
	if err != nil {
		return nil, err
	}
	return li, nil
}

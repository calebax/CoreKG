package mutex

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/insmtx/corekg/pkgs/utils/nettools"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

var _ sync.Locker = (*DBLocker)(nil)

// TODO
// DBLocker .
type DBLocker struct {
	db *gorm.DB

	lockTimeout       time.Duration
	heartbeatInterval time.Duration

	local  *lockInfo
	remote *lockInfo
}

// NewDBLocker TODO
func NewDBLocker(db *gorm.DB, namespace, name, key string) (*DBLocker, error) {
	l := &DBLocker{
		db: db,
		local: &lockInfo{
			Namespace: namespace,
			Name:      name,
			Master:    key,
		},
	}

	err := db.AutoMigrate(&lockInfo{})
	if err != nil {
		logs.ErrorContextf(context.TODO(), "[mutex_db] auto migrate lockinfo failed, %s", err)
		return nil, err
	}

	return l, nil
}

// Lock .
func (l *DBLocker) Lock() {

}

// Unlock .
func (l *DBLocker) Unlock() {

}

func (l *DBLocker) waitLocker() {

}

// ProcessKey .
func ProcessKey() string {
	hostname, _ := os.Hostname()
	ip := nettools.MustLocalPrimearyIP()
	return fmt.Sprintf("%s(%s):%v", hostname, ip, os.Getpid())
}

package dbutil

import (
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"gorm.io/gorm"
)

func Account() *gorm.DB {
	return dbtools.DB("account")
}

func Core() *gorm.DB {
	return dbtools.Core()
}

func AIGC() *gorm.DB {
	return dbtools.DB("aigc")
}

func Zhy() *gorm.DB {
	return dbtools.DB("zhy")
}

func Chat() *gorm.DB {
	return dbtools.DB("chat")
}

func Cook() *gorm.DB {
	return dbtools.DB("cook")
}

func OA() *gorm.DB {
	return dbtools.DB("oa")
}

func LLM() *gorm.DB {
	return dbtools.DB("llm")
}

func Cluster() *gorm.DB {
	return dbtools.DB("cluster")
}

func Knownow() *gorm.DB {
	return dbtools.DB("knownow")
}

func Dryang() *gorm.DB {
	return dbtools.DB("dryang")
}

func Aiagents() *gorm.DB {
	return dbtools.DB("aiagents")
}

func OPO() *gorm.DB { return dbtools.DB("opo") }

func OPOTemp() *gorm.DB {
	return dbtools.DB("opo_temp")
}

func APIGateway() *gorm.DB {
	return dbtools.DB("apigateway")
}

func Coze() *gorm.DB {
	return dbtools.DB("coze")
}

func Sale() *gorm.DB {
	return dbtools.DB("sale")
}

func GetDB(name, dsn string) (*gorm.DB, error) {
	if dbtools.DBExists(name) {
		return dbtools.DB(name), nil
	}
	mysqlURL := RawDSNToURL(dsn)
	if err := dbtools.InitMultiDBConn(map[string]string{name: mysqlURL}); err != nil {
		return nil, err
	}
	return dbtools.DB(name), nil
}

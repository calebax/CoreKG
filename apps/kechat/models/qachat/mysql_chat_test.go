package qachat

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/pkgs/testutils"
	"github.com/stretchr/testify/assert"
)

func TestIsWriteOperationSQL(t *testing.T) {
	sqlList := []string{
		"update table1 set name = 'yg'",
		"update table1 set name = 'yg' where id = 1",
		"update table1 set name = 'yg' where id = 1 and name = 'yg'",
		"delete from table1 where id = 1",
		"select * from table1 where update = 1",
		"select * from table1 where deleted_at is null",
		"DROP TABLE table1",
	}
	w := &ChatWapper{
		ctx: &gin.Context{},
	}
	var res []bool
	for _, v := range sqlList {
		res = append(res, w.isWriteOperationSQL(v))
	}
	t.Log(res)
}

func TestChoiceTableList(t *testing.T) {
	testutils.Initialize(testutils.AppNameKechat)
	defer testutils.Close()
	ctx := testutils.NewCtx()
	w := &ChatWapper{
		ctx: ctx,
	}
	ddlMap := map[string]string{
		"users":   "CREATE TABLE users ( id bigint PRIMARY KEY, name varchar(255));",
		"orders":  "CREATE TABLE orders ( id bigint PRIMARY KEY, user_id bigint, amount decimal(10,2));",
		"company": "CREATE TABLE company ( id bigint PRIMARY KEY, name varchar(255));",
	}
	tables, err := w.ChoiceTableList("查询用户的订单信息", ddlMap)
	assert.Nil(t, err)
	t.Log(tables)
}

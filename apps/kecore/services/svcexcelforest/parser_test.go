package svcexcelforest

import (
	"testing"

	"github.com/cloudwego/eino/schema"
	"github.com/stretchr/testify/assert"
)

func TestDetectColumnTypes(t *testing.T) {
	if true {
		t.Skip("skip test")
		return
	}
	headers := []string{"姓名", "性别", "年龄", "薪资", "入职日期", "是否在职"}

	sampleData := []*schema.Document{
		{
			ID:      "1",
			Content: "张三\t男\t30\t50000.5\t2023-01-15\ttrue",
		},
		{
			ID:      "2",
			Content: "李四\t女\t25\t$60,000\t2023年02月01日\t是",
		},
		{
			ID:      "3",
			Content: "王五\t男\t28\t55000\t2023/03/10\tyes",
		},
	}

	// 分析列类型
	columnTypes, err := DetectColumnTypes(headers, sampleData)
	assert.Nil(t, err)
	t.Log(columnTypes)
}

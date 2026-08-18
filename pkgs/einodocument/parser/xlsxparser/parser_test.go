package xlsxparser

import (
	"context"
	"os"
	"testing"

	"github.com/cloudwego/eino/schema"
	"github.com/stretchr/testify/assert"
)

func TestParser3466(t *testing.T) {
	f, err := os.Open("./testdata/3466_sh.xlsx")
	assert.NoError(t, err)

	ctx := context.Background()
	epr := &ExcelParser{}
	docs, err := epr.Parse(ctx, f, WithSheetName("“6+2”主要指标"))
	assert.NoError(t, err)
	docsGroup := make(map[string][]*schema.Document)
	for _, doc := range docs {
		subSheetIDVal := doc.MetaData[SubSheetID]
		subSheetID, ok := subSheetIDVal.(string)
		if !ok {
			t.Logf("subSheetIDVal is not string, %v", subSheetIDVal)
			continue
		}

		docsGroup[subSheetID] = append(docsGroup[subSheetID], doc)
	}
	sub := docsGroup["“6+2”主要指标_1"]
	t.Log(sub)
	t.Log(docs)
}

func TestExcelParserYG4431(t *testing.T) {
	f, err := os.Open("./testdata/言古研发_需求_20250914104431.xlsx")
	assert.NoError(t, err)

	ctx := context.Background()
	epr := &ExcelParser{}
	docs, err := epr.Parse(ctx, f)
	assert.NoError(t, err)
	docsGroup := make(map[int][]*schema.Document)
	for _, doc := range docs {
		subSheetIDVal := doc.MetaData[SubSheetID]
		subSheetID, ok := subSheetIDVal.(int)
		if !ok {
			t.Logf("subSheetIDVal is not string, %v", subSheetIDVal)
			continue
		}

		docsGroup[subSheetID] = append(docsGroup[subSheetID], doc)
	}
	t.Log(docs)

}

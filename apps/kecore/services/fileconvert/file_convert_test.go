package fileconvert

import (
	"strings"
	"testing"

	"github.com/insmtx/corekg/pkgs/testutils"
	"github.com/insmtx/corekg/pkgs/utils/dbutil"
	"github.com/stretchr/testify/assert"
	"github.com/ygpkg/yg-go/storage"
)

func TestConvertFileIfNeeded(t *testing.T) {
	testutils.Initialize(testutils.AppNameKecore)
	defer testutils.Close()

	ctx := testutils.NewCtx(testutils.WithUin(384))

	fileID := uint(72609)

	coreFile, err := storage.GetFileByID(dbutil.Core(), fileID)
	if err != nil {
		t.Skipf("文件不存在或查询失败: %v", err)
	}

	t.Logf("原始文件: ID=%d, Filename=%s, Ext=%s, StoragePath=%s",
		coreFile.ID, coreFile.Filename, coreFile.FileExt, coreFile.StoragePath)

	result, err := ConvertFileIfNeeded(ctx, coreFile)
	if err != nil {
		t.Fatalf("转换失败: %v", err)
	}

	t.Logf("转换后: ID=%d, Filename=%s, Ext=%s, StoragePath=%s",
		result.ID, result.Filename, result.FileExt, result.StoragePath)

	assert.NotNil(t, result)
}

func TestIsStrictWorkbookXML(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    bool
	}{
		{
			name:    "strict namespace",
			content: `<workbook xmlns="http://purl.oclc.org/ooxml/spreadsheetml/main"></workbook>`,
			want:    true,
		},
		{
			name:    "transitional namespace",
			content: `<workbook xmlns="http://schemas.openxmlformats.org/spreadsheetml/2006/main"></workbook>`,
			want:    false,
		},
		{
			name:    "unknown namespace",
			content: `<workbook xmlns="urn:unknown"></workbook>`,
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := isStrictWorkbookXML(strings.NewReader(tt.content))
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

package mysql

import (
	"bytes"
	"context"
	"os"
	"text/template"

	"github.com/insmtx/corekg/apps/keinit/models/helper"
	"github.com/ygpkg/yg-go/logs"
	"gorm.io/gorm"
)

// ExecuteVariableTpl 执行变量替换
func ExecuteVariableTpl(ctx context.Context, db *gorm.DB, tplFile string, envs map[string]string) error {
	tpl, err := os.ReadFile(tplFile)
	if err != nil {
		logs.ErrorContextf(ctx, "read tpl file (%s) error: %v", tplFile, err)
		return err
	}

	t, err := template.New("sqltpl").Parse(string(tpl))
	if err != nil {
		logs.ErrorContextf(ctx, "parse tpl file (%s) error: %v", tplFile, err)
		return err
	}

	var buf bytes.Buffer
	err = t.Execute(&buf, envs)
	if err != nil {
		logs.ErrorContextf(ctx, "execute tpl file (%s) error: %v", tplFile, err)
		return err
	}

	n, err := ExecuteSQLReader(ctx, db, &buf, 0)
	if err != nil {
		logs.ErrorContextf(ctx, "execute sql tpl (%s:%v) error: %v", tplFile, n, err)
		return err
	}
	return nil
}

// ListVariableTplFiles 返回需要替换变量的文件
func ListVariableTplFiles(ctx context.Context, sqlDir string) ([]string, error) {
	return helper.ListFileWithExt(ctx, sqlDir, ".sqltpl")
}

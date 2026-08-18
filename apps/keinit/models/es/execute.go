package es

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/insmtx/corekg/apps/keinit/models/helper"
	"github.com/insmtx/corekg/apps/keinit/models/mysql"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"github.com/ygpkg/yg-go/logs"
)

// ExecuteDSLFile executes a DSL file
func ExecuteDSLFile(ctx context.Context, esCli *elasticsearch.Client, path string, body string, method string) {
	// esapi.GetRequest{
	// req := esapi.Request{
	// 	Method: method,
	// 	Path:   path,
	// 	Body:   bytes.NewReader([]byte(body)),
	// }
}

func ExecuteDSLWithFile(ctx context.Context, esCli *elasticsearch.Client, dslDir string) error {
	dslFiles, err := helper.ListFileWithExt(ctx, dslDir, ".dsl")
	if err != nil {
		logs.ErrorContextf(ctx, "list dsl files error: %v", err)
		return err
	}
	// sort.StringSlice(dslFiles).Sort()
	for _, dslFile := range dslFiles {
		rcd, err := mysql.LastExecRecourd(dbtools.Core(), filepath.Base(dslFile))
		if err != nil {
			logs.ErrorContextf(ctx, "get last exec record error: %v", err)
			return err
		}
		if rcd != nil && rcd.ExecStatus == mysql.ExecStatusSuccess {
			continue
		}
		err = doExecuteDSL(ctx, esCli, dslFile)
		if err != nil {
			logs.ErrorContextf(ctx, "execute dsl error: %v", err)
			return err
		}
		logs.InfoContextf(ctx, "execute dsl file success: %s", dslFile)
	}

	return err
}

// ExecuteDSL executes a DSL file
func ExecuteDSL(ctx context.Context, esCli *elasticsearch.Client, path string, body string, method string) error {
	// 构造 HTTP 请求
	req, err := http.NewRequest(method, path, bytes.NewReader([]byte(body)))
	if err != nil {
		logs.ErrorContextf(ctx, "build request error: %v", err)
		return fmt.Errorf("构造请求失败: %w", err)
	}

	// 设置 Content-Type
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// 执行请求
	res, err := esCli.Transport.Perform(req.WithContext(ctx))
	if err != nil {
		logs.ErrorContextf(ctx, "perform request error: %v", err)
		return fmt.Errorf("请求失败: %w", err)
	}
	defer res.Body.Close()

	// 读取响应
	respBody, err := io.ReadAll(res.Body)
	if err != nil {
		logs.ErrorContextf(ctx, "read response error: %v", err)
		return fmt.Errorf("读取响应失败: %w", err)
	}

	if res.StatusCode >= 300 {
		logs.ErrorContextf(ctx, "request error: %s", respBody)
		return fmt.Errorf("ES 请求失败: %s", respBody)
	}

	logs.InfoContextf(ctx, "ES 请求成功, 状态码: %d, 响应: %s", res.StatusCode, string(respBody))
	// fmt.Printf("ES 请求成功, 状态码: %d, 响应: %s\n", res.StatusCode, string(respBody))
	return nil
}

func doExecuteDSL(ctx context.Context, esCli *elasticsearch.Client, dslFile string) error {
	rcd := &mysql.InitExecRecourd{
		FileName:   filepath.Base(dslFile),
		StartTime:  time.Now(),
		ExecStatus: mysql.ExecStatusFailed,
	}
	defer func() {
		rcd.EndTime = time.Now()
		rcd.ExecTime = rcd.EndTime.Sub(rcd.StartTime).Seconds()
		if err := dbtools.Core().Save(rcd).Error; err != nil {
			logs.ErrorContextf(ctx, "save init exec record error: %v", err)
		}
	}()

	dsls, err := parseMultipleDSLFile(dslFile)
	if err != nil {
		logs.ErrorContextf(ctx, "parse dsl file error: %v", err)
		return fmt.Errorf("解析 DSL 文件失败: %w", err)
	}
	for _, v := range dsls {
		if err = ExecuteDSL(ctx, esCli, v.Path, v.Body, v.Method); err != nil {
			logs.ErrorContextf(ctx, "execute dsl error: %v", err)
			return err
		}
	}
	rcd.ExecStatus = mysql.ExecStatusSuccess
	rcd.EndTime = time.Now()
	rcd.ExecTime = rcd.EndTime.Sub(rcd.StartTime).Seconds()
	logs.InfoContextf(ctx, "execute dsl file success: %s", dslFile)
	return nil
}

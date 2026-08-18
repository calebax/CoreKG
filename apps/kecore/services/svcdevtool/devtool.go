package svcdevtool

import (
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/kecore/internal/dto/dtodevtool"
	"github.com/insmtx/corekg/apps/kecore/models/fs"
	"github.com/insmtx/corekg/apps/kecore/services/devtool"
	"github.com/ygpkg/yg-go/logs"
)

func RewriteMarkdownURL(ctx *gin.Context, req *dtodevtool.RewriteMarkdownURLRequest) (res *dtodevtool.RewriteMarkdownURLResponse, err error) {
	res = &dtodevtool.RewriteMarkdownURLResponse{}

	// 获取请求参数
	path := req.Request.Path
	oldPrefix := req.Request.OldPrefix
	newPrefix := req.Request.NewPrefix
	bizBucket := req.Request.BizBucket
	service := req.Request.Service

	// 创建临时目录用于存放下载的文件
	tmpDir := filepath.Join(os.TempDir(), "rewrite-markdown-url")
	defer func() {
		// 清理临时目录
		if err := os.RemoveAll(tmpDir); err != nil {
			logs.WarnContextf(ctx, "failed to cleanup temp directory %s: %v", tmpDir, err)
		}
	}()

	// 准备修改器参数
	args := devtool.ModifierArgs{
		"oldPrefix": oldPrefix,
		"newPrefix": newPrefix,
		"bizBucket": bizBucket,
		"service":   service,
	}

	// 使用 ProcessFilesWithModification 和 ReplaceURLModifier 处理文件
	count, err := devtool.ProcessFilesWithModification(ctx, path, ".md", tmpDir, devtool.ReplaceURLModifier(ctx), args)
	if err != nil {
		logs.ErrorContextf(ctx, "ProcessFilesWithModification failed: %v", err)
		return nil, err
	}

	logs.InfoContextf(ctx, "RewriteMarkdownURL: successfully processed %d files in path %s", count, path)

	return res, nil
}

func StatAlgoMarkdown(ctx *gin.Context, req *dtodevtool.StatAlgoMarkdownRequest) (res *dtodevtool.StatAlgoMarkdownResponse, err error) {
	res = &dtodevtool.StatAlgoMarkdownResponse{}

	// 获取请求中的路径
	path := req.Request.Path
	if path == "" {
		// 如果没有指定路径，使用默认的算法路径前缀
		path = fs.PurposeForestAlgo
	}

	// 统计 markdown 文件数量
	count, err := devtool.CountFiles(ctx.Request.Context(), path, ".md")
	if err != nil {
		logs.ErrorContextf(ctx, "CountMarkdownFiles failed: %v", err)
		return nil, err
	}

	// 设置响应
	res.Response.Count = count
	res.Response.Path = path

	return res, nil
}

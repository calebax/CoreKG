package chunk

import (
	"bytes"
	"context"
	"io"
	"net/http"

	"github.com/ygpkg/yg-go/dbtools/esquery"
	"github.com/ygpkg/yg-go/logs"
)

// DisableFileChunk 禁用或启用文件分片
func DisableFileChunk(ctx context.Context, fileID uint, isDisable bool) error {
	query := esquery.NewBuilder().
		SetQuery(esquery.BuildMap("term", esquery.BuildMap("file_id", fileID))).
		Set("script", esquery.BuildMap("source", "ctx._source.is_disable = params.is_disable",
			"lang", "painless",
			"params", esquery.BuildMap("is_disable", isDisable)))
	querybyte, err := query.BuildBytes()
	if err != nil {
		logs.ErrorContextf(ctx, "query build error: %v", err)
		return err
	}
	logs.InfoContextf(ctx, "UpdateChunkFileName querybyte: %v", string(querybyte))
	resp, err := escli.UpdateByQuery(
		[]string{"ke_0"},
		escli.UpdateByQuery.WithBody(bytes.NewBuffer(querybyte)),
		escli.UpdateByQuery.WithContext(ctx),
	)
	if err != nil {
		logs.ErrorContextf(ctx, "UpdateChunkFileName error: %v", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logs.ErrorContextf(ctx, "UpdateChunkFileName error: %v", string(body))
		return err
	}
	return nil
}

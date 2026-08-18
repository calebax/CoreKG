package essearch

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/ygpkg/yg-go/config"
	"github.com/ygpkg/yg-go/dbtools/esquery"
	"github.com/ygpkg/yg-go/dbtools/estool"
	"github.com/ygpkg/yg-go/logs"
	"github.com/ygpkg/yg-go/settings"
)

var escli *elasticsearch.Client

func InitESClient(ctx context.Context) (*elasticsearch.Client, error) {
	if escli != nil {
		return escli, nil
	}
	cfg := config.ESConfig{}
	err := settings.GetYaml("knowledge", "es", &cfg)
	if err != nil {
		logs.ErrorContextf(ctx, "get es config failed: %s", err)
		return nil, err
	}
	client, err := estool.InitES(cfg)
	if err != nil {
		logs.ErrorContextf(ctx, "init es client failed: %s", err)
		return nil, err
	}
	escli = client
	return client, nil
}

// DeleteFileReferences 根据文件id删除对应的references
func DeleteFileReferences(ctx context.Context, index_name string, fileIds []uint) error {
	client, err := InitESClient(ctx)
	if err != nil {
		return err
	}
	sourceStr := `
      if (params.file_ids_to_remove.contains(ctx._source.file_id)) {
        ctx.op = 'delete';
      } else if (ctx._source.references != null) {
        ArrayList updated = new ArrayList();
        for (item in ctx._source.references) {
          if (!params.file_ids_to_remove.contains(item.file_id)) {
            updated.add(item);
          }
        }
        if (updated.isEmpty()) {
          ctx.op = 'delete';
        } else {
          ctx._source.references = updated;
        }
      }
	`
	query := esquery.NewBuilder().
		SetQuery(esquery.BuildMap("bool", esquery.BuildMap("should", []esquery.Map{
			esquery.BuildMap("terms", esquery.BuildMap("file_id", fileIds)),
			esquery.BuildMap("nested",
				esquery.BuildMap("path", "references",
					"query", esquery.BuildMap("terms", esquery.BuildMap("references.file_id", fileIds)))),
		}))).
		Set("script",
			esquery.BuildMap(
				"source", sourceStr,
				"lang", "painless",
				"params",
				esquery.BuildMap("file_ids_to_remove", fileIds)))

	querybyte, err := query.BuildBytes()
	if err != nil {
		return err
	}
	resp, err := client.UpdateByQuery(
		[]string{index_name},
		client.UpdateByQuery.WithBody(bytes.NewBuffer(querybyte)),
		client.UpdateByQuery.WithContext(context.Background()),
	)
	if err != nil {
		return err
	}
	// 打印返回结果
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		bodyerr, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("es query failed: %s err: %s", resp.Status(), bodyerr)
	}
	return nil
}

// DeleteForest 根据forest_id删除对应的forest
func DeleteForest(ctx context.Context, index_name string, forest_id uint) error {
	client, err := InitESClient(ctx)
	if err != nil {
		return err
	}
	query := esquery.NewBuilder().
		SetQuery(esquery.BuildMap("term", esquery.BuildMap("forest_id", forest_id)))
	querybyte, err := query.BuildBytes()
	if err != nil {
		return err
	}
	resp, err := client.DeleteByQuery(
		[]string{index_name},
		bytes.NewBuffer(querybyte),
		client.DeleteByQuery.WithContext(context.Background()),
	)
	if err != nil {
		return err
	}
	// 打印返回结果
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("es query failed: %s err: ", resp.Status())
	}
	return nil
}

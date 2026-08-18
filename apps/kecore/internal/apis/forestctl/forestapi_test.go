package forestctl

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/insmtx/corekg/apps/kecore/models/forest"
	"github.com/ygpkg/yg-go/config"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"github.com/ygpkg/yg-go/dbtools/esquery"
	"github.com/ygpkg/yg-go/dbtools/estool"
	"github.com/ygpkg/yg-go/logs"
)

func InitESClient() (*elasticsearch.Client, error) {
	cfg := config.ESConfig{
		Addresses:     []string{"http://example.com:53082/"},
		SlowThreshold: time.Millisecond,
		Username:      "elastic",
		Password:      "CHANGE_ME_PASSWORD",
	}
	client, err := estool.InitES(cfg)
	if err != nil {
		logs.ErrorContextf(context.Background(), "init es client failed: %s", err)
		return nil, err
	}
	return client, nil
}

func DoDeleteRef(index_name string, fileIds []uint) error {
	client, err := InitESClient()
	if err != nil {
		return err
	}
	sourceStr := `
      if (ctx._source.references != null) {
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
		SetQuery(esquery.BuildMap("nested", esquery.BuildMap("path", "references", "query", esquery.BuildMap("terms", esquery.BuildMap("references.file_id", fileIds))))).
		Set("script", esquery.BuildMap("source", sourceStr, "lang", "painless", "params", esquery.BuildMap("file_ids_to_remove", fileIds)))

	querybyte, err := query.BuildBytes()
	if err != nil {
		return err
	}
	resp, err := client.UpdateByQuery([]string{index_name}, client.UpdateByQuery.WithBody(bytes.NewBuffer(querybyte)), client.UpdateByQuery.WithContext(context.Background()))
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

func TestDeleteFileRef(t *testing.T) {
	if err := DoDeleteRef("ke_0", []uint{2946}); err != nil {
		t.Fatal(err)
	}
}

func TestViewAbleForests(t *testing.T) {
	if err := dbtools.InitMultiDBConn(map[string]string{
		"knownow": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"core":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	}); err != nil {
		panic(err)
	}
	res, err := forest.ViewAbleForests(581, 72)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(res)
}

package es

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
	"github.com/ygpkg/yg-go/logs"
)

func InitESMapping(ctx context.Context) error {
	var (
		escli *elasticsearch.Client
		err   error
	)

	for {
		escli, err = essearch.InitESClient(ctx)
		if err != nil {
			logs.WarnContextf(ctx, "NewWrapper InitESClient error: %v, wait 10s and retry", err)
			time.Sleep(time.Second * 10)
			continue
		}

		resp, err := escli.Info()
		if err != nil {
			logs.ErrorContextf(ctx, "es query failed: %v, wait 10s and retry", err)
			time.Sleep(time.Second * 10)
			continue
		}
		logs.DebugContextf(ctx, "es query success: %v", resp.String())
		break
	}

	logs.InfoContextf(ctx, "start init es mapping")
	resp, err := escli.Indices.Create(
		"ke_0",
		escli.Indices.Create.WithBody(strings.NewReader(dsl)),
		escli.Indices.Create.WithContext(context.Background()),
	)
	if err != nil {
		logs.ErrorContextf(ctx, "es query failed: %v", err)
		return err
	}
	// 打印返回结果
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("es query failed: %s, error: %s", resp.Status(), string(body))
	}
	return nil
}

var dsl = `{
  "mappings": {
    "properties": {
      "version": {
        "type": "keyword"
      },
      "forest_id": {
        "type": "integer"
      },
      "company_id": {
        "type": "integer"
      },
      "uin": {
        "type": "integer"
      },
      "type": {
        "type": "keyword"
      },
      "description_hash": {
        "type": "keyword"
      },
      "content": {
        "type": "text",
        "analyzer": "ik_smart",
        "search_analyzer": "ik_smart",
        "fields": {
          "keyword": {
            "type": "keyword",
            "ignore_above": 256
          }
        }
      },
      "content_source": {
        "type": "text",
        "analyzer": "ik_smart",
        "search_analyzer": "ik_smart",
        "fields": {
          "keyword": {
            "type": "keyword",
            "ignore_above": 256
          }
        }
      },
      "content_target": {
        "type": "text",
        "analyzer": "ik_smart",
        "search_analyzer": "ik_smart",
        "fields": {
          "keyword": {
            "type": "keyword",
            "ignore_above": 256
          }
        }
      },
      "embedding": {
        "type": "dense_vector",
        "dims": 1024,
        "index": true
      },
      "tokens": {
        "type": "integer"
      },
      "mind_map": {
        "type": "text",
        "fields": {
          "keyword": {
            "type": "keyword",
            "ignore_above": 256
          }
        }
      },
      "abstract": {
        "type": "text",
        "fields": {
          "keyword": {
            "type": "keyword",
            "ignore_above": 256
          }
        }
      },
      "questions": {
        "type": "keyword"
      },
      "file_id": {
        "type": "integer"
      },
      "location": {
        "type": "integer",
        "index": false
      },
      "sequence": {
        "type": "integer"
      },
      "description": {
        "type": "text",
        "analyzer": "ik_smart",
        "search_analyzer": "ik_smart",
        "fields": {
          "keyword": {
            "type": "keyword",
            "ignore_above": 256
          }
        }
      },
      "image_url": {
        "type": "keyword"
      },
      "formula": {
        "type": "keyword"
      },
      "qa_question": {
        "type": "text",
        "analyzer": "ik_smart",
        "search_analyzer": "ik_smart",
        "fields": {
          "keyword": {
            "type": "keyword",
            "ignore_above": 256
          }
        }
      },
      "qa_answer": {
        "type": "text",
        "analyzer": "ik_smart",
        "search_analyzer": "ik_smart",
        "fields": {
          "keyword": {
            "type": "keyword",
            "ignore_above": 256
          }
        }
      },
      "qa_lable": {
        "type": "keyword"
      },
      "qa_answer_id": {
        "type": "keyword"
      },
      "qa_main_id": {
        "type": "keyword"
      },
      "title_level": {
        "type": "keyword"
      },
      "title_level_ids": {
        "type": "keyword"
      },
      "level": {
        "type": "integer"
      },
      "created_at": {
        "type": "date",
        "format": "strict_date_optional_time||epoch_millis"
      },
      "updated_at": {
        "type": "date",
        "format": "strict_date_optional_time||epoch_millis"
      },
      "references": {
        "type": "nested",
        "properties": {
          "file_id": {
            "type": "integer"
          },
          "description": {
            "type": "text",
            "analyzer": "ik_smart",
            "search_analyzer": "ik_smart",
            "fields": {
              "keyword": {
                "type": "keyword",
                "ignore_above": 256
              }
            }
          },
          "chunk_id": {
            "type": "keyword"
          },
          "relationship_id":{
            "type": "keyword"
          }
        }
      }
    },
    "dynamic": "strict"
  },
  "settings": {
    "number_of_shards": 2,
    "number_of_replicas": 1
  }
}`

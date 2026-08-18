package essearch

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/google/uuid"
	"github.com/insmtx/corekg/apps/ketask/models/ragtypes"
	"github.com/ygpkg/yg-go/logs"
)

// InsertFQA 插入问答对
func InsertFQA(ctx context.Context, indexName string, fqas []*ragtypes.FQA) error {
	client, err := InitESClient(ctx)
	if err != nil {
		return err
	}
	// 批量插入到ES
	var buf bytes.Buffer
	for _, doc := range fqas {
		meta := []byte(fmt.Sprintf(`{ "index" : { "_index" : "%s", "_id" : "%s" } }%s`, indexName, doc.ID, "\n"))
		doc.ID = ""
		data, err := json.Marshal(doc)
		if err != nil {
			logs.ErrorContextf(ctx, "json.Marshal error: %v", err)
			return err
		}
		data = append(data, byte('\n'))
		buf.Grow(len(meta) + len(data))
		buf.Write(meta)
		buf.Write(data)
	}
	logs.InfoContextf(ctx, "bulk insert: %s", buf.String())
	resp, err := client.Bulk(
		bytes.NewReader(buf.Bytes()),
		// client.Bulk.WithIndex(indexName),
		client.Bulk.WithContext(ctx),
		client.Bulk.WithRefresh("true"),
	)
	if err != nil {
		logs.ErrorContextf(ctx, "client.Bulk error: %v", err)
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("es Bulk failed: %s, error: %s", resp.Status(), string(body))
	}
	return nil
}

// GeneratFQA 生成并插入问答对
func GeneratFQA(ctx context.Context, indexName string, fqa *ragtypes.FQA, fqa_child []string) ([]*ragtypes.FQA, error) {
	fqa_list := []*ragtypes.FQA{}
	fqa.ID = uuid.NewString()
	fqa.QAAnswerID = uuid.NewString()
	eb, err := GetEmbedding(fqa.QAQuestion)
	if err != nil {
		logs.ErrorContextf(ctx, "GetEmbedding error: %v", err)
		return nil, err
	}
	fqa.Embedding = eb
	fqa_list = append(fqa_list, fqa)
	for _, v := range fqa_child {
		child := &ragtypes.FQA{
			Common: ragtypes.Common{
				ID:         uuid.NewString(), // 为 child 分配新 ID
				CreatedAt:  fqa.CreatedAt,
				UpdatedAt:  fqa.UpdatedAt,
				ForestID:   fqa.ForestID,
				CompanyID:  fqa.CompanyID,
				Uin:        fqa.Uin,
				Type:       fqa.Type,
				SourceFrom: fqa.SourceFrom,
				Enable:     fqa.Enable,
			},
			QAQuestion: v,
			QALable:    fqa.QALable,
			QAAnswer:   fqa.QAAnswer,
			QAMainID:   fqa.ID,
			QAAnswerID: fqa.QAAnswerID,
		}

		eb, err := GetEmbedding(child.QAQuestion)
		if err != nil {
			logs.ErrorContextf(ctx, "GetEmbedding for child error: %v", err)
			return nil, err
		}
		child.Embedding = eb

		fqa_list = append(fqa_list, child)
	}
	return fqa_list, nil
}

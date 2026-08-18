package essearch

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/insmtx/corekg/apps/ketask/models/ragtypes"
	"github.com/xuri/excelize/v2"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
)

func TestDeltetFile(t *testing.T) {
	// DeleteFileReferences("ke_test", []uint{123})
	dbtools.InitMultiDBConn(map[string]string{
		"chat":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"knownow": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"core":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	})

	wrapper, err := NewEsSearchWrapper(context.Background(), "ke_0", "我们公司最近进了不少新面孔啊，上手得怎么样啊", []uint{41}, nil)
	if err != nil {
		//logs.Errorf("GetEmbedding error: %v", err)
		return
	}

	a, err := wrapper.SearchQuestionChunk()
	// for _, v := range a.Hits.Hits {
	// 	fmt.Println(v.Source.References)
	// }
	fmt.Println(11111111, err)
	fmt.Println(len(a.Hits.Hits))
	// err := DeleteFileReferences("ke_0", []uint{0})
	// fmt.Println(11111111, err)
	b, err := wrapper.SearchChunkSequence(a)
	for _, v := range b.Hits.Hits {
		fmt.Println(v.Source.Sequence)
	}
	fmt.Println(11111111, err)
	fmt.Println(len(a.Hits.Hits), len(b.Hits.Hits))
}

func TestEmbeding(t *testing.T) {
	dbtools.InitMultiDBConn(map[string]string{
		"chat":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"knownow": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"core":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	})
	InitEbConfig(context.Background())
	eb, err := GetEmbedding("请回答测试问题")
	if err != nil {
		t.Error(err)
	}
	fmt.Println(eb)
	// jsondata, _ := json.Marshal(eb)
	// fmt.Println(string(jsondata))
}

func TestGetKeyWords(t *testing.T) {
	// words, err := GetKeyWords("我现在需要一个至少有3个RJ45口的面板AP，请给我一个推荐？")
	// if err != nil {
	// 	t.Error(err)
	// }
	// fmt.Println(words)
}

func TestAnalyze(t *testing.T) {
	res, err := Analyze(context.Background(), "我需要一个至少有3个RJ45口的面板AP，请给我一个推荐？")
	fmt.Println(res, err)

}

func TestSearchFQA(t *testing.T) {
	dbtools.InitMultiDBConn(map[string]string{
		"chat":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"knownow": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"core":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	})
	wrapper, err := NewEsSearchWrapper(context.Background(), "ke_0", "请回答测试问题", []uint{41, 160}, nil)
	if err != nil {
		//logs.Errorf("GetEmbedding error: %v", err)
		return
	}
	res, err := wrapper.FindFQAByQuestion()
	fmt.Println(res, err)
}

func TestInsertFQA(t *testing.T) {
	dbtools.InitMultiDBConn(map[string]string{
		"chat":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"knownow": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"core":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	})
	fqa, err := GeneratFQA(context.Background(), "ke_0", &ragtypes.FQA{
		Common: ragtypes.Common{
			ForestID:  27,
			CompanyID: 18,
			Uin:       434,
			Type:      ragtypes.ChunkTypeFQA,
		},
		QAQuestion: "请回答测试问题1",
		QAAnswer:   "我是测试回答2",
	}, []string{"请回答测试问题2", "请回答测试问题3"})
	if err != nil {
		fmt.Println(111111111, err)
		return
	}
	err = InsertFQA(context.Background(), "ke_0", fqa)
	fmt.Println(err)
}

func TestInsertExcel(t *testing.T) {
	dbtools.InitMultiDBConn(map[string]string{
		// "chat":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		// "knownow": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"core": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	})
	// 打开Excel文件
	f, err := excelize.OpenFile("小贝测试问答对.xlsx")
	if err != nil {
		//logs.Fatal("打开文件失败: ", err)
	}
	// 获取第一个工作表的名字
	sheetName := f.GetSheetName(0)
	// 获取该工作表的所有行
	rows, err := f.GetRows(sheetName)
	if err != nil {
		log.Fatal("获取行失败: ", err)
	}
	// 遍历每一行
	for i, row := range rows {
		if i == 0 {
			continue
		}
		fmt.Printf("第 %d 行数据:\n", i+1)
		//遍历每一列
		for j, col := range row {
			fmt.Printf("  列 [%d] 值: %s\n", j, col)
		}
		fmt.Println()
		ch := strings.Split(row[2], "\n")
		ch = removeEmptyOrBlankStrings(ch)
		fmt.Println(ch, len(ch))
		fqa, err := GeneratFQA(context.Background(), "ke_0", &ragtypes.FQA{
			Common: ragtypes.Common{
				ForestID:  27,
				CompanyID: 18,
				Uin:       434,
				Type:      ragtypes.ChunkTypeFQA,
			},
			QAQuestion: row[1],
			QAAnswer:   row[3],
			QALable:    []string{row[0]},
		}, ch)
		if err != nil {
			fmt.Println(111111111, err)
			return
		}
		err = InsertFQA(context.Background(), "ke_0", fqa)
		if err != nil {
			fmt.Println(2222222222, err)
			return
		}
	}
}

func removeEmptyOrBlankStrings(arr []string) []string {
	var result []string
	for _, s := range arr {
		trimmed := strings.TrimSpace(s)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func TestUpdateEb(t *testing.T) {
	//dbtools.InitMultiDBConn(map[string]string{
	//	"chat":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	//	"knownow": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	//	"core":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	//})
	//ctx := context.TODO()
	//client, err := InitESClient(ctx)
	//if err != nil {
	//	return
	//}
	//query := esquery.NewBuilder().
	//	SetQuery(esquery.BuildMap("bool", esquery.BuildMap("must_not", []esquery.Map{esquery.BuildMap("exists", esquery.BuildMap("field", "embedding"))}))).
	//	SetSize(100)
	//querybyte, err := query.BuildBytes()
	//if err != nil {
	//	logs.Errorf("query build error: %v", err)
	//	return
	//}
	//logs.Infof("querybyte: %v", string(querybyte))
	//resp, err := client.Search(
	//	client.Search.WithIndex("ke_0"),
	//	client.Search.WithBody(bytes.NewBuffer(querybyte)),
	//)
	//if err != nil {
	//	logs.Errorf("es query failed: %v", err)
	//	return
	//}
	//// 打印返回结果
	//defer resp.Body.Close()
	//if resp.StatusCode != http.StatusOK {
	//	body, _ := io.ReadAll(resp.Body)
	//	logs.Errorf("es query failed: %s, error: %s", resp.Status(), string(body))
	//	return
	//}
	//// 转换结果
	//// 读取完整响应体
	//body, err := io.ReadAll(resp.Body)
	//if err != nil {
	//	logs.Errorf("error reading body: %v", err)
	//	return
	//}
	//// 解析JSON响应
	//var searchRes SearchResult
	//if err := json.Unmarshal(body, &searchRes); err != nil {
	//	logs.Errorf("unmarshal ChatResponse error: %v", err)
	//	return
	//}
	//
	//for _, hit := range searchRes.Hits.Hits {
	//	fmt.Println(hit.Source)
	//	eb, err := GetEmbedding(hit.Source.Description)
	//	if err != nil {
	//		logs.Errorf("GetEmbedding error: %v", err)
	//		return
	//	}
	//	fmt.Println(eb)
	//	query := esquery.NewBuilder().Set("doc", esquery.BuildMap("embedding", eb))
	//	querybyte, err := query.BuildBytes()
	//	if err != nil {
	//		logs.Errorf("query build error: %v", err)
	//		return
	//	}
	//	logs.Infof("querybyte: %v", string(querybyte))
	//	resp, err := client.Update(
	//		"ke_0",
	//		hit.ID,
	//		bytes.NewBuffer(querybyte),
	//	)
	//	if err != nil {
	//		logs.Errorf("es query failed: %v", err)
	//		return
	//	}
	//	// 打印返回结果
	//	defer resp.Body.Close()
	//	if resp.StatusCode != http.StatusOK {
	//		body, _ := io.ReadAll(resp.Body)
	//		logs.Errorf("es query failed: %s, error: %s", resp.Status(), string(body))
	//		return
	//	}
	//}

}

func TestUpdateEbaa(t *testing.T) {
	dbtools.InitMultiDBConn(map[string]string{
		// "chat":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		// "knownow": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"core": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	})
	//client, err := InitESClient()
	//if err != nil {
	//	return
	//}
	//// 初始查询
	//var fullResult SearchResult
	//scrollID := ""
	//
	//var buf bytes.Buffer
	//query := map[string]interface{}{
	//	"size": 10000,
	//	"query": map[string]interface{}{
	//		"match_all": map[string]interface{}{},
	//	},
	//}
	//if err := json.NewEncoder(&buf).Encode(query); err != nil {
	//	log.Fatalf("Error encoding query: %s", err)
	//}
	//
	//res, err := client.Search(
	//	client.Search.WithContext(context.Background()),
	//	client.Search.WithIndex("ke_0"),
	//	client.Search.WithBody(&buf),
	//	client.Search.WithScroll(time.Minute*10),
	//)
	//if err != nil {
	//	log.Fatalf("Error getting response: %s", err)
	//}
	//defer res.Body.Close()
	//
	//// 保存 response 内容
	//buf.ReadFrom(res.Body)
	//res.Body.Close()
	//
	//var scrollHits SearchResult
	//var scrollIDHolder struct {
	//	ScrollID string `json:"_scroll_id"`
	//}
	//
	//json.Unmarshal(buf.Bytes(), &scrollIDHolder)
	//json.Unmarshal(buf.Bytes(), &scrollHits)
	//
	//scrollID = scrollIDHolder.ScrollID
	//fullResult.Hits.Total = scrollHits.Hits.Total
	//fullResult.Hits.MaxScore = scrollHits.Hits.MaxScore
	//fullResult.Hits.Hits = append(fullResult.Hits.Hits, scrollHits.Hits.Hits...)
	//
	//// 滚动
	//for {
	//	if len(scrollHits.Hits.Hits) == 0 {
	//		break
	//	}
	//
	//	res, _ := client.Scroll(client.Scroll.WithScrollID(scrollID), client.Scroll.WithScroll(time.Minute*10))
	//	buf.Reset()
	//	buf.ReadFrom(res.Body)
	//	res.Body.Close()
	//
	//	json.Unmarshal(buf.Bytes(), &scrollIDHolder)
	//	json.Unmarshal(buf.Bytes(), &scrollHits)
	//
	//	scrollID = scrollIDHolder.ScrollID
	//	fullResult.Hits.Hits = append(fullResult.Hits.Hits, scrollHits.Hits.Hits...)
	//	fmt.Printf("累计：%d 条\n", len(fullResult.Hits.Hits))
	//}
	//
	//const maxConcurrency = 500
	//limiter := make(chan struct{}, maxConcurrency)
	//var wg sync.WaitGroup
	//
	//for i, hit := range fullResult.Hits.Hits {
	//	// 控制并发数量
	//	limiter <- struct{}{}
	//	wg.Add(1)
	//	hit_item := hit
	//	index := i
	//	// 启动 goroutine 处理每个 hit
	//	go func() {
	//		defer func() {
	//			<-limiter // 释放一个槽位
	//			wg.Done()
	//		}()
	//
	//		// 打印 source
	//		// fmt.Println(hit_item.Source)
	//
	//		// 获取 embedding
	//		eb, err := GetEmbedding(hit_item.Source.Description)
	//		if err != nil {
	//			logs.Errorf("GetEmbedding error: %v", err)
	//			return
	//		}
	//		// fmt.Println(eb)
	//
	//		// 构造更新语句
	//		query := esquery.NewBuilder().Set("doc", esquery.BuildMap("embedding", eb))
	//		querybyte, err := query.BuildBytes()
	//		if err != nil {
	//			logs.Errorf("query build error: %v", err)
	//			return
	//		}
	//		// logs.Infof("querybyte: %v", string(querybyte))
	//
	//		// 发送更新请求
	//		resp, err := client.Update(
	//			"ke_0",
	//			hit_item.ID,
	//			bytes.NewBuffer(querybyte),
	//		)
	//		if err != nil {
	//			logs.Errorf("es query failed: %v", err)
	//			return
	//		}
	//
	//		// 立即关闭 body，避免泄露
	//		defer resp.Body.Close()
	//
	//		if resp.StatusCode != http.StatusOK {
	//			body, _ := io.ReadAll(resp.Body)
	//			logs.Errorf("es query failed: %s, error: %s", resp.Status(), string(body))
	//		}
	//		fmt.Println(index)
	//	}()
	//}
	//
	//// 等待所有任务完成
	//wg.Wait()
}

func TestIntentionRecognition(t *testing.T) {
	dbtools.InitMultiDBConn(map[string]string{
		"core": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	})
	//wrapper, err := NewEsSearchWrapper(context.Background(), "ke_0", "请回答测试问题", []uint{41, 160}, nil)
	//if err != nil {
	//	logs.Errorf("GetEmbedding error: %v", err)
	//	return
	//}
	//content, err := wrapper.IntentionRecognition()
	//if err != nil {
	//	logs.Errorf("IntentionRecognition error: %v", err)
	//	return
	//}
	//logs.Infof("IntentionRecognition---------: %s", content)
}

func TestDescriptionSearch(t *testing.T) {
	dbtools.InitMultiDBConn(map[string]string{
		"core": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	})
	//wrapper, err := NewEsSearchWrapper(context.Background(), "ke_0", "请回答测试问题", []uint{41, 160}, nil)
	//if err != nil {
	//	logs.Errorf("GetEmbedding error: %v", err)
	//	return
	//}
	//res, err := wrapper.DescriptionSearch()
	//fmt.Println(len(res.Hits.Hits), err)
}

func TestSearchForestData(t *testing.T) {
	dbtools.InitMultiDBConn(map[string]string{
		"core": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
	})
	res, err := SearchForestData(context.Background(), 137, "ke_0")
	fmt.Println(err)
	fmt.Println(len(res.Hits.Hits))
}

func InitEscli() *elasticsearch.Client {
	esCfg := elasticsearch.Config{
		Addresses: []string{"http://example.com:53084"},
		Username:  "corekg_api",
		Password:  "CHANGE_ME_PASSWORD",
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: true,
			},
		},
	}
	client, err := elasticsearch.NewClient(esCfg)
	if err != nil {
		panic(err)
	}
	return client
}

func TestEsSearchWrapper_GetFileChunk(t *testing.T) {
	ctx := context.Background()
	esIndex := "ke_0"
	var forestID uint = 39
	var fileID uint = 3481

	escli := InitEscli()

	chunks, err := NewPureWrapper(ctx, esIndex, []uint{forestID}, []uint{fileID}, escli).GetFileChunk(ragtypes.ChunkTypeVideo)
	if err != nil {
		panic(err)
	}
	fmt.Println(chunks)

}

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/insmtx/corekg/pkgs/agentclient"
	"github.com/ygpkg/yg-go/config"
	"github.com/ygpkg/yg-go/dbtools/estool"
)

func TestExtractMarkdownTitles(t *testing.T) {
	f, err := os.Open("/home/zoe/Downloads/nice.md")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	all, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	titles := ExtractMarkdownTitles(string(all))
	fmt.Println(titles)
}

//func TestApi(t *testing.T) {
//	resp, err := doAgentRequest("###你好你是谁", "https://api.example.com/v3/chat.Agent/chat/completions", "xxx", "NHfdA8m")
//	if err != nil {
//		t.Fatal(err)
//	}
//	fmt.Println(resp)
//}

func TestProcessEmbedded(t *testing.T) {
	f, err := os.Open("/home/zoe/Downloads/mindMap.json")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	all, err := io.ReadAll(f)
	if err != nil {
		t.Fatal(err)
	}
	embeddedUuid, err := ProcessEmbeddedUuid(string(all))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(embeddedUuid)
}

func TestLoadYaml(t *testing.T) {
	//loadConfig("/home/zoe/CodeSpace/roc/clients/task_worker/config/test.yaml")
	loadConfig(context.TODO(), "/home/zoe/CodeSpace/roc/clients/task_worker/config/test.yaml")
	fmt.Println(s3m)
	fmt.Println(esClient)
}

/*
export ES_HOST='http://example.com:53082'
export ES_USERNAME='elastic'
export ES_PASSWORD='CHANGE_ME_PASSWORD'
index name='ke test'
*/
func TestEsSearch(t *testing.T) {
	cfg := config.ESConfig{
		Addresses:  []string{"http://example.com:53082"},
		Username:   "elastic",
		Password:   "CHANGE_ME_PASSWORD",
		MaxRetries: 3, // 最大重试次数
	}
	cli, err := estool.InitES(cfg)
	if err != nil {
		panic(err)
	}
	chunkIDs, err := GetChunkIDsByFileID(cli, "ke_0", 1036)
	if err != nil {
		panic(err)
	}
	ids, err := json.Marshal(chunkIDs)
	if err != nil {
		panic(err)
	}
	fmt.Println(string(ids))
}

func TestMutiParse(t *testing.T) {
	resp, err := doMultiParseRequest(context.TODO(), []string{"https://example.com:53081/dev-knownow/algo-lke/434/41/928/content.md/images/befd24d90fb27c98a8d0519434fc9ba21e39264c9df0bb795ca41f245092ec78.jpg"}, "https://api.example.com/v3/llm.chat/chat/completions", "xxx", "qwen2.5-vl-72b-instruct", agentclient.ContentTypeImage)
	if err != nil {
		t.Fatal(err)
	}
	all, err := io.ReadAll(resp)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(string(all))

}

func TestExtractTitles(t *testing.T) {
	//file, err := os.ReadFile("/home/zoe/content (1).md")
	//if err != nil {
	//	panic(err)
	//}
	//titles := strings.Join(ExtractMarkdownTitles(string(file)), "\n")
	//fmt.Printf("titles:len:%v \n%v\n", len(titles), titles)
	////do agent request
	//fmt.Println("============do agent request==============")
	//resp, err := doAgentRequest(titles, "https://api.example.com/v3/chat.Agent/chat/completions", "xxx", "JAXwWUd")
	//if err != nil {
	//	t.Fatal(err)
	//}
	//logs.Debugf("analysis agent chat response: %v", resp)

	//extract json code block
	file, err := os.ReadFile("/home/zoe/CodeSpace/roc/mindmap.json")
	if err != nil {
		t.Fatal(err)
	}
	//jsonCode := ExtractCode("json", resp)

	//process resp with embedded uuid
	embeddedUuid, err := ProcessEmbeddedUuid(string(file))
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(embeddedUuid)
}

package test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	ygagent "github.com/insmtx/corekg/pkgs/einotools/agent"
	"github.com/insmtx/corekg/pkgs/einotools/models"
	"github.com/insmtx/corekg/pkgs/einotools/printer"
	agentservice "github.com/insmtx/corekg/pkgs/einotools/service"
	"github.com/insmtx/corekg/pkgs/einotools/tools"
	"github.com/insmtx/corekg/pkgs/einotools/utils"
	dbtools "github.com/ygpkg/yg-go/dbtools/v2"
	_ "github.com/ygpkg/yg-go/dbtools/v2/mysqldrv"
	"github.com/ygpkg/yg-go/dbtools/redispool"
	"github.com/ygpkg/yg-go/logs"
)

func TestReactAgent(t *testing.T) {
	ctx := context.Background()
	initTestDB()

	agentRequest := &models.AgentRequest{}
	agentRequest.SessionID = 1
	agentRequest.RequestID = "request-id"
	agentRequest.IsStream = false

	t.Run("TestSimpleCase", func(t *testing.T) {
		agentRequest.Query = "代码计算，求从 1 到 50000 的所有自然数中，偶数之和减去奇数之和的值，并推导出对任意正整数 𝑁 时该差值的通项公式。"
		_, err := HandleRequest(t, ctx, agentRequest)
		if err != nil {
			t.Errorf("HandleRequest failed: %v", err)
			return
		}
	})

	t.Run("TestSimpleCase2", func(t *testing.T) {
		agentRequest.Query = "如何让自己稳步发展，避免迷茫和焦虑？并保证自己的发展方向与动力"
		_, err := HandleRequest(t, ctx, agentRequest)
		if err != nil {
			t.Errorf("HandleRequest failed: %v", err)
			return
		}
	})

	t.Run("TestSimpleCase_Stream", func(t *testing.T) {
		agentRequest.Query = "求从 1 到 3000 的所有自然数中，偶数之和减去奇数之和的值，并推导出对任意正整数 𝑁 时该差值的通项公式。"
		agentRequest.IsStream = true
		_, err := HandleRequest(t, ctx, agentRequest)
		if err != nil {
			t.Errorf("HandleRequest failed: %v", err)
			return
		}
	})

	t.Run("TestTableAnalysis", func(t *testing.T) {
		file := models.File{
			FileName: "3466-split-sub.xlsx",
			// FileSize:    1024,
			FileOssUrl:  "./data/3466-split-sub.xlsx",
			Description: "3466-split-sub.xlsx",
		}

		ts := httptest.NewServer(http.StripPrefix(
			"/data/",
			http.FileServer(http.Dir("./data")),
		))
		defer ts.Close()

		fmt.Println("File URL:", ts.URL)
		file.FileOssUrl = ts.URL + "/data/3466-split-sub.xlsx"

		agentRequest.Query = "五大产业生态圈营业总收入是多少？每一项的占比是多少, 生成图表"
		agentRequest.InputFiles = []models.File{
			file,
		}
		_, err := HandleRequest(t, ctx, agentRequest)
		if err != nil {
			t.Errorf("HandleRequest failed: %v", err)
			return
		}
	})

}

func HandleRequest(t *testing.T, ctx context.Context, agentRequest *models.AgentRequest) (string, error) {
	aiKey := ""
	if len(aiKey) == 0 {
		t.Skip("DS_KEY not set, skipping TestReactAgent")
	}

	wg := &sync.WaitGroup{}
	wg.Add(1)

	msgPrinter := printer.NewLogPrinter(agentRequest)
	// msgPrinter := printer.NewSSEPrinter(agentRequest, nil, nil)

	result, err := func() (string, error) {
		defer wg.Done()
		chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
			APIKey:  aiKey,
			Model:   "deepseek-v3",
			Timeout: 300 * time.Second,
			BaseURL: "https://api.example.com/v3/llm.chat",
		})
		if err != nil {
			t.Errorf("创建OpenAiChatModel失败: %v", err)
		}

		agentContext := &ygagent.AgentContext{}
		agentContext.SessionID = agentRequest.SessionID
		agentContext.RequestID = agentRequest.RequestID
		agentContext.Query = agentRequest.Query
		agentContext.IsStream = agentRequest.IsStream
		agentContext.AgentRequest = agentRequest
		agentContext.Printer = msgPrinter
		agentContext.DateInfo = utils.CurrentDateInfoRFC1123()
		agentContext.ChatModel = chatModel
		agentContext.Tools = []tools.ToolOption{
			tools.ToolOptionCode,
			tools.ToolOptionFile,
			tools.ToolOptionChart,
		}

		reactAgentService := &agentservice.ReactAgentService{}
		result, err := reactAgentService.Handler(ctx, agentContext, agentRequest)
		if err != nil {
			return "", err
		}

		logs.InfoContextf(ctx, "===== reactAgentService.Stats =====")
		logs.InfoContextf(ctx, "reactAgentService.Stats: %v", reactAgentService.Stats)

		fmt.Println()

		logs.InfoContextf(ctx, "===== reactAgent.Memory.Messages =====")
		for i, msg := range reactAgentService.Memory.Messages {
			jsonData, err := json.MarshalIndent(msg, "", "  ")
			if err != nil {
				logs.ErrorContextf(ctx, "failed to marshal message %d: %v", i, err)
				continue
			}
			logs.InfoContextf(ctx, "Message[%d]:\n%s", i, string(jsonData))
		}

		fmt.Println()

		return result, nil
	}()
	wg.Wait()

	logs.InfoContext(ctx, "===== result =====")
	logs.InfoContext(ctx, result)
	fmt.Println()

	return result, err
}

func initTestDB() {
	if err := dbtools.InitMultiDBConn(map[string]string{
		"account": "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=true&loc=Local",
		"core":    "mysql://yygu_test_rw:CHANGE_ME_PASSWORD@CHANGE_ME_HOST:63807/yygu_test?charset=utf8mb4&parseTime=True&loc=Local",
	}); err != nil {
		panic(err)
	}
	_ = redispool.InitRedis("core", "loc_redis")
}

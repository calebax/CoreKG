package modes

import (
	"context"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/schema"
	"github.com/gin-gonic/gin"
	"github.com/go-pay/errgroup"
	"github.com/insmtx/corekg/apps/kechat/chat/core"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kecore/models/nebulagraph"
	graphagent "github.com/insmtx/corekg/apps/kecore/services/graphragsearch/agent"
	"github.com/insmtx/corekg/apps/kecore/services/graphragsearch/search"
	"github.com/insmtx/corekg/pkgs/einotools/models"
	"github.com/insmtx/corekg/pkgs/einotools/printer"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/apis/sseclient"
	"github.com/ygpkg/yg-go/dbtools/redispool"
	"github.com/ygpkg/yg-go/logs"
)

type GraphSearchChatMode struct {
	ctx *gin.Context
}

func NewGraphSearchChatMode(ctx *gin.Context) *GraphSearchChatMode {
	return &GraphSearchChatMode{ctx}
}

func (c *GraphSearchChatMode) Run(ctx context.Context, ctxData *core.ChatContext) (*core.ChatResult, error) {
	// 查询历史记录
	history, err := getForestChatHistory(c.ctx, ctxData.Session)
	if err != nil {
		logs.ErrorContextf(c.ctx, "[GraphSearchMode] ForestChat GetForestChatHistory error: %v", err)
		return nil, err
	}
	agentReq := &models.AgentRequest{
		SessionID: ctxData.Session.ID,
		RequestID: ctxData.Question.Source.ReqID,
		Query:     ctxData.Question.Source.Question,
		IsStream:  true,
		Messages:  history,
	}
	sseClient := sseclient.New(sseclient.WithRedisClient(redispool.Redis()),
		sseclient.WithExpiration(time.Minute*5))
	defer sseClient.Close(c.ctx, runtime.RequestID(c.ctx))
	msgPrinter := printer.NewSSEPrinter(agentReq, sseClient, c.ctx.Writer)

	// chatModel
	temperature := float32(0.4)
	chatModel, err := newToolCallingChatModel(c.ctx, ctxData.Model, chatModelOptions{
		Timeout:     300 * time.Second,
		Temperature: &temperature,
	})
	if err != nil {
		logs.ErrorContextf(ctx, "[GraphSearchMode] failed to create OpenAiChatModel: %v", err)
		return nil, err
	}

	fileD, err := graphagent.ExecuteCatalogueAgent(ctx, chatModel, agentReq.Query)
	if err != nil {
		logs.ErrorContextf(ctx, "[GraphSearchMode] ExecuteCatalogueAgent error: %v", err)
		return nil, err
	}

	anInput := []*graphagent.FileData{}
	{
		agg, fData, err := graphagent.NewFilterAgent(ctx, chatModel, fileD)
		if err != nil {
			logs.ErrorContextf(ctx, "ExecuteCatalogueAgent NewAnalystAgent error: %v", err)
			return nil, err
		}
		// 创建 Runner
		runner := adk.NewRunner(ctx, adk.RunnerConfig{
			Agent: agg,
		})
		iter := runner.Query(ctx, agentReq.Query)
		var idStr []string
		for {
			event, ok := iter.Next()
			if !ok {
				break
			}
			if event.Err != nil {
				logs.ErrorContextf(ctx, "ExecuteCatalogueAgent iter.Next error: %v", event.Err)
				continue
			}
			if event.Output != nil && event.Output.MessageOutput != nil {
				if m := event.Output.MessageOutput.Message; m != nil {
					if len(m.Content) > 0 {
						logs.InfoContextf(ctx, "ExecuteCatalogueAgent answer: %s", m.Content)
						idStr = append(idStr, strings.Split(m.Content, ",")...)
					}
				}
			}
		}
		// 转uint
		idUint := make([]uint, 0, len(idStr))
		for _, s := range idStr {
			id, _ := strconv.ParseUint(s, 10, 64)
			if id > 0 {
				idUint = append(idUint, uint(id))
			}
		}
		fDataMap := make(map[uint]*graphagent.FileData)
		// 在fData中找到idUint对应的数据
		for _, v := range fData {
			fDataMap[v.FileID] = v
		}
		anInput = []*graphagent.FileData{}
		for _, v := range idUint {
			date, ok := fDataMap[v]
			if ok {
				anInput = append(anInput, date)
			}
		}
	}
	var g errgroup.Group
	g.Go(func(ctx context.Context) error {
		allgraph, refGraph, err := getGraph(ctx, fileD, anInput)
		ctxData.Question.Source.GraphReference = allgraph
		ctxData.Question.Source.GraphChatReference = refGraph
		if err != nil {
			logs.ErrorContextf(ctx, "[GraphSearchMode] getGraph error: %v", err)
			return err
		}
		return nil
	})

	anAgent, err := graphagent.NewAnalystAgent(ctx, chatModel, anInput)
	if err != nil {
		logs.ErrorContextf(ctx, "[GraphSearchMode] NewAnalystAgent error: %v", err)
		return nil, err
	}
	// 创建 Runner
	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           anAgent,
		EnableStreaming: true,
	})
	iter := runner.Query(ctx, agentReq.Query)
	messages := []*schema.Message{}
	for {
		event, ok := iter.Next()
		if !ok {
			break
		}
		if event.Output != nil && event.Output.MessageOutput != nil {
			if s := event.Output.MessageOutput.MessageStream; s != nil {
				for {
					chunk, err := s.Recv()
					if err != nil {
						if err == io.EOF {
							break
						}
						logs.ErrorContextf(ctx, "get chunk error: %v")
						return nil, err
					}
					if chunk.Content != "" {
						msgPrinter.Send(ctx, "", models.MsgTypeResult, chunk.Content, false)
					}
					messages = append(messages, chunk)
				}
			}
		}

	}
	msg, err := schema.ConcatMessages(messages)
	if err != nil {
		logs.ErrorContextf(ctx, "[GraphSearchMode] ConcatMessages error: %v", err)
		return nil, err
	}

	chatResult := &core.ChatResult{
		Answer: msg.Content,
		Status: chattype.QuestionStatusAnswered,
	}
	// 设置Token使用情况
	chatResult.Usage.PromptTokens = msg.ResponseMeta.Usage.PromptTokens
	chatResult.Usage.CompletionTokens = msg.ResponseMeta.Usage.CompletionTokens
	chatResult.Usage.CacheHitTokens = msg.ResponseMeta.Usage.PromptTokenDetails.CachedTokens
	chatResult.Usage.CacheMissTokens = msg.ResponseMeta.Usage.PromptTokens - msg.ResponseMeta.Usage.PromptTokenDetails.CachedTokens
	chatResult.Usage.TotalTokens = msg.ResponseMeta.Usage.TotalTokens

	ctxData.Question.Source.Answer = msg.Content

	if err := g.Wait(); err != nil {
		logs.ErrorContextf(ctx, "[GraphSearchMode] getGraph error: %v", err)
		return chatResult, nil
	}

	return chatResult, nil
}

// getGraph 获取图谱数据
func getGraph(ctx context.Context, fileD []search.DirectoryInfo, anInput []*graphagent.FileData) (*nebulagraph.Graph, *nebulagraph.Graph, error) {
	nodeID := []string{}
	refNodeID := []string{}
	for _, v := range fileD {
		nodeID = append(nodeID, v.ID)
		for _, vv := range anInput {
			if vv.FileID == v.FileID {
				refNodeID = append(refNodeID, v.ID)
			}
		}
	}
	cli, err := nebulagraph.NewNebulaCLI(ctx, "a_car_test")
	if err != nil {
		logs.ErrorContextf(ctx, "NewCatalpgueParallelAgent NewNebulaCLI error: %v", err)
		return nil, nil, err
	}
	defer cli.Release()

	// allgraph, err := cli.GetNodesGraphWithStep(nodeID, 3)
	// if err != nil {
	// 	logs.ErrorContextf(ctx, "NewCatalpgueParallelAgent GetNodesGraph error: %v", err)
	// 	return nil, nil, err
	// }

	refgraph, err := cli.GetNodesGraphWithStep(refNodeID, 3)
	if err != nil {
		logs.ErrorContextf(ctx, "NewCatalpgueParallelAgent GetNodesGraph error: %v", err)
		return nil, refgraph, err
	}

	return refgraph, refgraph, nil
}

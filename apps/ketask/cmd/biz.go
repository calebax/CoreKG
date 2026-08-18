package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/ketask/models/ragtypes"
	"github.com/insmtx/corekg/apps/kesearch/models/essearch"
	"github.com/insmtx/corekg/pkgs/agentclient"
	"github.com/ygpkg/yg-go/encryptor"
	"github.com/ygpkg/yg-go/logs"
	"golang.org/x/sync/errgroup"
)

type Worker interface {
	Init() error
	PreRun() error
	Run() error
	PostRun() error
	CallBack() error
	Close() error
}

type SubCall func(context.Context) error

type Node struct {
	ID       string  `json:"id"`
	UUID     string  `json:"uuid,omitempty"`
	Children []*Node `json:"children,omitempty"`
}

func ProcessEmbeddedUuid(js string) (string, error) {
	var node *Node

	if err := json.Unmarshal([]byte(js), &node); err != nil {
		return "", err
	}

	generateUUIDs(node)

	res, err := json.Marshal(node)
	if err != nil {
		return "", err
	}

	return string(res), nil
}

func generateUUIDs(node *Node) {
	if node == nil {
		return
	}
	node.UUID = encryptor.UUID()

	for _, child := range node.Children {
		generateUUIDs(child)
	}
}

func doAgentRequest(ctx context.Context, inputs map[string]string, agentUrl, agentApikey, model string, backups ...string) (string, error) {
	client := agentclient.NewChatClient(nil, agentUrl, agentApikey)

	var inputList []chattype.Input
	for name, value := range inputs {
		if len(value) > int(appCfg.Agent.MaxTokenSize) {
			if len(backups) != 2 {
				return "", fmt.Errorf("invalid backup model for agent[%v], backups[%+v] content len[%v]", model, backups, len(value))
			}
			return doSplitMerge(ctx, value, backups[0], backups[1])
		}
		inputList = append(inputList, chattype.Input{
			Name:  name,
			Value: value,
		})
	}

	req := &agentclient.ChatRequestBody{
		Model: model,
		ChatOptions: agentclient.ChatOptions{
			Input: inputList,
		},
		Stream: false,
	}

	var (
		resp *agentclient.ChatResponseBody
		err  error
	)
	logs.InfoContextf(ctx, "send request to model:[%v] with content:[%+v]", model, inputs)
	for i := 0; i < 20; i++ {
		resp, err = client.SendChat(ctx, req)
		if err != nil {
			if errors.Is(err, context.Canceled) {
				break
			}
			logs.ErrorContextf(ctx, "doAgentRequest SendChat err: %v retry with %v time", err, i)
			time.Sleep(500 * time.Millisecond)
			continue
		}
		break
	}
	if resp == nil || err != nil {
		logs.ErrorContextf(ctx, "doAgentRequest nil resp err: %v \n requestBody:%v", err, req)
		return "", err
	}

	if len(resp.Choices) == 0 {
		logs.WarnContextf(ctx, "no choices: \n%+v", *resp)
		return "", fmt.Errorf("no choices found")
	}
	if len(resp.Choices[0].Message.Content) <= 0 {
		logs.WarnContextf(ctx, "no choices: \n%+v", *resp)
		return "", fmt.Errorf("no content found")
	}

	logs.DebugContextf(ctx, "agent return:\n%v",
		resp.Choices[0].Message.Content[:min(len(resp.Choices[0].Message.Content), 100)])
	return resp.Choices[0].Message.Content, nil
}

func ExtractMarkdownTitles(mdContent string) []string {
	re := regexp.MustCompile(`(?m)(^#{1,6}\s+(.+)$)|(^(.+)\n[-=]{3,}$)`)

	matches := re.FindAllStringSubmatch(mdContent, -1)

	var titles []string
	for _, match := range matches {
		switch {
		case len(match[2]) > 0:
			titles = append(titles, strings.TrimSpace(match[2]))
		case len(match[4]) > 0:
			titles = append(titles, strings.TrimSpace(match[4]))
		}
	}

	return titles
}

func ExtractCode(codetype, text string) string {
	re := regexp.MustCompile("(?s)```" + codetype + "\\s*(.*?)\\s*```")
	match := re.FindStringSubmatch(text)
	if len(match) > 1 {
		return match[1]
	} else {
		return text
	}
}

func doMultiParseRequest(ctx context.Context, muliUrl []string, agentUrl, agentApikey, model string, ctype agentclient.ContentType) (io.Reader, error) {
	var requestBody map[string]interface{}

	switch ctype {
	case agentclient.ContentTypeImage:
		requestBody = map[string]interface{}{
			"model": model,
			"messages": []interface{}{
				map[string]interface{}{
					"role": "user",
					"content": []interface{}{
						map[string]string{"type": "text", "text": "这是什么"},
						map[string]interface{}{
							"type": "image_url",
							"image_url": map[string]string{
								"url": muliUrl[0],
							},
						},
					},
				},
			},
			"stream": false,
		}
	case agentclient.ContentTypeVideo:
		requestBody = map[string]interface{}{
			"model": model,
			"messages": []interface{}{
				map[string]interface{}{
					"role": "user",
					"content": []interface{}{
						map[string]string{"type": "text", "text": "这是什么"},
						map[string]interface{}{
							"type": "image_url",
							"image_url": map[string][]string{
								"url": muliUrl,
							},
						},
					},
				},
			},
			"stream": false,
		}
	default:
		return nil, fmt.Errorf("unsupported content type")
	}

	jsonData, _ := json.Marshal(requestBody)
	req, err := http.NewRequest("POST", agentUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+agentApikey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Request err:", err)
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		logs.ErrorContextf(ctx, "Received non-OK HTTP status: %s, body: %s", resp.Status, string(body))
		return nil, fmt.Errorf("received non-OK HTTP status: %s", resp.Status)
	}

	var res agentclient.ChatResponseBody
	if err = json.NewDecoder(resp.Body).Decode(&res); err != nil {
		logs.ErrorContextf(ctx, "Unmarshal response err:%v", err)
		return nil, err
	}

	if len(res.Choices) == 0 {
		return nil, fmt.Errorf("no choices found")
	}
	if len(res.Choices[0].Message.Content) <= 0 {
		return nil, fmt.Errorf("no content found")
	}

	return bytes.NewReader([]byte(res.Choices[0].Message.Content)), nil
}

const (
	MDAbstract      = "abstractMD"
	MDAbsChunk      = "absChunkMD"
	MDMergeAbstract = "mergeAbstractMD"
	MDMindmap       = "mindmapMD"
	MDMindChunk     = "mindChunkMD"
	MDMergeMindmap  = "mergeMindmapMD"
	MDShortDesc     = "shortDescMD"
	MDQuestions     = "questionsMD"
)

func doSplitMerge(ctx context.Context, content, chunkModel, mergeModel string) (string, error) {
	chunks := splitContent(content)
	logs.InfoContextf(ctx, "Split into %d chunks", len(chunks))

	summaries, err := mapPhase(ctx, chunks, chunkModel)
	if err != nil {
		return "", err
	}

	return reducePhase(ctx, summaries, chunkModel, mergeModel)

}

func splitContent(content string) []string {
	maxChunkSize := int(appCfg.Agent.ChunkSize)

	if len(content) <= maxChunkSize {
		return []string{content}
	}

	var chunks []string
	paragraphs := strings.Split(content, "\n")

	currentChunk := ""
	for _, para := range paragraphs {
		if len(para) > maxChunkSize {
			if currentChunk != "" {
				chunks = append(chunks, currentChunk)
				currentChunk = ""
			}
			chunks = append(chunks, splitLongString(para, maxChunkSize)...)
			continue
		}

		chunkLen := len(currentChunk)
		if currentChunk != "" {
			chunkLen += len("\n")
		}

		if chunkLen+len(para) <= maxChunkSize {
			if currentChunk != "" {
				currentChunk += "\n"
			}
			currentChunk += para
		} else {
			if currentChunk != "" {
				chunks = append(chunks, currentChunk)
			}
			currentChunk = para
		}
	}

	if currentChunk != "" {
		chunks = append(chunks, currentChunk)
	}

	return chunks
}

func splitLongString(s string, chunkSize int) []string {
	var result []string
	runes := []rune(s)

	for len(runes) > 0 {
		if len(runes) <= chunkSize {
			result = append(result, string(runes))
			break
		}

		splitAt := chunkSize
		if splitAt < len(runes) {
			for i := chunkSize; i > max(0, chunkSize-30); i-- {
				if i >= len(runes) {
					continue
				}
				if isBreakRune(runes[i]) {
					splitAt = i + 1
					break
				}
			}
		}

		result = append(result, string(runes[:splitAt]))
		runes = runes[splitAt:]
	}
	return result
}

func isBreakRune(r rune) bool {
	switch r {
	case ' ', '\t', '\n', '\r':
		return true
	case ',', '.', ';', '!', '?', ':', '-', '—':
		return true
	case '，', '。', '；', '！', '？', '：', '、', '「', '」', '『', '』', '【', '】', '《', '》':
		return true
	default:
		return false
	}
}

func mapPhase(ctx context.Context, chunks []string, chunkModel string) ([]string, error) {
	results := make([]string, len(chunks))
	group, gCtx := errgroup.WithContext(ctx)
	semaphore := make(chan struct{}, appCfg.Agent.MaxWorkers)

	for i, chunk := range chunks {
		idx := i
		content := chunk

		group.Go(func() error {
			select {
			case <-gCtx.Done():
				return gCtx.Err()
			case semaphore <- struct{}{}:
				defer func() { <-semaphore }()

				inputs := map[string]string{
					"input1": content,
					"input2": fmt.Sprintf("第%d块，共%d块", idx+1, len(chunks)),
				}

				summary, err := doAgentRequest(gCtx, inputs, appCfg.Agent.APIUrl, appCfg.Agent.APIKey, chunkModel)
				if err != nil {
					logs.ErrorContextf(gCtx, "Chunk %d failed: %v", idx, err)
					return fmt.Errorf("chunk %d failed: %w", idx, err)
				}
				results[idx] = summary
				logs.InfoContextf(gCtx, "Chunk %d completed", idx+1)
				return nil
			}
		})
	}

	if err := group.Wait(); err != nil {
		logs.ErrorContextf(ctx, "mapPhase failed: %v", err)
		return nil, err
	}

	return results, nil
}

func reducePhase(ctx context.Context, summaries []string, chunkModel, mergeModel string) (resp string, err error) {
	combined := strings.Join(summaries, "\n---\n")
	inputsAbs := map[string]string{
		"input1": combined,
		"input2": fmt.Sprintf("共%d个分块的总结", len(summaries)),
	}

	resp, err = doAgentRequest(ctx, inputsAbs, appCfg.Agent.APIUrl, appCfg.Agent.APIKey, mergeModel, chunkModel, mergeModel)
	if err != nil {
		logs.ErrorContextf(ctx, "reducePhase merge err: %v", err)
		return "", err
	}
	return
}

func GetEmbedding(question string) (ragtypes.Embedding, error) {
	reqBody := map[string]interface{}{
		"model": appCfg.Agent.Embedding.ModelName,
		"input": question,
	}
	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("POST", appCfg.Agent.Embedding.Url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+appCfg.Agent.Embedding.Key)
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, err
	}
	var respBody essearch.EmbeddingData
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return nil, err
	}
	if respBody.Data == nil {
		return nil, fmt.Errorf("no embedding data")
	}
	return respBody.Data[0].Embedding, nil
}

func WrapWithContext[T any](ctx context.Context, f func() (T, error)) (T, error) {
	resultChan := make(chan struct {
		result T
		err    error
	}, 1)

	go func() {
		res, err := f()
		resultChan <- struct {
			result T
			err    error
		}{res, err}
	}()

	select {
	case <-ctx.Done():
		var zero T
		return zero, ctx.Err()
	case res := <-resultChan:
		return res.result, res.err
	}
}

package llmchat

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/pkgs/global"
	"github.com/insmtx/corekg/version"
	"github.com/ygpkg/yg-go/apis/runtime"
	"github.com/ygpkg/yg-go/apis/sseclient"
	"github.com/ygpkg/yg-go/dbtools/redispool"
	"github.com/ygpkg/yg-go/logs"
)

// ChatRequest 发请求到模型 最底层发送临门一脚
func (w *LLMChatWrapper) ChatRequest() (*http.Response, error) {
	jsonPayload, err := w.req.ToString()
	if err != nil {
		logs.ErrorContext(w.ctx, "ChatRequest::ToString Failed to marshal: %v", err)
		return nil, err
	}
	jsonPayload, err = withDisableThinkingIfNeeded(jsonPayload, w.model.ModelName)
	if err != nil {
		logs.ErrorContextf(w.ctx, "ChatRequest failed to append chat_template_kwargs: %v", err)
		return nil, err
	}
	var reqBody map[string]any
	if err := json.Unmarshal([]byte(jsonPayload), &reqBody); err == nil {
		if kwargs, ok := reqBody["chat_template_kwargs"]; ok {
			logs.InfoContextf(w.ctx, "chat request model=%s, chat_template_kwargs=%v", reqBody["model"], kwargs)
		} else {
			logs.InfoContextf(w.ctx, "chat request model=%s", reqBody["model"])
		}
	}
	logs.InfoContextf(w.ctx, "chat request with json:\n%s\n", jsonPayload)
	// 创建 HTTP
	req, err := http.NewRequestWithContext(w.ctx, "POST", w.model.ModelUrl, strings.NewReader(jsonPayload))
	if err != nil {
		logs.ErrorContextf(w.ctx, "ChatRequest Failed to create HTTP request: %s", err.Error())
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+w.model.APIKey)

	// 创建 HTTP 客户端
	client := &http.Client{}
	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		logs.ErrorContextf(w.ctx, "ChatRequest Failed to make HTTP request: %v", err)
		return nil, fmt.Errorf("ChatRequest Failed to make HTTP request: %w", err)
	}
	// 检查响应状态码
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		body, _ := io.ReadAll(resp.Body)
		logs.ErrorContextf(w.ctx, "ChatRequest Error status: %s,err: %s", resp.Status, string(body))
		return nil, fmt.Errorf("ChatRequest Error status: %s,err: %s", resp.Status, string(body))
	}

	// 返回响应
	return resp, nil
}

// withDisableThinkingIfNeeded 为私有化场景下的指定 Qwen3-5-35B-A3B 模型关闭思考模式。
func withDisableThinkingIfNeeded(jsonPayload string, modelName string) (string, error) {
	logs.Infof("[shouldDisableThinking] deploy mode: %s", version.DeployMode())
	logs.Infof("[shouldDisableThinking] model name: %s", modelName)
	if !shouldDisableThinking(modelName) {
		return jsonPayload, nil
	}

	reqBody := map[string]any{}
	if err := json.Unmarshal([]byte(jsonPayload), &reqBody); err != nil {
		return "", err
	}
	reqBody["chat_template_kwargs"] = map[string]any{
		"enable_thinking": false,
	}

	logs.Infof("enable_thinking value: %s", logs.JSON(reqBody["chat_template_kwargs"]))

	updatedPayload, err := json.Marshal(reqBody)
	if err != nil {
		return "", err
	}
	return string(updatedPayload), nil
}

func shouldDisableThinking(modelName string) bool {
	if !strings.EqualFold(version.DeployMode(), global.DeployModeOnPremise) {
		return false
	}
	lower := strings.ToLower(modelName)
	for _, keyword := range global.DisableThinkingModelKeywords {
		if strings.Contains(lower, strings.ToLower(keyword)) {
			return true
		}
	}
	return false
}

// ChatResult 内部调用不返回前端结果 非流式
func (w *LLMChatWrapper) ChatResult(resp *http.Response) (*QaRes, error) {
	// 读取响应体
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logs.ErrorContextf(w.ctx, "ChatResponse read response body failed: %v", err)
		return nil, fmt.Errorf("read response failed: %w", err)
	}
	// 解析响应
	var response chattype.ChatResponseBody
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		logs.ErrorContextf(w.ctx, "ChatResponse error unmarshalling json: %v", err)
		return nil, fmt.Errorf("response Parsing Error: %w", err)
	}
	// 验证响应有效性
	if len(response.Choices) == 0 && response.Error.Code == 0 {
		logs.ErrorContextf(w.ctx, "ChatResponse response is empty")
		return nil, fmt.Errorf("invalid API Response")
	}
	resault := &QaRes{}
	resault.Content = response.Choices[0].Message.Content
	resault.Reasoning = response.Choices[0].Message.ReasoningContent
	resault.Usage = Usage(response.Usage)
	return resault, nil
}

// ChatResponse 非流式返回问题
func (w *LLMChatWrapper) ChatResponse(resp *http.Response) (*QaRes, error) {
	// 读取响应体
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		logs.ErrorContextf(w.ctx, "ChatResponse read response body failed: %v", err)
		return nil, fmt.Errorf("read response failed: %w", err)
	}
	// 解析响应
	var response chattype.ChatResponseBody
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		logs.ErrorContextf(w.ctx, "ChatResponse error unmarshalling json: %v", err)
		return nil, fmt.Errorf("response Parsing Error: %w", err)
	}
	// 验证响应有效性
	if len(response.Choices) == 0 && response.Error.Code == 0 {
		logs.ErrorContextf(w.ctx, "ChatResponse response is empty")
		return nil, fmt.Errorf("invalid API Response")
	}
	resault := &QaRes{}
	resault.Content = response.Choices[0].Message.Content
	resault.Reasoning = response.Choices[0].Message.ReasoningContent
	resault.Usage = Usage(response.Usage)
	// 返回前端
	w.ctx.JSON(http.StatusOK, response)
	return resault, nil
}

// ChatStreamResponse 流式返回问题
func (w *LLMChatWrapper) ChatStreamResponse(resp *http.Response) (*QaRes, error) {
	defer resp.Body.Close()
	sseClient := sseclient.New(sseclient.WithRedisClient(redispool.Redis()),
		sseclient.WithExpiration(time.Minute*5))
	defer sseClient.Close(w.ctx, runtime.RequestID(w.ctx))
	// 设置响应头
	sseClient.SetHeaders(w.ctx.Writer)

	var (
		reader           = bufio.NewReader(resp.Body)
		storedLinedata   string
		contentBuilder   strings.Builder
		reasoningBuilder strings.Builder
	)
	resault := &QaRes{}
	reasoningStartTime := time.Now()
	for {
		// 读取一行数据
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			// 读取完毕,结束请求
			break
		}
		if err != nil {
			logs.ErrorContextf(w.ctx, "ChatStreamResponse error reading line: %v", err)
			handleError(resault, contentBuilder, reasoningBuilder)
			return resault, err
		}

		// 跳过空行
		if line == "\n" {
			continue
		}

		// 去除前缀 "data: "
		line = strings.TrimPrefix(line, "data: ")
		// 检查是否是结束标记
		if strings.TrimSpace(line) == "[DONE]" {
			token, err := GetUsage(storedLinedata)
			if err != nil {
				handleError(resault, contentBuilder, reasoningBuilder)
				logs.ErrorContextf(w.ctx, "ChatStreamResponse get usage total failed: %v", err)
				return resault, err
			}
			resault.Usage = *token
			break
		}

		// 解析 JSON 数据
		linedata := &chattype.ChatStreamResponseBody{}
		if err := json.Unmarshal([]byte(line), linedata); err != nil {
			handleError(resault, contentBuilder, reasoningBuilder)
			logs.ErrorContextf(w.ctx, "ChatStreamResponse error unmarshalling json: %v", err)
			return resault, err
		}
		storedLinedata = line
		for _, msg := range linedata.Choices {

			// 如果 reasoning_content 和 content 都为空，跳出循环
			if msg.Delta.ReasoningContent == "" && msg.Delta.Content == "" {
				continue
			}
			// 累积内容
			contentBuilder.WriteString(msg.Delta.Content)
			reasoningBuilder.WriteString(msg.Delta.ReasoningContent)
			// 记录推理结束时间
			if msg.Delta.ReasoningContent == "" && msg.Delta.Content != "" {
				duration := time.Since(reasoningStartTime)
				resault.ReasoningTime = int(duration.Seconds())
			}
			// 构造 WriteResult 结构体
			writeResult := WriteResult{
				ReasoningContent: msg.Delta.ReasoningContent,
				Content:          msg.Delta.Content,
			}
			if stoped, err := sseClient.WriteMessage(w.ctx, w.ctx.Writer, runtime.RequestID(w.ctx), writeResult.String()); err != nil {
				if isBrokenPipeError(err) {
					continue
				}
				logs.ErrorContextf(w.ctx, "ChatStreamResponse error writing message: %v", err)
				handleError(resault, contentBuilder, reasoningBuilder)
				return resault, err
			} else if stoped {
				logs.InfoContextf(w.ctx, "ChatStreamResponse stream Stoped by client")
				handleError(resault, contentBuilder, reasoningBuilder)
				return resault, nil
			}
		}
	}
	resault.Content = contentBuilder.String()
	resault.Reasoning = reasoningBuilder.String()
	// 返回结果
	return resault, nil
}

// ChatStreamResult 内部调用流式响应，onMessage为nil时写入gin中，不为nil时调用onMessage（原版本）
func (w *LLMChatWrapper) ChatStreamResult(resp *http.Response, onMessage func(*chattype.ChatStreamResponseBody) error) (*QaRes, error) {
	defer resp.Body.Close()
	var sseClient *sseclient.SSEClient
	if onMessage == nil {
		// 设置响应头
		sseClient = sseclient.New(sseclient.WithRedisClient(redispool.Redis()),
			sseclient.WithExpiration(time.Minute*5))
		defer sseClient.Close(w.ctx, runtime.RequestID(w.ctx))
		sseClient.SetHeaders(w.ctx.Writer)
	}

	var (
		reader           = bufio.NewReader(resp.Body)
		storedLinedata   string
		contentBuilder   strings.Builder
		reasoningBuilder strings.Builder
	)
	resault := &QaRes{}
	reasoningStartTime := time.Now()
	for {
		// 读取一行数据
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			// 读取完毕,结束请求
			break
		}
		if err != nil {
			logs.ErrorContextf(w.ctx, "ChatStreamResponse error reading line: %v", err)
			handleError(resault, contentBuilder, reasoningBuilder)
			return resault, err
		}

		// 跳过空行
		if line == "\n" {
			continue
		}

		// 去除前缀 "data: "
		line = strings.TrimPrefix(line, "data: ")
		// 检查是否是结束标记
		if strings.TrimSpace(line) == "[DONE]" {
			token, err := GetUsage(storedLinedata)
			if err != nil {
				handleError(resault, contentBuilder, reasoningBuilder)
				logs.ErrorContextf(w.ctx, "ChatStreamResponse get usage total failed: %v", err)
				return resault, err
			}
			resault.Usage = *token
			break
		}

		// 解析 JSON 数据
		linedata := &chattype.ChatStreamResponseBody{}
		if err := json.Unmarshal([]byte(line), linedata); err != nil {
			handleError(resault, contentBuilder, reasoningBuilder)
			logs.ErrorContextf(w.ctx, "ChatStreamResponse error unmarshalling json: %v", err)
			return resault, err
		}
		storedLinedata = line

		// msg := linedata.Choices[0]
		for _, msg := range linedata.Choices {
			// 如果 reasoning_content 和 content 都为空，跳出循环
			if msg.Delta.ReasoningContent == "" && msg.Delta.Content == "" {
				continue
			}
			// 累积内容
			contentBuilder.WriteString(msg.Delta.Content)
			reasoningBuilder.WriteString(msg.Delta.ReasoningContent)
			// 记录推理结束时间
			if msg.Delta.ReasoningContent == "" && msg.Delta.Content != "" {
				duration := time.Since(reasoningStartTime)
				resault.ReasoningTime = int(duration.Seconds())
			}
			if onMessage != nil {
				if err := onMessage(linedata); err != nil {
					handleError(resault, contentBuilder, reasoningBuilder)
					logs.ErrorContextf(w.ctx, "ChatStreamResponse error onMessage: %v", err)
					return resault, err
				}
				continue
			}
			// 构造 WriteResult 结构体
			writeResult := WriteResult{
				ReasoningContent: msg.Delta.ReasoningContent,
				Content:          msg.Delta.Content,
			}
			reqid := runtime.RequestID(w.ctx)
			if stoped, err := sseClient.WriteMessage(w.ctx, w.ctx.Writer, reqid, writeResult.String()); err != nil {
				if isBrokenPipeError(err) {
					continue
				}
				logs.ErrorContextf(w.ctx, "ChatStreamResponse error writing message: %v", err)
				handleError(resault, contentBuilder, reasoningBuilder)
				return resault, err
			} else if stoped {
				logs.InfoContextf(w.ctx, "ChatStreamResponse stream Stoped by client")
				handleError(resault, contentBuilder, reasoningBuilder)
				return resault, nil
			}
		}

	}
	resault.Content = contentBuilder.String()
	resault.Reasoning = reasoningBuilder.String()
	// 返回结果
	return resault, nil
}

// ChatStreamResultSSE 内部调用流式响应，符合 OpenAI SSE 规范
func (w *LLMChatWrapper) ChatStreamResultSSE(resp *http.Response, onMessage func(*chattype.ChatStreamResponseBody) error) (*QaRes, error) {
	defer resp.Body.Close()
	sseClient := sseclient.New(sseclient.WithRedisClient(redispool.Redis()),
		sseclient.WithExpiration(time.Minute*5))
	defer sseClient.Close(w.ctx, runtime.RequestID(w.ctx))
	sseClient.SetHeaders(w.ctx.Writer)

	var (
		reader           = bufio.NewReader(resp.Body)
		storedLinedata   string
		contentBuilder   strings.Builder
		reasoningBuilder strings.Builder
	)
	resault := &QaRes{}
	reasoningStartTime := time.Now()

	// 写入 SSE 格式数据的辅助函数
	writeSSEData := func(data []byte) error {
		chunkData := "data: " + string(data) + "\n\n"
		if _, err := w.ctx.Writer.Write([]byte(chunkData)); err != nil {
			return fmt.Errorf("write SSE data failed: %w", err)
		}
		if flusher, ok := w.ctx.Writer.(http.Flusher); ok {
			flusher.Flush()
		}
		return nil
	}

	for {
		// 读取一行数据
		line, err := reader.ReadString('\n')
		if err == io.EOF {
			// 读取完毕,结束请求
			break
		}
		if err != nil {
			logs.ErrorContextf(w.ctx, "ChatStreamResponse error reading line: %v", err)
			handleError(resault, contentBuilder, reasoningBuilder)
			return resault, err
		}

		// 跳过空行
		if line == "\n" {
			continue
		}

		// 去除前缀 "data: "
		line = strings.TrimPrefix(line, "data: ")
		// 检查是否是结束标记
		if strings.TrimSpace(line) == "[DONE]" {
			token, err := GetUsage(storedLinedata)
			if err != nil {
				handleError(resault, contentBuilder, reasoningBuilder)
				logs.ErrorContextf(w.ctx, "ChatStreamResponse get usage total failed: %v", err)
				return resault, err
			}
			resault.Usage = *token
			break
		}

		// 解析 JSON 数据
		linedata := &chattype.ChatStreamResponseBody{}
		if err := json.Unmarshal([]byte(line), linedata); err != nil {
			handleError(resault, contentBuilder, reasoningBuilder)
			logs.ErrorContextf(w.ctx, "ChatStreamResponse error unmarshalling json: %v", err)
			return resault, err
		}
		storedLinedata = line

		// msg := linedata.Choices[0]
		for _, msg := range linedata.Choices {
			// 如果 reasoning_content 和 content 都为空，跳出循环
			if msg.Delta.ReasoningContent == "" && msg.Delta.Content == "" {
				continue
			}
			// 累积内容
			contentBuilder.WriteString(msg.Delta.Content)
			reasoningBuilder.WriteString(msg.Delta.ReasoningContent)
			// 记录推理结束时间
			if msg.Delta.ReasoningContent == "" && msg.Delta.Content != "" {
				duration := time.Since(reasoningStartTime)
				resault.ReasoningTime = int(duration.Seconds())
			}
			if onMessage != nil {
				if err := onMessage(linedata); err != nil {
					handleError(resault, contentBuilder, reasoningBuilder)
					logs.ErrorContextf(w.ctx, "ChatStreamResponse error onMessage: %v", err)
					return resault, err
				}
				// 将 linedata 序列化为 JSON 并格式化为 SSE 格式
				chunkData, err := json.Marshal(linedata)
				if err != nil {
					handleError(resault, contentBuilder, reasoningBuilder)
					logs.ErrorContextf(w.ctx, "ChatStreamResponse error marshalling linedata: %v", err)
					return resault, fmt.Errorf("marshal chunk failed: %w", err)
				}
				// 写入 SSE 格式的数据
				if err := writeSSEData(chunkData); err != nil {
					handleError(resault, contentBuilder, reasoningBuilder)
					logs.ErrorContextf(w.ctx, "ChatStreamResponse error writing SSE data: %v", err)
					return resault, err
				}
				continue
			}
			// 检查是否被停止
			reqid := runtime.RequestID(w.ctx)
			if stoped, err := sseClient.GetStopSignal(w.ctx, reqid); err != nil {
				logs.ErrorContextf(w.ctx, "ChatStreamResponse error checking stop signal: %v", err)
				handleError(resault, contentBuilder, reasoningBuilder)
				return resault, err
			} else if stoped {
				logs.InfoContextf(w.ctx, "ChatStreamResponse stream Stoped by client")
				handleError(resault, contentBuilder, reasoningBuilder)
				return resault, nil
			}
			// 构造 WriteResult 结构体并格式化为 SSE 格式
			writeResult := WriteResult{
				ReasoningContent: msg.Delta.ReasoningContent,
				Content:          msg.Delta.Content,
			}
			chunkData, err := json.Marshal(writeResult)
			if err != nil {
				logs.ErrorContextf(w.ctx, "ChatStreamResponse error marshalling writeResult: %v", err)
				handleError(resault, contentBuilder, reasoningBuilder)
				return resault, fmt.Errorf("marshal writeResult failed: %w", err)
			}
			// 写入 SSE 格式的数据
			if err := writeSSEData(chunkData); err != nil {
				if isBrokenPipeError(err) {
					continue
				}
				logs.ErrorContextf(w.ctx, "ChatStreamResponse error writing SSE data: %v", err)
				handleError(resault, contentBuilder, reasoningBuilder)
				return resault, err
			}
		}

	}
	resault.Content = contentBuilder.String()
	resault.Reasoning = reasoningBuilder.String()

	// 发送 [DONE] 标记
	if _, err := w.ctx.Writer.Write([]byte("data: [DONE]\n\n")); err != nil {
		logs.ErrorContextf(w.ctx, "ChatStreamResponse error writing [DONE]: %v", err)
		return resault, fmt.Errorf("write [DONE] failed: %w", err)
	}
	if flusher, ok := w.ctx.Writer.(http.Flusher); ok {
		flusher.Flush()
	}

	// 返回结果
	return resault, nil
}

type streamResponse struct {
	Usage *Usage `json:"usage,omitempty"`
}

// GetUsage 获取使用总量,流式和非流式都可使用该函数获取
func GetUsage(prevLine string) (*Usage, error) {
	var data streamResponse
	if err := json.Unmarshal([]byte(prevLine), &data); err != nil {
		return nil, fmt.Errorf("unmarshal response data failed: %w", err)
	}
	if data.Usage == nil {
		return nil, fmt.Errorf("usage not found")
	}
	return data.Usage, nil
}

// isBrokenPipeError 判断是否是 broken pipe 错误
func isBrokenPipeError(err error) bool {
	return strings.Contains(err.Error(), "broken pipe") ||
		strings.Contains(err.Error(), "connection reset by peer")
}

// handleError 处理中途错误等操作的token计算
func handleError(resault *QaRes, contentBuilder strings.Builder, reasoningBuilder strings.Builder) {
	resault.Content = contentBuilder.String()
	resault.Reasoning = reasoningBuilder.String()
	an := utf8.RuneCountInString(reasoningBuilder.String()) + utf8.RuneCountInString(contentBuilder.String())
	if an > 0 {
		resault.Usage.PromptCacheMissTokens = an / 2
		return
	}
}

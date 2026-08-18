package forestctl

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/insmtx/corekg/apps/keapi/internal/dto/dtokeapi"
	"github.com/insmtx/corekg/apps/keapi/internal/services/svcforestchat"
	"github.com/ygpkg/yg-go/apis/apiobj"
)

// ChatCompletions 对话接口
func ChatCompletions(ctx *gin.Context) {
	req := &dtokeapi.ChatCompletionsRequest{}
	baseResp := &apiobj.BaseResponse{}
	if err := ctx.ShouldBindJSON(req); err != nil {
		writeOpenAIError(ctx, http.StatusBadRequest, "invalid request body", "invalid_request_error", "invalid_request_body")
		return
	}
	if !req.ValidChatCompletions(baseResp) {
		writeOpenAIError(ctx, http.StatusBadRequest, baseResp.Message, "invalid_request_error", baseResp.Message)
		return
	}

	created := time.Now().Unix()
	model := "forest-chat"

	var (
		result        *svcforestchat.Result
		err           error
		streamPrinter *svcforestchat.OpenAIPrinter
	)
	if req.Request.Stream {
		streamPrinter = svcforestchat.NewOpenAIPrinter(ctx, "", created, model)
		result, err = svcforestchat.RunWithPrinter(ctx, req, streamPrinter)
	} else {
		result, err = svcforestchat.RunWithPrinter(ctx, req, svcforestchat.NewNoopPrinter())
	}
	if err != nil {
		if req.Request.Stream && streamPrinter != nil && streamPrinter.HasStarted() {
			_ = streamPrinter.Finish("")
			return
		}
		if errors.Is(err, svcforestchat.ErrInvalidChatMessages) || errors.Is(err, svcforestchat.ErrInvalidForestFiles) {
			writeOpenAIError(ctx, http.StatusBadRequest, err.Error(), "invalid_request_error", "invalid_request")
			return
		}
		if errors.Is(err, svcforestchat.ErrChatSessionNotFound) {
			writeOpenAIError(ctx, http.StatusNotFound, err.Error(), "invalid_request_error", "session_not_found")
			return
		}
		if errors.Is(err, svcforestchat.ErrChatModelNotFound) {
			writeOpenAIError(ctx, http.StatusBadRequest, err.Error(), "invalid_request_error", "model_not_found")
			return
		}
		writeOpenAIError(ctx, http.StatusInternalServerError, err.Error(), "server_error", "chat_failed")
		return
	}

	if !req.Request.Stream {
		ctx.JSON(http.StatusOK, dtokeapi.OpenAIChatCompletion{
			ID:      result.ID,
			Object:  "chat.completion",
			Created: created,
			Model:   model,
			Choices: []dtokeapi.OpenAIChatChoice{
				{
					Index: 0,
					Message: dtokeapi.OpenAIChatMessage{
						Role:    "assistant",
						Content: result.Answer,
					},
					FinishReason: "stop",
				},
			},
			Usage: result.Usage,
		})
		return
	}

	if streamPrinter != nil {
		_ = streamPrinter.Finish(result.Answer)
	}
}

func writeOpenAIError(ctx *gin.Context, statusCode int, message, errorType, code string) {
	ctx.AbortWithStatusJSON(statusCode, gin.H{
		"error": gin.H{
			"message": message,
			"type":    errorType,
			"code":    code,
		},
	})
}

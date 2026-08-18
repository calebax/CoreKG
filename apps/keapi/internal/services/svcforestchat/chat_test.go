package svcforestchat

import (
	"errors"
	"testing"

	"github.com/insmtx/corekg/apps/keapi/internal/dto/dtokeapi"
	chatcore "github.com/insmtx/corekg/apps/kechat/chat/core"
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
	"github.com/insmtx/corekg/apps/kellm/models/kellmtype"
)

func TestBuildChatInputFromSessionQuestions(t *testing.T) {
	pendingQuestion := &chattype.ChatQuestion{
		Source: &chattype.Question{
			Question:     "latest question",
			Status:       chattype.QuestionStatusPending,
			ImageUrlList: []string{"https://example.com/a.png"},
		},
	}
	answeredQuestion := &chattype.ChatQuestion{
		Source: &chattype.Question{
			Question: "previous question",
			Answer:   "previous answer",
			Status:   chattype.QuestionStatusAnswered,
		},
	}

	inputPlan, err := buildChatInputFromSessionQuestions([]*chattype.ChatQuestion{
		answeredQuestion,
		nil,
		{Source: nil},
		pendingQuestion,
	})
	if err != nil {
		t.Fatalf("buildChatInputFromSessionQuestions() error = %v", err)
	}
	if inputPlan.QuestionEntity != pendingQuestion {
		t.Fatalf("QuestionEntity = %#v, want pending question", inputPlan.QuestionEntity)
	}
	if inputPlan.Question != "latest question" {
		t.Fatalf("Question = %q, want %q", inputPlan.Question, "latest question")
	}
	if len(inputPlan.Images) != 1 || inputPlan.Images[0] != "https://example.com/a.png" {
		t.Fatalf("Images = %#v, want pending question images", inputPlan.Images)
	}
}

func TestBuildChatInputFromSessionQuestionsRejectsAnsweredLastMessage(t *testing.T) {
	_, err := buildChatInputFromSessionQuestions([]*chattype.ChatQuestion{
		{
			Source: &chattype.Question{
				Question: "already answered",
				Answer:   "answer",
				Status:   chattype.QuestionStatusAnswered,
			},
		},
	})
	if !errors.Is(err, ErrInvalidChatMessages) {
		t.Fatalf("error = %v, want ErrInvalidChatMessages", err)
	}
}

func TestBuildChatInputFromSessionQuestionsRejectsEmptySession(t *testing.T) {
	_, err := buildChatInputFromSessionQuestions(nil)
	if !errors.Is(err, ErrInvalidChatMessages) {
		t.Fatalf("error = %v, want ErrInvalidChatMessages", err)
	}
}

func TestBuildCurrentChatInputSeparatesSystemPrompt(t *testing.T) {
	question, images, summaryPrompt, err := buildCurrentChatInput([]kellmtype.Message{
		{
			Role: "system",
			Content: kellmtype.MessageContent{
				Text: "answer in Chinese",
			},
		},
		{
			Role: "user",
			Content: kellmtype.MessageContent{
				Text: "hello",
				Items: []kellmtype.MessageContentItem{
					{
						Type: "image_url",
						ImageURL: &kellmtype.ImageURL{
							URL: "https://example.com/a.png",
						},
					},
				},
			},
		},
	})
	if err != nil {
		t.Fatalf("buildCurrentChatInput() error = %v", err)
	}
	if question != "hello\nhttps://example.com/a.png" {
		t.Fatalf("question = %q, want user content only", question)
	}
	if summaryPrompt != "answer in Chinese" {
		t.Fatalf("summaryPrompt = %q, want system content", summaryPrompt)
	}
	if len(images) != 1 || images[0] != "https://example.com/a.png" {
		t.Fatalf("images = %#v, want user image", images)
	}
}

func TestBuildChatInputSeparatesSystemPromptFromHistoryAndCurrentQuestion(t *testing.T) {
	historyPairs, question, images, summaryPrompt, err := buildChatInput([]kellmtype.Message{
		{
			Role: "system",
			Content: kellmtype.MessageContent{
				Text: "summary instruction",
			},
		},
		{
			Role: "user",
			Content: kellmtype.MessageContent{
				Text: "previous question",
			},
		},
		{
			Role: "assistant",
			Content: kellmtype.MessageContent{
				Text: "previous answer",
			},
		},
		{
			Role: "user",
			Content: kellmtype.MessageContent{
				Text: "current question",
			},
		},
	})
	if err != nil {
		t.Fatalf("buildChatInput() error = %v", err)
	}
	if len(historyPairs) != 1 {
		t.Fatalf("len(historyPairs) = %d, want 1", len(historyPairs))
	}
	if historyPairs[0].Question != "previous question" || historyPairs[0].Answer != "previous answer" {
		t.Fatalf("historyPairs[0] = %#v, want previous user/assistant pair", historyPairs[0])
	}
	if question != "current question" {
		t.Fatalf("question = %q, want current user content only", question)
	}
	if summaryPrompt != "summary instruction" {
		t.Fatalf("summaryPrompt = %q, want system content", summaryPrompt)
	}
	if len(images) != 0 {
		t.Fatalf("images = %#v, want empty", images)
	}
}

func TestBuildChatModelOptions(t *testing.T) {
	t.Run("default temperature only", func(t *testing.T) {
		options := buildChatModelOptions(&dtokeapi.ChatCompletionsRequest{})

		if options.Temperature == nil || *options.Temperature != defaultChatCompletionsTemperature {
			t.Fatalf("Temperature = %v, want %v", valueOf(options.Temperature), defaultChatCompletionsTemperature)
		}
		if options.TopP != nil {
			t.Fatalf("TopP = %v, want nil", valueOf(options.TopP))
		}
		if options.PresencePenalty != nil {
			t.Fatalf("PresencePenalty = %v, want nil", valueOf(options.PresencePenalty))
		}
	})

	t.Run("explicit zero values are preserved", func(t *testing.T) {
		temperature := float32(0)
		topP := float32(1)
		presencePenalty := float32(0)
		req := &dtokeapi.ChatCompletionsRequest{}
		req.Request.Temperature = &temperature
		req.Request.TopP = &topP
		req.Request.PresencePenalty = &presencePenalty

		options := buildChatModelOptions(req)

		if options.Temperature == nil || *options.Temperature != 0 {
			t.Fatalf("Temperature = %v, want 0", valueOf(options.Temperature))
		}
		if options.TopP == nil || *options.TopP != 1 {
			t.Fatalf("TopP = %v, want 1", valueOf(options.TopP))
		}
		if options.PresencePenalty == nil || *options.PresencePenalty != 0 {
			t.Fatalf("PresencePenalty = %v, want 0", valueOf(options.PresencePenalty))
		}
	})

	t.Run("response format is converted without validation", func(t *testing.T) {
		req := &dtokeapi.ChatCompletionsRequest{}
		req.Request.ResponseFormat = &kellmtype.ResponseFormat{Type: "unknown_format"}

		options := buildChatModelOptions(req)

		if options.ResponseFormat == nil {
			t.Fatalf("ResponseFormat = nil, want converted response format")
		}
		if string(options.ResponseFormat.Type) != "unknown_format" {
			t.Fatalf("ResponseFormat.Type = %q, want unknown_format", options.ResponseFormat.Type)
		}
	})
}

func TestBuildChatContextExtra(t *testing.T) {
	enableReference := false
	req := &dtokeapi.ChatCompletionsRequest{}
	req.Request.ExtraBody = &dtokeapi.ChatCompletionsExtraBody{
		EnableReference: &enableReference,
	}
	inputPlan := &chatInputPlan{
		SummaryPrompt: "  answer in Chinese  ",
	}

	extra := buildChatContextExtra(req, inputPlan, nil)

	if got := extra[chatcore.ExtraKeyEnableReference]; got != false {
		t.Fatalf("enable_reference extra = %#v, want false", got)
	}
	if got := extra[chatcore.ExtraKeySummarySystemPrompt]; got != "answer in Chinese" {
		t.Fatalf("summary prompt extra = %#v, want trimmed prompt", got)
	}
}

func valueOf(v *float32) any {
	if v == nil {
		return nil
	}
	return *v
}

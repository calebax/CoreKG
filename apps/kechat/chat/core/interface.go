package core

import (
	"context"
)

type ChatMode interface {
	Run(ctx context.Context, c *ChatContext) (*ChatResult, error)
}

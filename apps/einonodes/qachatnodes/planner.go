package qachatnodes

import (
	"context"

	"github.com/cloudwego/eino/compose"
)

func NewPlanner[I, O any](ctx context.Context) *compose.Graph[I, O] {
	cag := compose.NewGraph[I, O]()
	return cag
}

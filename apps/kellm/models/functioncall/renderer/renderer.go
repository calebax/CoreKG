package renderer

import "github.com/insmtx/corekg/apps/kellm/models/kellmtype"

type Renderer interface {
	Render(messages []kellmtype.Message, tools []kellmtype.Tool) (string, error)
}

func NewRenderer() Renderer {
	return &GenericRenderer{}
}

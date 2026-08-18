package tools

import (
	"context"
	"io"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

func WrapToolWithOutputTransform(t tool.BaseTool, transformer OutputTransformer) tool.BaseTool {
	ih := &infoHelper{info: t.Info}

	var s tool.StreamableTool
	if st, ok := t.(tool.StreamableTool); ok {
		s = st
	}

	if it, ok := t.(tool.InvokableTool); ok {
		if s == nil {
			// 只实现 InvokableTool
			return &outputWrapper{
				infoHelper: ih,
				outputHelper: &outputHelper{
					i:           it.InvokableRun,
					transformer: transformer,
				},
			}
		} else {
			// 同时实现两个接口
			return &combinedOutputWrapper{
				infoHelper: ih,
				outputHelper: &outputHelper{
					i:           it.InvokableRun,
					transformer: transformer,
				},
				streamOutputHelper: &streamOutputHelper{
					s:           s.StreamableRun,
					transformer: transformer,
				},
			}
		}
	}

	if s != nil {
		// 只实现 StreamableTool
		return &streamOutputWrapper{
			infoHelper: ih,
			streamOutputHelper: &streamOutputHelper{
				s:           s.StreamableRun,
				transformer: transformer,
			},
		}
	}

	return t
}

type ErrorHandler func(context.Context, error) string

type OutputTransformer func(context.Context, string) (string, error)

type infoHelper struct {
	info func(context.Context) (*schema.ToolInfo, error)
}

func (i *infoHelper) Info(ctx context.Context) (*schema.ToolInfo, error) {
	return i.info(ctx)
}

type outputHelper struct {
	i           func(context.Context, string, ...tool.Option) (string, error)
	transformer OutputTransformer
}

func (o *outputHelper) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	output, err := o.i(ctx, argumentsInJSON, opts...)
	if err != nil {
		return "", err
	}
	return o.transformer(ctx, output)
}

type streamOutputHelper struct {
	s           func(context.Context, string, ...tool.Option) (*schema.StreamReader[string], error)
	transformer OutputTransformer
}

func (s *streamOutputHelper) StreamableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (*schema.StreamReader[string], error) {
	streamReader, err := s.s(ctx, argumentsInJSON, opts...)
	if err != nil {
		return nil, err
	}

	return wrapStreamReaderWithTransform(streamReader, s.transformer, ctx), nil
}

func wrapStreamReaderWithTransform(sr *schema.StreamReader[string], transformer OutputTransformer, ctx context.Context) *schema.StreamReader[string] {
	outReader, outWriter := schema.Pipe[string](1)

	go func() {
		defer outWriter.Close()

		var chunks []string
		for {
			chunk, err := sr.Recv()
			if err != nil {
				if err == io.EOF {
					break
				}
				outWriter.Close()
				return
			}
			chunks = append(chunks, chunk)
		}

		// 合并所有 chunks 并转换
		combined := strings.Join(chunks, "")
		transformed, err := transformer(ctx, combined)
		if err != nil {
			outWriter.Close()
			return
		}

		outWriter.Send(transformed, nil)
	}()

	return outReader
}

type outputWrapper struct {
	*infoHelper
	*outputHelper
}

type streamOutputWrapper struct {
	*infoHelper
	*streamOutputHelper
}

type combinedOutputWrapper struct {
	*infoHelper
	*outputHelper
	*streamOutputHelper
}

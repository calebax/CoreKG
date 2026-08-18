package dtodevtool

import (
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type RewriteMarkdownURLResponse struct {
	apiobj.BaseResponse
	Response RewriteMarkdownURLEmbedResponse
}

type RewriteMarkdownURLEmbedResponse struct {
}

type StatAlgoMarkdownResponse struct {
	apiobj.BaseResponse
	Response StatAlgoMarkdownEmbedResponse
}
type StatAlgoMarkdownEmbedResponse struct {
	Count int    `json:"count"` // markdown 文件数量
	Path  string `json:"path"`  // 统计的路径
}

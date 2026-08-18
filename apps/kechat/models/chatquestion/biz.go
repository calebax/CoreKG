package chatquestion

import (
	"github.com/insmtx/corekg/apps/kechat/models/chattype"
)

// createRespnse 创建问题返回值获取id结构
type createRespnse struct {
	ID string `json:"_id"`
}

// QuestionSearchResult es 查询问题结构体
type QuestionSearchResult struct {
	ScrollID string `json:"_scroll_id"`
	Hits     struct {
		MaxScore float64 `json:"max_score"`
		Total    struct {
			Value int64 `json:"value"`
		} `json:"total"`
		Hits []*chattype.ChatQuestion `json:"hits"`
	} `json:"hits"`
}

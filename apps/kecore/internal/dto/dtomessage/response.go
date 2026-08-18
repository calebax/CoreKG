package dtomessage

import (
	"time"

	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type ListMessageResponse struct {
	apiobj.BaseResponse
	Response ListMessageEmbedResponse
}

type ListMessageEmbedResponse struct {
	Data []Message `json:"data"`
	apiobj.QueryResponse
}

type Message struct {
	// 消息ID
	ID uint `json:"id"`
	// 创建时间
	CreatedAt time.Time `json:"created_at"`
	// 消息标题
	Title string `json:"title"`
	// 消息内容
	Content string `json:"content"`
	// 消息模板类型
	TemplateType foresttype.MessageTemplateType `json:"template_type"`
	// 消息路由路径
	RoutePath string `json:"route_path"`
	// 消息已读状态
	ReadStatus foresttype.MessageReadStatus `json:"read_status"`
	// 消息已读时间
	ReadAt *time.Time `json:"read_at"`
}

type SetMessageStatusResponse struct {
	apiobj.BaseResponse
	Response SetMessageStatusEmbedResponse
}
type SetMessageStatusEmbedResponse struct {
}

type DeleteMessagesResponse struct {
	apiobj.BaseResponse
	Response DeleteMessagesEmbedResponse
}
type DeleteMessagesEmbedResponse struct {
}

type GetMessageCountResponse struct {
	apiobj.BaseResponse
	Response GetMessageCountEmbedResponse
}
type GetMessageCountEmbedResponse struct {
	// 消息数量
	Count int64 `json:"count"`
}

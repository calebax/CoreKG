package dtomessage

import (
	"github.com/insmtx/corekg/apps/kecore/models/foresttype"
	"github.com/ygpkg/yg-go/apis/apiobj"
)

type ListMessageRequest struct {
	apiobj.BaseRequest
	Request ListMessageEmbedRequest
}

type ListMessageEmbedRequest struct {
	apiobj.PageQuery
}

func (opt *ListMessageRequest) Validity(resp *ListMessageResponse) {
}

type SetMessageStatusRequest struct {
	apiobj.BaseRequest
	Request SetMessageStatusEmbedRequest
}
type SetMessageStatusEmbedRequest struct {
	// 消息ID
	MessageID uint `json:"message_id"`
	// 消息已读状态
	Status foresttype.MessageReadStatus `json:"status"`
}

func (opt *SetMessageStatusRequest) Validity(resp *SetMessageStatusResponse) {
}

type DeleteMessagesRequest struct {
	apiobj.BaseRequest
	Request DeleteMessagesEmbedRequest
}
type DeleteMessagesEmbedRequest struct {
	// 消息ID列表
	MessageIDs []uint `json:"message_ids"`
	// DeleteAll 是否删除所有消息
	DeleteAll bool `json:"delete_all"`
}

func (opt *DeleteMessagesRequest) Validity(resp *DeleteMessagesResponse) {
}

type GetMessageCountRequest struct {
	apiobj.BaseRequest
	Request GetMessageCountEmbedRequest
}
type GetMessageCountEmbedRequest struct {
	apiobj.PageQuery
}

func (opt *GetMessageCountRequest) Validity(resp *GetMessageCountResponse) {
}

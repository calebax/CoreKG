package wecoms

type MessageType string

// MessageTypeText 文本消息
const MessageTypeText MessageType = "text"

// MessageTypeImage 图片消息
const MessageTypeImage MessageType = "image"

// MessageTypeVoice 语音消息
const MessageTypeVoice MessageType = "voice"

// MessageTypeVideo 视频消息
const MessageTypeVideo MessageType = "video"

// MessageTypeLocation 位置消息
const MessageTypeLocation MessageType = "location"

// MessageTypeLink 链接消息
const MessageTypeLink MessageType = "link"

// MessageTypeEvent 事件消息
const MessageTypeEvent MessageType = "event"

// EventType 事件类型
type EventType string

// EventTypeLocation
const EventTypeLocation EventType = "LOCATION"

// EventTypeChangeExternalContact 企业客户事件
const EventTypeChangeExternalContact EventType = "change_external_contact"

// EventTypeChangeExternalChat 客户群变更事件
const EventTypeChangeExternalChat EventType = "change_external_chat"

// EventTypeSysApprovalChange 审批申请状态变化回调通知
const EventTypeSysApprovalChange EventType = "sys_approval_change"

// ChangeType 变更类型
type ChangeType string

// ChangeTypeAddExternalContact 添加企业客户事件
const ChangeTypeAddExternalContact ChangeType = "add_external_contact"

// ChangeTypeEditExternalContact 编辑企业客户事件
const ChangeTypeEditExternalContact ChangeType = "edit_external_contact"

// ChangeTypeAddHalfExternalContact 外部联系人免验证添加成员事件
const ChangeTypeAddHalfExternalContact ChangeType = "add_half_external_contact"

// ChangeTypeDelExternalContact 删除企业客户事件
const ChangeTypeDelExternalContact ChangeType = "del_external_contact"

// ChangeTypeDelFollowUser 删除跟进成员事件
const ChangeTypeDelFollowUser ChangeType = "del_follow_user"

// ChangeTypeTransferFail 客户接替失败事件
const ChangeTypeTransferFail ChangeType = "transfer_fail"

type ReceiveMessage struct {
	ToUserName   string      `xml:"ToUserName"`
	FromUserName string      `xml:"FromUserName"`
	CreateTime   int64       `xml:"CreateTime"`
	MsgType      MessageType `xml:"MsgType"`
	MsgId        int64       `xml:"MsgId"`
	AgentID      int64       `xml:"AgentID"`

	Content    string     `xml:"Content,omitempty"`
	Event      EventType  `xml:"Event,omitempty"`
	ChangeType ChangeType `xml:"ChangeType,omitempty"`

	MediaID string `xml:"MediaId,omitempty"`
	// PicURL 图片链接
	PicURL string `xml:"PicUrl,omitempty"`
	// Format 语音格式，如amr，speex等
	Format string
	// ThumbMediaID 视频消息缩略图的媒体id，可以调用获取媒体文件接口拉取数据，仅三天内有效
	ThumbMediaID string `xml:"ThumbMediaId,omitempty"`

	// Lat 地理位置纬度
	Lat float64 `xml:"Latitude,omitempty"`
	// Lon 地理位置经度
	Lon float64 `xml:"Longitude,omitempty"`

	// Lat 地理位置纬度
	LatX float64 `xml:"Location_X,omitempty"`
	// Lon 地理位置经度
	LonY float64 `xml:"Location_Y,omitempty"`
	// Sca 地图缩放大小
	Scale int `xml:"Scale,omitempty"`
	// Lab 地理位置信息
	Label string `xml:"Label,omitempty"`

	// App app类型，在企业微信固定返回wxwork，在微信不返回该字段
	AppType string `xml:"AppType,omitempty"`

	// Title 标题
	Title string `xml:"Title,omitempty"`
	// Description 描述
	Description string `xml:"Description,omitempty"`
	// URL 链接跳转的url
	URL string `xml:"Url,omitempty"`
	// PicURL 封面缩略图的url
	// PicURL string `xml:"PicUrl,omitempty"`

	// UserID 企业服务人员的UserID
	UserID string `xml:"UserID,omitempty"`
	// ExternalUserID 外部联系人的userid，注意不是企业成员的帐号
	ExternalUserID string `xml:"ExternalUserID,omitempty"`
	// State 添加此用户的「联系我」方式配置的state参数，可用于识别添加此用户的渠道
	State string `xml:"State,omitempty"`
	// WelcomeCode 欢迎语code，可用于发送欢迎语
	WelcomeCode string `xml:"WelcomeCode,omitempty"`
}

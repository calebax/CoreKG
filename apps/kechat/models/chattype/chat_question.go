package chattype

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/insmtx/corekg/apps/kecore/models/nebulagraph"
	"github.com/insmtx/corekg/apps/ketask/models/ragtypes"
	agentservice "github.com/insmtx/corekg/pkgs/einotools/service"
)

const (
	HistoryIndex = "ke_chat_history"
)

type ChatQuestion struct {
	ID     string    `json:"_id"`
	Score  float64   `json:"_score"`
	Source *Question `json:"_source"`
}

// Question ES question结构体
type Question struct {
	ReqID              string              `json:"req_id"`
	AgentStep          int                 `json:"agent_step"`
	CompanyID          uint                `json:"company_id"`
	Uin                uint                `json:"uin"`
	ApiKeyID           uint                `json:"api_key_id"`
	OutToken           int                 `json:"out_token"`
	CacheHitToken      int                 `json:"cache_hit_token"`
	CacheMissToken     int                 `json:"cache_miss_token"`
	TotalTokens        int                 `json:"total_tokens"`
	IsCharged          bool                `json:"is_charged"`
	BaseAgentID        uint                `json:"base_agent_id"`
	AgentVersion       uint                `json:"agent_version"`
	ModelID            uint                `json:"model_id"`
	CostSeconds        int                 `json:"cost_seconds"`
	ReasoningSeconds   int                 `json:"reasoning_seconds"`
	ImageUrlList       []string            `json:"image_url_list"`
	Status             QuestionStatus      `json:"status"`
	ExternalID         string              `json:"external_id"`
	SessionID          uint                `json:"session_id"`
	CreatedAt          time.Time           `json:"created_at"`
	Question           string              `json:"question"`
	Answer             string              `json:"answer"`
	Reasoning          string              `json:"reasoning"`
	ImageContent       string              `json:"image_content"`
	QueryReferenceList *QueryReferenceList `json:"query_reference_list"`
	ChatReferenceList  *ChatReferenceList  `json:"chat_reference_list"`

	// UserInput 用户完整输入
	UserInput *ChatRequestBody `json:"user_input"`
	// AgentName
	AgentName string `json:"agent_name"`

	// SubQuestion 拆的根据问答结果识别的子问题
	SubQuestion []string `json:"sub_question"`

	// 拓展数据，工具开关
	Extra *ExtraInfo `json:"extra,omitempty"`

	GraphReference     *nebulagraph.Graph `json:"graph_reference"`
	GraphChatReference *nebulagraph.Graph `json:"graph_chat_reference"`

	ReactAgentService *agentservice.ReactAgentService `json:"react_agent_service"`
}

func (q *Question) String() string {
	jsonPayload, err := json.Marshal(q)
	if err != nil {
		// logs.Errorf("[ChatReqBody] Failed to convert request body to JSON: %s", err.Error())
		return ""
	}
	return string(jsonPayload)
}

type QuestionStatus string

const (
	QuestionStatusPending  QuestionStatus = "pending"
	QuestionStatusAnswered QuestionStatus = "answered"
	QuestionStatusTimeout  QuestionStatus = "timeout"
	QuestionStatusStop     QuestionStatus = "stop"
	QuestionStatusError    QuestionStatus = "error"
)

type ExtraInfo struct {
	Agent *AgentExtra `json:"agent,omitempty"`
	Input *InputExtra `json:"input,omitempty"`
}

type AgentExtra struct {
	EnableWebSearch bool `json:"enable_web_search"`
}

type InputExtra struct {
	Attachments []AttachmentInfo `json:"attachments,omitempty"`
}

type AttachmentInfo struct {
	Url string `json:"url"`
	// 文件解析后，生成md的产物
	MdUrl string `json:"md_url,omitempty"`
	Type  string `json:"type"` // image / pdf / doc / csv / url
	Name  string `json:"name,omitempty"`
}

type ChatReference struct {
	FileID   uint           `json:"file_id"`
	Abstract string         `json:"abstract"`
	Chunks   map[int]string `json:"chunks"`
}

type ChatReferenceChunkList []ChatReference

type ChatReferenceList struct {
	Reference ChatReferenceChunkList `json:"reference"`
}

// QueryReference 聊天的引用内容
type QueryReference struct {
	FileID    uint      `json:"file_id"`
	FileName  string    `json:"file_name"`
	ForestID  uint      `json:"forest_id"`
	UserName  string    `json:"user_name"`
	AvatarURL string    `json:"avatar_url"`
	CreatedAt time.Time `json:"created_at"`
	Uin       uint      `json:"uin"`
	Abstract  string    `json:"abstract"`
	// DataSourceType 数据源类型 DC
	DataSourceType DataSourceType `json:"data_source"`
	// ChatReferenceChunk `json:",inline"`
	ChunkList QueryReferenceChunkList `json:"chunk_list"`
}

// QueryReferenceChunk 引用内容
type QueryReferenceChunk struct {
	Type     string            `json:"type"`
	ChunkID  string            `json:"chunk_id"`
	Sequence int               `json:"sequence"`
	Content  string            `json:"content"`
	ImageURL string            `json:"image_url"`
	Score    float64           `json:"score"`
	Location ragtypes.Location `json:"location"`
}

// ChatReferenceChunkList 引用内容列表
type QueryReferenceChunkList []QueryReferenceChunk

// ChatReferenceList 引用内容列表
type QueryReferenceList []*QueryReference

type DataSourceType string

const (
	DataSourceTypeDC DataSourceType = "DC"
)

func (ep QueryReferenceList) Value() (driver.Value, error) {
	return json.Marshal(ep)
}

func (ep *QueryReferenceList) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("invalid type for ExamPosition")
	}
	return json.Unmarshal(bytes, ep)
}

func (crl QueryReferenceList) Len() int {
	return len(crl)
}

func (crl QueryReferenceList) Less(i, j int) bool {
	if len(crl[i].ChunkList) > len(crl[j].ChunkList) {
		return true
	}
	return crl[i].FileID < crl[j].FileID
}

func (crl QueryReferenceList) Swap(i, j int) {
	crl[i], crl[j] = crl[j], crl[i]
}

func (crcl QueryReferenceChunkList) Len() int {
	return len(crcl)
}

func (crcl QueryReferenceChunkList) Less(i, j int) bool {
	return crcl[i].Sequence < crcl[j].Sequence
}

func (crcl QueryReferenceChunkList) Swap(i, j int) {
	crcl[i], crcl[j] = crcl[j], crcl[i]
}

func (crcl *QueryReferenceChunkList) DeduplicateByContent() {
	seen := make(map[string]bool)
	writeIndex := 0

	for _, chunk := range *crcl {
		if !seen[chunk.Content] {
			seen[chunk.Content] = true
			(*crcl)[writeIndex] = chunk
			writeIndex++
		}
	}

	*crcl = (*crcl)[:writeIndex]
}

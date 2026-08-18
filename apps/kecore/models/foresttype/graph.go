package foresttype

import (
	"time"

	"github.com/ygpkg/yg-go/types"
	"gorm.io/gorm"
)

type GraphStatus string

const (
	GraphStatusUnCreated GraphStatus = "uncreated" // 未创建
	GraphStatusDraft     GraphStatus = "draft"     // 未发布
	GraphStatusPending   GraphStatus = "pending"   // 排队中
	GraphStatusRunning   GraphStatus = "running"   // 运行中
	GraphStatusSuccess   GraphStatus = "success"   // 运行成功
	GraphStatusFailed    GraphStatus = "failed"    // 运行失败
	GraphStatusUpdatable GraphStatus = "updatable" // 可更新
)

type ParseMode string

const (
	ParseModeAuto ParseMode = "auto"
	ParseModeRule ParseMode = "rule"
)

// ForestGraph 知识图谱
type ForestGraph struct {
	gorm.Model
	Uin         uint   `gorm:"column:uin;type:bigint;not null;default:0;comment:'用户ID';index" json:"uin"`
	CompanyID   uint   `gorm:"column:company_id;type:bigint;not null;default:0;comment:'公司ID';index" json:"company_id"`
	Name        string `gorm:"type:varchar(255);column:name;not null;comment:图谱名称;index" json:"name"`
	Description string `gorm:"type:varchar(255);column:description;not null;default:'';comment:知识森林描述" json:"description"`
	// PublicScope 公开范围
	PublicScope PublicScope `gorm:"column:public_scope;type:varchar(32);not null;default:'company';comment:公开范围" json:"public_scope"`
	ForestID    uint        `gorm:"column:forest_id;type:bigint;not null;default:0;comment:'知识森林ID';index" json:"forest_id"`
	VersionID   uint        `gorm:"column:version_id;type:bigint;not null;default:0;comment:'知识图谱版本ID';index" json:"version_id"`
	//Avatar 知识图谱头像
	AvatarUrl string `gorm:"type:varchar(255);column:avatar_url;not null;default:'';comment:知识图谱头像" json:"avatar_url"`
}

type ForestGraphList []ForestGraph

func (ForestGraph) TableName() string {
	return TableNameKeForestGraph
}

// ForestGraph 知识图谱
type ForestGraphVersion struct {
	gorm.Model
	Uin         uint            `gorm:"column:uin;type:bigint;not null;default:0;comment:'用户ID';index" json:"uin"`
	CompanyID   uint            `gorm:"column:company_id;type:bigint;not null;default:0;comment:'公司ID';index" json:"company_id"`
	GraphID     uint            `gorm:"column:graph_id;type:bigint;not null;default:0;comment:'图谱ID';index" json:"graph_id"`
	Name        string          `gorm:"type:varchar(255);column:name;not null;comment:图谱名称" json:"name"`
	Status      GraphStatus     `gorm:"type:varchar(64);column:status;not null;default:draft;comment:知识图谱状态" json:"status"`
	Description string          `gorm:"type:varchar(255);column:description;not null;default:'';comment:知识森林描述" json:"description"`
	FileIDList  types.UintArray `gorm:"column:file_id_list;type:text;comment:'文件ID列表'" json:"file_id_list"`
	SpaceName   string          `gorm:"type:varchar(255);column:space_name;not null;default:'';comment:知识图谱空间名称;unique" json:"space_name"`
	ParseMode   ParseMode       `gorm:"type:varchar(255);column:parse_mode;not null;default:'auto';comment:解析模式" json:"parse_mode"`
	// PublicScope 公开范围
	PublicScope PublicScope `gorm:"column:public_scope;type:varchar(32);not null;default:'company';comment:公开范围" json:"public_scope"`
	//Avatar 知识图谱头像
	AvatarUrl string `gorm:"type:varchar(511);column:avatar_url;not null;default:'';comment:知识图谱头像" json:"avatar_url"`
}

func (ForestGraphVersion) TableName() string {
	return TableNameKeForestGraphVersion
}

// ForestGraphInfo 知识图谱详细信息
type ForestGraphInfo struct {
	ID               uint            `gorm:"column:id"` // 适配之前的大写id
	CreatedAt        time.Time       `json:"created_at"`
	UpdatedAt        time.Time       `json:"updated_at"`
	DeletedAt        gorm.DeletedAt  `json:"deleted_at"`
	Uin              uint            `gorm:"column:uin;type:bigint;not null;default:0;comment:'用户ID';index" json:"uin"`
	CompanyID        uint            `gorm:"column:company_id;type:bigint;not null;default:0;comment:'公司ID';index" json:"company_id"`
	Name             string          `gorm:"type:varchar(255);column:name;not null;comment:图谱名称;index" json:"name"`
	Description      string          `gorm:"type:varchar(255);column:description;not null;default:'';comment:知识森林描述" json:"description"`
	ParseMode        ParseMode       `gorm:"type:varchar(255);column:parse_mode;not null;default:'auto';comment:解析模式" json:"parse_mode"`
	PublicScope      PublicScope     `gorm:"column:public_scope;type:varchar(32);not null;default:'company';comment:公开范围" json:"public_scope"`
	ForestID         uint            `gorm:"column:forest_id;type:bigint;not null;default:0;comment:'知识森林ID';index" json:"forest_id"`
	VersionID        uint            `gorm:"column:version_id;type:bigint;not null;default:0;comment:'知识图谱版本ID';index" json:"version_id"`
	Status           GraphStatus     `gorm:"type:varchar(64);column:status;not null;default:draft;comment:知识图谱状态" json:"status"`
	SpaceName        string          `gorm:"type:varchar(255);column:space_name;not null;default:'';comment:知识图谱空间名称;unique" json:"space_name"`
	FileIDList       types.UintArray `gorm:"column:file_id_list;type:text;comment:'文件ID列表'" json:"file_id_list"`
	AvatarUrl        string          `gorm:"type:varchar(255);column:avatar_url;not null;default:'';comment:知识图谱头像" json:"avatar_url"`
	TaskCount        int64           `json:"task_count"`
	SuccessTaskCount int64           `json:"success_task_count"`
}

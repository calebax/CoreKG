package user

import (
	"time"

	"github.com/insmtx/corekg/pkgs/types"
)

// UserInfo 是第三方用户信息
type UserInfo struct {
	GithubID         uint   `json:"github_id,omitempty"`
	WorkWechatUserID string `json:"work_wechat_user_id,omitempty"`
	WechatUnionID    string `json:"wechat_union_id,omitempty"`
	WechatWebOpenID  string `json:"wechat_web_open_id,omitempty"`

	Identify  string    `json:"identify,omitempty"`
	AvatarURL string    `json:"avatar_url,omitempty"`
	Bio       string    `json:"bio,omitempty"`
	Email     string    `json:"email,omitempty"`
	Phone     string    `json:"phone,omitempty"`
	Name      string    `json:"name,omitempty"`
	RealName  string    `json:"real_name,omitempty"`
	IDcard    string    `json:"id_card,omitempty"`
	ID        uint      `json:"id,omitempty"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	Uin       uint      `json:"uin,omitempty"`

	HasPassword     int        `json:"has_password"`
	PasswordChanged types.Bool `json:"password_changed"`
	CompanyQuota    uint       `json:"company_quota"`
}

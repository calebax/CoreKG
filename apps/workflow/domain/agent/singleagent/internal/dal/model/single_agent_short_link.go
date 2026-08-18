package model

const TableNameSingleAgentShortLink = "single_agent_short_link"

// SingleAgentShortLink stores the short link info for a bot.
type SingleAgentShortLink struct {
	ID         int64  `gorm:"column:id;type:bigint unsigned;primaryKey;autoIncrement:false;comment:id" json:"id"`
	BotID      int64  `gorm:"column:bot_id;type:bigint unsigned;not null;index:idx_bot_id;comment:bot id" json:"bot_id"`
	ShortCode  string `gorm:"column:short_code;type:varchar(16);not null;uniqueIndex:uk_short_code;comment:short route code" json:"short_code"`
	UserID     int64  `gorm:"column:user_id;type:bigint unsigned;not null;comment:user id" json:"user_id"`
	SpaceID    int64  `gorm:"column:space_id;type:bigint unsigned;not null;comment:space id" json:"space_id"`
	Status     int32  `gorm:"column:status;type:tinyint;not null;default:0;comment:status 0: public disabled 1: normal" json:"status"`
	LastUsedAt int64  `gorm:"column:last_used_at;type:bigint unsigned;not null;default:0;comment:last used time" json:"last_used_at"`
	UserToken  string `gorm:"column:user_token;type:varchar(255);not null;comment:user token" json:"user_token"`
	CreatedAt  int64  `gorm:"column:created_at;type:bigint unsigned;not null;autoCreateTime:milli;comment:create time" json:"created_at"`
	UpdatedAt  int64  `gorm:"column:updated_at;type:bigint unsigned;not null;autoUpdateTime:milli;comment:update time" json:"updated_at"`
}

// TableName SingleAgentShortLink's table name.
func (*SingleAgentShortLink) TableName() string {
	return TableNameSingleAgentShortLink
}

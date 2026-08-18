package entity

const (
	// ShortLinkStatusPublicDisabled 表示可用但关闭公开访问
	ShortLinkStatusPublicDisabled int32 = 0
	// ShortLinkStatusNormal 表示正常公开访问
	ShortLinkStatusNormal int32 = 1
	ShortLinkStatusActive int32 = ShortLinkStatusPublicDisabled
)

// ShortLink describes the short route bound to a bot.
type ShortLink struct {
	ID         int64
	BotID      int64
	ShortCode  string
	UserID     int64
	SpaceID    int64
	Status     int32
	LastUsedAt int64
	UserToken  string
	CreatedAt  int64
	UpdatedAt  int64
}

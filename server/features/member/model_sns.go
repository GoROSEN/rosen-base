package member

import "gorm.io/gorm"

// SnsID 社交网站ID
type SnsID struct {
	gorm.Model

	SnsType     string `gorm:"comment:三方平台标识;size:16"`         // 三方平台标识
	OpenID      string `gorm:"comment:open id;size:64"`        // OpenID
	AccessToken string `gorm:"comment:access token;size:1024"` // 三方登录令牌
	AvatarURL   string `gorm:"comment:头像URL;size:256"`         // 头像URL
	MemberID    uint   `gorm:"index;comment:关联会员ID"`
}

func (SnsID) TableName() string {
	return "member_sns_ids"
}

// SnsSummary 社交子模块统计数据
type SnsSummary struct {
	MemberID        uint `gorm:"primaryKey;comment:关联会员ID"`
	FollowersCount  uint `gorm:"size:64;comment:粉丝数量"`
	FollowingsCount uint `gorm:"size:64;comment:关注对象数量"`
}

func (SnsSummary) TableName() string {
	return "member_sns_summaries"
}

// SnsFriendRequest 加好友请求
type SnsFriendRequest struct {
	gorm.Model

	SenderID   uint   `gorm:"comment:发起人ID"`
	ReceiverID uint   `gorm:"comment:接收人ID"`
	Message    string `gorm:"size:1024;comment:申请附言"`
	Status     uint   `gorm:"size:8;index;comment:状态：0-待处理；1-同意；2-拒绝"`
	Sender     *Member
	Receiver   *Member
}

func (SnsFriendRequest) TableName() string {
	return "member_sns_friend_requests"
}

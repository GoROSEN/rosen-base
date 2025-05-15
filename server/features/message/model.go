package message

import "gorm.io/gorm"

// Message 消息
type Message struct {
	gorm.Model

	FromID       uint   `gorm:"index;comment:发送人ID"`
	DestUserID   uint   `gorm:"index;comment:接收人ID"`
	DestMemberID uint   `gorm:"index;comment:接收人ID"`
	Title        string `gorm:"size:256;comment:标题"`
	Conent       string `gorm:"comment:内容"`
	Unread       bool   `gorm:"index;size:1;comment:未读"`
}

func (Message) TableName() string {
	return "message_messages"
}

// Destination 接收端
type Destination struct {
	gorm.Model

	UserID   uint   `gorm:"index:idx_uiddev;comment:接收人用户ID"`
	MemberID uint   `gorm:"index:idx_middev;comment:接收人会员ID"`
	Channel  string `gorm:"index;size:32;comment:通道名称"`
	Address  string `gorm:"index:idx_uiddev;index:idx_middev;size:256;comment:接收标识"`
}

func (Destination) TableName() string {
	return "message_destinations"
}

func MigrateDB(db *gorm.DB) {
	db.AutoMigrate(&Message{})
	db.AutoMigrate(&Destination{})
}

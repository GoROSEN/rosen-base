package notification

import "gorm.io/gorm"

type MessageTemplate struct {
	gorm.Model

	Name   string `gorm:"index:idx_name_lang;comment:名称;size:256"`
	Type   string `gorm:"index;comment:模板类别;size:16"`
	Title  string `gorm:"comment:标题;size:256"`
	Body   string `gorm:"comment:内容;size:8192"`
	Lang   string `gorm:"index:idx_name_lang;comment:语言;size:16"`
	Module string `gorm:"comment:所属模块;size:256"`
}

func (MessageTemplate) TableName() string {
	return "notification_message_templates"
}

type MessageLog struct {
	gorm.Model

	TemplateID  uint   `gorm:"index;comment:模板ID"`
	Destination string `gorm:"index;comment:接收人;size:256"`
	Channel     string `gorm:"index;comment:消息通道;size:16"`
	Title       string `gorm:"comment:标题;size:256"`
	Body        string `gorm:"comment:内容;size:8192"`
	Result      string `gorm:"comment:结果;size:256"`
	Module      string `gorm:"comment:所属模块;size:256"`

	MessageTemplate *MessageTemplate `gorm:"ForeignKey:TemplateID"`
}

func (MessageLog) TableName() string {
	return "notification_message_logs"
}

// MigrateDB 更新数据库表结构
func MigrateDB(db *gorm.DB) {

	db.AutoMigrate(&MessageTemplate{})
	db.AutoMigrate(&MessageLog{})
}

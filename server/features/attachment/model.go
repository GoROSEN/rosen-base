package attachment

import (
	"github.com/GoROSEN/rosen-apiserver/core/common"
	"gorm.io/gorm"
)

type Attachment struct {
	common.CrudModel

	UUID            string `gorm:"size:64;unique_index;comment:文件UUID" json:"uuid"`
	ModuleName      string `gorm:"size:32;comment:所属模块名" json:"moduleName"`
	FileName        string `gorm:"size:512;index;comment:文件显示名" json:"filename"`
	OssFilePathName string `gorm:"size:2048;comment:OSS存储中的文件全路径名" json:"ossFilePath"`
	FileHash        string `gorm:"size:65;index;comment:文件哈希" json:"filehash"`
	IsPublic        bool   `gorm:"size:1;comment:是否公开" json:"public"`
}

// TableName 设置表名
func (Attachment) TableName() string {
	return "attachment_attachments"
}

type MemberAttachment struct {
	common.CrudModel
	MemberID     uint        `json:"memberId"`
	AttachmentID uint        `json:"-"`
	Attachment   *Attachment `json:"attachment"`
}

// TableName 设置表名
func (MemberAttachment) TableName() string {
	return "attachment_member_attachments"
}

type UserAttachment struct {
	common.CrudModel
	UserID       uint        `json:"userId"`
	AttachmentID uint        `json:"-"`
	Attachment   *Attachment `json:"attachment"`
}

// TableName 设置表名
func (UserAttachment) TableName() string {
	return "attachment_user_attachments"
}

// MigrateDB 自动建表
func MigrateDB(db *gorm.DB) {

	db.AutoMigrate(
		&Attachment{},
		&MemberAttachment{},
		&UserAttachment{},
	)
}

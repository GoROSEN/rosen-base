// Package member 会员管理
package member

import (
	"github.com/GoROSEN/rosen-apiserver/core/common"
	"gorm.io/gorm"
)

// Member 会员（前台用户）
type Member struct {
	common.CrudModel

	UserName    string     `gorm:"uniqueIndex;comment:用户名;size:256" json:"userName"` // 用户名
	DisplayName string     `gorm:"index;comment:昵称;size:64" json:"displayName"`      // 昵称
	Bio         string     `gorm:"comment:bio;size:1024" json:"bio"`
	CellPhone   string     `gorm:"index;comment:手机号;size:16" json:"cellphone"`  // 手机号
	Email       string     `gorm:"uniqueIndex;comment:邮箱;size:64" json:"email"` // 电邮
	LoginPwd    string     `gorm:"comment:登录密码;size:64" json:"-"`               // 登录密码
	Gender      string     `gorm:"comment:性别;size:8" json:"gender"`             // 性别
	Avatar      string     `gorm:"comment:头像URL;size:1024" json:"avatar"`       // 头像
	IDCard      string     `gorm:"comment:身份证号;size:32" json:"idcard"`          // 身份证、驾驶证号
	SnsIDs      []*SnsID   `json:"-"`                                           // 社交平台OpenID
	Followers   []*Member  `gorm:"many2many:member_sns_followers;" json:"-"`    // 粉丝
	Followings  []*Member  `gorm:"many2many:member_sns_followings;" json:"-"`   // 关注对象
	Blockeds    []*Member  `gorm:"many2many:member_sns_blockeds;" json:"-"`
	Friends     []*Member  `gorm:"many2many:member_sns_friends;" json:"-"`
	SnsSummary  SnsSummary `gorm:"ForeignKey:MemberID" json:"-"` // 社交小计
	Level       uint       `gorm:"index;size:16;comment:级别" json:"level"`
	Language    string     `gorm:"default:en-US;size:32;comment:用户语言" json:"language"`
}

func (Member) TableName() string {
	return "member_users"
}

// MigrateDB 更新数据库表结构
func MigrateDB(db *gorm.DB) {

	db.AutoMigrate(&Member{})
	db.AutoMigrate(&SnsID{})
	db.AutoMigrate(&SnsSummary{})
	db.AutoMigrate(&SnsFriendRequest{})
}

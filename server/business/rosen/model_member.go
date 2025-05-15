package rosen

import (
	"github.com/GoROSEN/rosen-apiserver/core/common"
	"github.com/GoROSEN/rosen-apiserver/features/member"
)

// MemberExtra Rosen会员扩展信息
type MemberExtra struct {
	MemberID              uint               `gorm:"primaryKey;comment:会员ID" json:"memberId"`
	Role                  string             `gorm:"comment:用户角色;size:16" json:"role"`
	Level                 uint               `gorm:"comment:用户级别" json:"level"`
	VirtualImageID        uint               `gorm:"comment:虚拟形象ID" json:"virtualImageId"`
	OccupyLimit           uint               `gorm:"comment:占地上限" json:"occupyLimit"`
	ShareLocation         bool               `gorm:"default:0;comment:是否共享位置信息" json:"shareLocation"`
	CurrentEquip          *Asset             `gorm:"foreignKey:VirtualImageID" json:"currentEquip"`
	EnableChatTranslation bool               `gorm:"size:1;default:0;comment:是否开启翻译" json:"enableChatTranslation"`
	ChatTranslationLang   string             `gorm:"size:32;comment:翻译语言" json:"chatTranslationLang"`
	PayPassword           string             `gorm:"size:64;comment:支付密码" json:"-"`
	Area                  string             `gorm:"size:256;comment:区域" json:"area"`
	StarSign              string             `gorm:"size:64;comment:星座" json:"starSign"`
	Personality           string             `gorm:"size:256;comment:人格" json:"personality"`
	WantsCount            int                `gorm:"size:8;default:0;comment:需求数量" json:"-"`
	OffersCount           int                `gorm:"size:8;default:0;comment:供给数量" json:"-"`
	Member                member.Member      `gorm:"foreignKey:MemberID" json:"detail"`
	Assets                []*Asset           `gorm:"foreignKey:OwnerID" json:"-"`
	Wallets               []*Wallet          `gorm:"foreignKey:OwnerID" json:"-"`
	Privileges            []*MemberPrivilege `gorm:"many2many:rosen_member_member_privileges" json:"-"`
}

func (MemberExtra) TableName() string {
	return "rosen_member_extras"
}

// MemberPosition 会员位置快照
type MemberPosition struct {
	MemberRefer uint           `gorm:"primaryKey;comment:会员ID"`
	Latitude    float64        `gorm:"index;comment:当前位置-纬度"`
	Longitude   float64        `gorm:"index;comment:当前位置-经度"`
	Timestamp   uint64         `gorm:"index;comment:快照时间"`
	Visible     *bool          `gorm:"default:0;comment:是否可见"`
	Extra       *MemberExtra   `gorm:"ForeignKey:MemberRefer"`
	Member      *member.Member `gorm:"ForeignKey:MemberRefer"`
}

func (MemberPosition) TableName() string {
	return "rosen_member_positions"
}

type MemberPrivilege struct {
	common.CrudModel
	Name string `gorm:"size:64;unique_index;comment:名称"`
}

func (MemberPrivilege) TableName() string {
	return "rosen_member_privileges"
}

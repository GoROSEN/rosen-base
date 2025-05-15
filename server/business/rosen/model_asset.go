package rosen

import (
	"github.com/GoROSEN/rosen-apiserver/core/common"
	"gorm.io/gorm"
)

// 资产
type Asset struct {
	gorm.Model

	Name            string  `gorm:"comment:名称;size:64"`
	Kind            string  `gorm:"comment:类型;size:16"`
	Logo            string  `gorm:"comment:图标URL;size:256"`
	Image           string  `gorm:"comment:图片URL;size:256"`
	Description     string  `gorm:"comment:说明;size:1024"`
	Count           uint    `gorm:"comment:数量/耐久度"`
	DueTo           int64   `gorm:"default:0;comment:买断型装备有效截止日期（unix时间戳）"`
	ChainName       string  `gorm:"comment:资产所在主链名称;size:32"`
	ContractAddress string  `gorm:"comment:资产所属合约地址;size:256"`
	NFTAddress      string  `gorm:"comment:资产NFT地址;size:256"`
	TokenId         uint64  `gorm:"index;comment:NFTTokenID;size:64"`
	OwnerID         uint    `gorm:"index;comment:持有人ID"`
	OwnerAddress    string  `gorm:"comment:持有人钱包地址;size:256"`
	EarnRate        float64 `gorm:"comment:Earn速度"`
	Type            int     `gorm:"comment:类型，0-gallery,1-producer装备,2-其他"`
	Level           int     `gorm:"size:16;default:1;comment:资产等级"`
	Transferrable   *bool   `gorm:"default:0;comment:是否可转移"`
	Owner           *MemberExtra
}

func (Asset) TableName() string {
	return "rosen_assets"
}

// Wallet 钱包
type Wallet struct {
	gorm.Model

	OwnerID         uint   `gorm:"index;uniqueIndex:owner_chain_token_address;comment:持有者ID"`
	Chain           string `gorm:"uniqueIndex:owner_chain_token_address;size:128;comment:所属链"`
	Token           string `gorm:"index;uniqueIndex:owner_chain_token_address;size:128;comment:代币类型"`
	ContractAddress string `gorm:"size:256;comment:合约地址"`
	Address         string `gorm:"uniqueIndex:owner_chain_token_address;size:256;comment:地址"`
	PubKey          string `gorm:"index;size:256;comment:公钥"`
	PriKey          string `gorm:"size:1024;comment:私钥，加密"`
	PassPhrase      string `gorm:"size:256;comment:私钥密码"`
	Cipher          string `gorm:"size:1024;comment:私钥加密密钥"`
}

func (Wallet) TableName() string {
	return "rosen_wallets"
}

// WithdrawRequest 提现申请
type WithdrawRequest struct {
	common.CrudModel

	RequesterID   uint         `gorm:"comment:申请人" json:"-"`
	Chain         string       `gorm:"size:128;comment:所属链" json:"chain"`
	Token         string       `gorm:"size:128;comment:代币类型" json:"token"`
	Amount        int64        `gorm:"size:64;comment:金额" json:"amount"`
	DisplayAmount float64      `gorm:"type:decimal(10,3);comment:显示用金额" json:"displayAmount"`
	Status        int          `gorm:"size:8;comment:状态：0-未处理;1-已通过;2-已拒绝" json:"status"`
	Memo          string       `gorm:"size:1024;comment:审核备注" json:"memo"`
	Requester     *MemberExtra `json:"requester"`
	TxHash        string       `gorm:"size:128;comment:处理后的交易哈希" json:"txhash"`
}

func (WithdrawRequest) TableName() string {
	return "rosen_withdraw_requests"
}

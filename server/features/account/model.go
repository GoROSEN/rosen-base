// Package account 资金账户管理模块
package account

import (
	"time"

	"gorm.io/gorm"
)

// Account 账户（资金或积分）
type Account struct {
	gorm.Model
	Type      string `gorm:"comment:币种、类型;size:10"`       // 币种、类型
	Available int64  `gorm:"comment:可用金额"`                // 可用金额
	Locked    int64  `gorm:"comment:锁定金额"`                // 锁定金额
	Frozen    int64  `gorm:"comment:冻结金额"`                // 冻结金额
	UserID    uint   `gorm:"index;comment:所属用户ID"`        // 所属用户ID
	UserName  string `gorm:"index;comment:用户名称;size:256"` // 用户名称
}

// Transaction 交易（账户变动记录）
type Transaction struct {
	gorm.Model
	FromAccountID uint   `gorm:"index;comment:转出账户ID"` // 转出账户ID
	ToAccountID   uint   `gorm:"index;comment:转入账户ID"` // 转入账户ID
	Operation     string `gorm:"comment:操作;size:32"`   // 操作
	Value         int64  `gorm:"comment:金额"`           // 金额
	Description   string `gorm:"comment:说明;size:1024"` // 操作
}

// Receiption 收据
type Receiption struct {
	ID         string `gorm:"unique_index;comment:收据ID"`
	CreatedAt  time.Time
	OwnerID    uint      `gorm:"comment:用户ID"`
	RawContent string    `gorm:"comment:收据内容"`
	Revoke     string    `gorm:"comment:吊销内容"`
	RevokeAt   time.Time `gorm:"type:TIMESTAMP;null;default:null"`
}

// MigrateDB 自动建表
func MigrateDB(db *gorm.DB) {

	db.AutoMigrate(&Account{})
	db.AutoMigrate(&Transaction{})
	db.AutoMigrate(&Receiption{})
}

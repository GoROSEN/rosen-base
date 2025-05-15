package rosen

import "gorm.io/gorm"

// MigrateDB 更新数据库表结构
func MigrateDB(db *gorm.DB) {

	db.AutoMigrate(&Asset{},
		&Wallet{},
		&MemberExtra{},
		&MemberPosition{},
		&MteSession{},
		&MteTrace{},
		&Plot{},
		&ListingPlot{},
		&SysConfig{},
		&MintLog{},
		&PlotCollection{},
		&CityRank{},
		&MemberPrivilege{},
		&WithdrawRequest{},
	)
}

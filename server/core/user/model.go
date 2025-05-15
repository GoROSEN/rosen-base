// Package user 用户信息管理模块
package user

import "gorm.io/gorm"

// User 系统后台用户
type User struct {
	gorm.Model
	// Login
	LoginName string          `gorm:"index;comment:登录名;size:16"` // 登录名
	LoginPwd  string          `gorm:"comment:登录密码（哈希）;size:64"`  // 登录密码哈希
	RoleID    uint            `gorm:"comment:角色ID "`             // 角色
	Role      *RolePermission // 角色
	ApiKey    string          `gorm:"index;size:64"`
	ApiSecret string          `gorm:"size:128"`
	// Profile
	Name         string `gorm:"comment:姓名;size:16"`        // 姓名
	Gender       string `gorm:"comment:性别;size:8"`         // 性别
	Avatar       string `gorm:"comment:头像URL;size:256"`    // 头像URL
	WechatOpenID string `gorm:"comment:微信OpenID;size:256"` // 微信openid
}

// RolePermission 角色权限
type RolePermission struct {
	gorm.Model
	Role       string `gorm:"unique;comment:角色名称;size:32"` // 角色名称
	Permission string `gorm:"comment:逗号分隔的权限名称;size:1024"` // 逗号分隔的权限名称
}

// MigrateDB 创建数据库表
func MigrateDB(db *gorm.DB) {
	db.AutoMigrate(&User{})
	db.AutoMigrate(&RolePermission{})
}

// InitBootstrapUser 若系统中没有任何用户，创建默认超管账号（admin/admin）
func InitBootstrapUser(db *gorm.DB) bool {
	if db.First(&User{}).Error != nil {
		role := &RolePermission{
			Role:       "admin",
			Permission: "*",
		}
		user := &User{
			LoginName:    "admin",
			LoginPwd:     "cad3fc0e677874ae9d85895630ced9735882f06464229200422d5721fa8be743", // adminpass
			Role:         role,
			Name:         "系统管理员",
			Gender:       "保密",
			WechatOpenID: "",
			Avatar:       "",
		}
		return db.Save(user).Error == nil
	}
	return false
}

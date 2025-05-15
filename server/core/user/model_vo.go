// Package user 用户信息管理模块
package user

// LoginUserVO 登录用户VO
type LoginUserVO struct {
	LoginName string `form:"loginName" json:"loginName"` // 登录名
	LoginPwd  string `form:"loginPwd" json:"loginPwd"`   // 登录密码哈希
}

// RoleVO 角色
type RoleVO struct {
	ID          uint     `json:"id"`
	Role        string   `json:"name"`
	Permission  string   `json:"-"`
	Permissions []string `json:"permissions"`
}

// UserVO 系统用户VO
type UserVO struct {
	ID        uint    `form:"id" json:"id"`
	LoginName string  `form:"loginName" json:"loginName"` // 登录名
	Role      *RoleVO `form:"role" json:"role"`           // 角色
	Name      string  `form:"name" json:"name"`           // 姓名
	Gender    string  `form:"gender" json:"gender"`       // 性别
	Avatar    string  `form:"avatar" json:"avatar"`       // 头像URL
}

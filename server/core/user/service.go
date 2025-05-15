package user

import (
	"github.com/GoROSEN/rosen-apiserver/core/common"
	"github.com/google/martian/log"
	"gorm.io/gorm"
)

// Service 用户服务
type Service struct {
	common.CrudService
}

// NewUserService 创建用户服务
func NewUserService(_db *gorm.DB) *Service {
	return &Service{*common.NewCrudService(_db)}
}

// CreateUser 新建用户
func (s *Service) CreateUser(u *User) error {

	return s.Db.Create(u).Error
}

// RegisterUser 注册新用户
func (s *Service) RegisterUser(loginName, password string) *User {

	u := &User{LoginName: loginName, LoginPwd: password}
	err := s.Db.Create(u).Error
	log.Errorf("user.RegisterUser error: %v", err)
	if err != nil {
		return nil
	}
	return u
}

// GetUserByLoginName 根据用户名获取用户
func (s *Service) GetUserByLoginName(name string) *User {

	user := User{}
	err := s.Db.Preload("Role").First(&user, "login_name = ?", name).Error
	if err != nil {
		return nil
	}
	return &user
}

// GetUserByLoginName 根据用户名获取用户
func (s *Service) GetUserByApiKey(apikey string) *User {

	user := User{}
	err := s.Db.Preload("Role").First(&user, "api_key = ?", apikey).Error
	if err != nil {
		return nil
	}
	return &user
}

// UpdateProfile 更新用户档案
func (s *Service) UpdateProfile(user *User) error {

	return s.Db.Model(user).Updates(User{Name: user.Name, Gender: user.Gender, WechatOpenID: user.WechatOpenID, Avatar: user.Avatar, RoleID: user.Role.ID}).Error
}

// UpdateRole 更新用户角色
func (s *Service) UpdateRole(user *User) error {

	return s.Db.Model(user).Updates(User{Role: user.Role}).Error
}

// UpdatePassword 更新登录密码
func (s *Service) UpdatePassword(user *User) error {

	return s.Db.Model(user).Updates(User{LoginPwd: user.LoginPwd}).Error
}

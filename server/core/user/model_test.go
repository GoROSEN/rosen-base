package user

import (
	"testing"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/jinzhu/copier"
)

func TestModelToVo(t *testing.T) {

	m := randomUserModel()
	v := &UserVO{}

	copier.Copy(v, m)

	if m.ID != v.ID {
		t.Errorf("ID is not identical")
	}

	if m.LoginName != v.LoginName {
		t.Errorf("LoginName is not identical")
	}

	if m.Name != v.Name {
		t.Errorf("Name is not identical")
	}

	if m.Gender != v.Gender {
		t.Errorf("Gender is not identical")
	}

	if m.Avatar != v.Avatar {
		t.Errorf("Avatar is not identical")
	}

	if m.RoleID != v.Role.ID || m.Role.ID != v.Role.ID {
		t.Errorf("Role ID is not identical")
	}

	if m.Role.Role != v.Role.Role {
		t.Errorf("Role Name is not identical")
	}

	if m.Role.Permission != v.Role.Permission {
		t.Errorf("Role Permission is not identical")
	}
}

func randomRoleModel() *RolePermission {

	m := &RolePermission{}
	m.ID = uint(randomdata.Number(20000))
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()
	m.Role = randomdata.SillyName()
	m.Permission = randomdata.Noun() + "," + randomdata.Noun()

	return m
}

func randomUserModel() *User {

	m := &User{}

	m.ID = uint(randomdata.Number(10000))
	m.CreatedAt = time.Now()
	m.UpdatedAt = time.Now()
	m.LoginName = randomdata.SillyName()
	m.LoginPwd = randomdata.Letters(10)
	m.Role = randomRoleModel()
	m.RoleID = m.Role.ID
	m.Name = randomdata.LastName()
	m.Gender = []string{"保密", "男", "女"}[randomdata.Number(0, 1, 2)]
	m.Avatar = randomdata.RandStringRunes(18)
	m.WechatOpenID = randomdata.RandStringRunes(24)

	return m
}

// Package auth 用户认证模块
package auth

import (
	"encoding/json"
	"time"
)

// Token 认证令牌
type Token struct {
	UserID                uint      `json:"userId"`                // 用户ID
	UserName              string    `json:"userName"`              // 用户昵称
	UserRole              string    `json:"userRole"`              // 用户角色
	Token                 string    `json:"token"`                 // 认证令牌
	RefreshToken          string    `json:"refreshToken"`          // 刷新令牌
	TokenExpiresAt        time.Time `json:"tokenExpiresAt"`        // 令牌过期时间
	RefreshTokenExpiresAt time.Time `json:"refreshTokenExpiresAt"` // 刷新令牌过期时间
}

func (m *Token) MarshalBinary() ([]byte, error) {
	return json.Marshal(m)
}

func (m *Token) UnmarshalBinary(data []byte) error {
	// convert data to yours, let's assume its json data
	return json.Unmarshal(data, m)
}

// AuthRequestParams 认证请求参数
type AuthRequestParams struct {
	LoginName string `form:"username" json:"username"`
	LoginPwd  string `form:"password" json:"password"`
	VCode     string `form:"vcode" json:"vcode"`
}

type ApikeyParams struct {
	ApiKey    string `json:"apikey"`
	ApiSecret string `json:"secret"`
	UserID    uint   `json:"userId"`
	UserName  string `json:"userName"` // 用户昵称
	UserRole  string `json:"userRole"` // 用户角色
}

func (m *ApikeyParams) MarshalBinary() ([]byte, error) {
	return json.Marshal(m)
}

func (m *ApikeyParams) UnmarshalBinary(data []byte) error {
	// convert data to yours, let's assume its json data
	return json.Unmarshal(data, m)
}

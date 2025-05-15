package member

import (
	"time"

	"github.com/vmihailenco/msgpack"
)

// MemberVO 会员基础视图模型
type MemberVO struct {
	ID          uint   `json:"id"`
	UserName    string `json:"userName"`    // 用户名
	DisplayName string `json:"displayName"` // 昵称
	Avatar      string `json:"avatar"`      // 头像
	Gender      string `json:"gender"`      // 性别
	Bio         string `json:"bio"`
	Level       uint   `json:"level"`
}

// MemberFullVO 会员全视图模型
type MemberFullVO struct {
	ID          uint      `json:"id"`
	UserName    string    `json:"userName"`    // 用户名
	DisplayName string    `json:"displayName"` // 昵称
	CellPhone   string    `json:"cellPhone"`   // 手机号
	Email       string    `json:"email"`       // 电邮
	Gender      string    `json:"gender"`      // 性别
	Avator      string    `json:"avatar"`      // 头像
	IDCard      string    `json:"idcard"`      // 身份证号
	NewPwd      string    `json:"pwd"`         // 新密码
	Bio         string    `json:"bio"`
	Level       uint      `json:"level"`
	CreatedAt   time.Time `json:"createdAt"` // 注册时间
}

func (s *MemberFullVO) MarshalBinary() ([]byte, error) {
	return msgpack.Marshal(s)
}

func (s *MemberFullVO) UnmarshalBinary(data []byte) error {
	return msgpack.Unmarshal(data, s)
}

type SnsIdVO struct {
	ID        uint   `json:"id"`
	SnsType   string `json:"type"`   // 平台类型
	OpenID    string `json:"openId"` // openid
	AvatarURL string `json:"avatar"` // 头像
}

type SnsSummaryVO struct {
	ID              uint `json:"-"`
	MemberID        uint `json:"-"`
	FollowersCount  uint `json:"followers"`
	FollowingsCount uint `json:"following"`
}

type SnsLoginRequestVo struct {
	Platform    string `json:"platform"`
	UserID      string `json:"uid"`
	AccessToken string `json:"accessToken"`
}

// LoginInfoVo 登录用户信息
type LoginInfoVo struct {
	ID          uint   `json:"id"`
	Token       string `json:"token"`
	BindMobile  string `json:"mobile"`   // 绑定手机号
	DisplayName string `json:"nickname"` // 显示名
	AvatarURL   string `json:"portrait"` // 头像
}

// LoginRequestVo 登录请求
type LoginRequestVo struct {
	LoginName string `json:"mobile"`
	LoginPwd  string `json:"password"`
	Captcha   string `json:"captcha"`
	DevId     string `json:"devid"`
	DevType   string `json:"devtype"`
	Lang      string `json:"lang"`
}

// SignUpRequestVo 会员注册请求
type SignUpRequestVo struct {
	UserName  string `json:"userName"`  // 用户名
	CellPhone string `json:"cellPhone"` // 手机号
	Email     string `json:"email"`     // 电邮
	NewPwd    string `json:"pwd"`       // 新密码
	VCode     string `json:"vcode"`     // 验证码
}

type SnsFriendRequestVo struct {
	ID       uint      `json:"id"`
	Message  string    `json:"message"`
	Status   uint      `json:"status"`
	Sender   *MemberVO `json:"sender,omitempty"`
	Receiver *MemberVO `json:"receiver,omitempty"`
}

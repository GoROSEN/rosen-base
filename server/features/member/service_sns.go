package member

import (
	"errors"
	"fmt"

	"github.com/GoROSEN/rosen-apiserver/core/config"
	"github.com/go-redis/redis/v7"
	"github.com/google/martian/log"
	fb "github.com/huandu/facebook/v2"
	"github.com/jinzhu/copier"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

// SnsService 三方平台登录服务
type SnsService struct {
	db     *gorm.DB
	client *redis.Client
	fbApp  *fb.App // facebook app
}

// NewSnsService 根据配置创建三方平台登录服务
func NewSnsService(_db *gorm.DB, rds *redis.Client) *SnsService {
	cfg := config.GetConfig()
	return &SnsService{
		_db,
		rds,
		fb.New(cfg.Sns.Facebook.AppId, cfg.Sns.Facebook.AppSecret),
	}
}

// SnsSignin 使用三方平台信息登录关联账户，若关联账户不存在，则创建一个
func (s *SnsService) SnsSignin(platform, userId, accessToken string) (string, *Member, bool, error) {

	var m Member
	var snsid SnsID
	var isnew bool

	if platform == "facebook" {
		if err := s.fillFacebookUserInfo(userId, accessToken, &m); err != nil {
			return "", nil, isnew, err
		}
		// userId = m.UserName
		m.UserName = ""
	} else if platform == "twitter" {
		if err := s.fillTwitterUserInfo(userId, accessToken, &m); err != nil {
			return "", nil, isnew, err
		}
		userId = m.UserName
		m.UserName = ""
	} else if platform == "google" {
		if err := s.fillGoogleUserInfo(userId, accessToken, &m); err != nil {
			return "", nil, isnew, err
		}
		userId = m.UserName
		m.UserName = ""
	} else if platform == "apple" {
		if err := s.fillAppleUserInfo(userId, accessToken, &m); err != nil {
			return "", nil, isnew, err
		}
		userId = m.UserName
		m.UserName = ""
	} else if platform == "apple-web" {
		if err := s.fillAppleWebUserInfo(userId, accessToken, &m); err != nil {
			return "", nil, isnew, err
		}
		userId = m.UserName
		m.UserName = ""
		platform = "apple"
	}

	if err := s.db.Where("sns_type = ? and open_id = ?", platform, userId).First(&snsid).Error; err != nil {

		// 创建关联用户
		m.UserName = uuid.NewV4().String()
		isnew = true
		// 检查邮箱是否被占用
		if err := s.db.First(&Member{}, "user_name = ? or email = ?", m.UserName, m.Email); err == nil {
			// 用户名已存在
			log.Errorf("sns user email %v already exist", m.Email)
			return "", nil, isnew, errors.New("message.member.email-registered")
		}

		t := s.db.Begin()
		if err := t.Save(&m).Error; err != nil {
			t.Rollback()
			return "", nil, isnew, err
		}
		// 创建新snsid
		snsid.MemberID = m.ID
		snsid.SnsType = platform
		snsid.OpenID = userId
		snsid.AccessToken = accessToken
		snsid.AvatarURL = m.Avatar
		if err := t.Save(&snsid).Error; err != nil {
			t.Rollback()
			return "", nil, isnew, err
		}
		t.Commit()
	} else {
		if err := s.db.First(&m, snsid.MemberID).Error; err != nil {
			return "", nil, isnew, err
		}
		if len(m.LoginPwd) > 0 {
			// 非关联账户
			return "", nil, isnew, errors.New("please login with password")
		}
		// update
		if err := s.db.Model(&snsid).Update("access_token", accessToken).Error; err != nil {
			return "", nil, isnew, err
		}
	}

	// do login
	var mvo MemberFullVO
	copier.Copy(&mvo, &m)

	// 创建token并存到redis
	token := uuid.NewV4().String()
	tokenLife := config.GetConfig().Token.TokenLife
	err := s.client.Set(token, &mvo, tokenLife).Err()

	return token, &m, isnew, err
}

func (s *SnsService) fillFacebookUserInfo(userId, accessToken string, m *Member) error {

	// verify access token
	res, err := fb.Get(fmt.Sprintf("/%v", userId), fb.Params{
		"fields":       "name,email,gender",
		"access_token": accessToken,
	})
	if err != nil {
		return err
	}
	log.Infof("res = %v", res) // 2022/11/03 02:23:51 INFO: res = map[__usage__:0xc0000ba2c0 email:jupiter@gorosen.xyz id:122809760604376 name:YU Jupiter]
	if _, ok := res["email"]; !ok {
		return errors.New("email is required")
	}
	if _, ok := res["name"]; !ok {
		return errors.New("name is required")
	}
	m.Email = res["email"].(string)
	m.DisplayName = res["name"].(string)
	if _, ok := res["gender"]; ok {
		m.Gender = res["gender"].(string)
	}
	if _, ok := res["id"]; ok {
		userId = res["id"].(string)
	}
	m.Avatar = fmt.Sprintf("https://api.gorosen.xyz/api/open/member/ext/sns-avatar/fb?uid=%v", userId)
	m.UserName = userId
	return nil
}

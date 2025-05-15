package auth

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/go-redis/redis/v7"
	"github.com/google/martian/log"
	"gorm.io/gorm"

	"github.com/GoROSEN/rosen-apiserver/core/config"
	"github.com/GoROSEN/rosen-apiserver/core/user"
	"github.com/GoROSEN/rosen-apiserver/core/utils"
	"github.com/GoROSEN/rosen-apiserver/features/member"
)

// Service 认证服务
type Service struct {
	client    *redis.Client
	userSrv   *user.Service
	memberSrv *member.Service
}

// NewAuthService 创建新认证服务
func NewAuthService(redisClient *redis.Client, db *gorm.DB) *Service {

	return &Service{client: redisClient, userSrv: user.NewUserService(db), memberSrv: member.NewService(db, redisClient)}
}

func newToken(user *user.User) *Token {

	tokenLife := config.GetConfig().Token.TokenLife               //time.ParserDuration("1h")
	refreshTokenLife := config.GetConfig().Token.RefreshTokenLife //time.ParserDuration("24h")

	token := &Token{UserID: user.ID, UserName: user.Name, UserRole: user.Role.Permission}
	token.Token = fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("TOKEN%v%v%v", user.ID, randomdata.Number(0, 99999999), time.Now()))))
	token.RefreshToken = fmt.Sprintf("%x", sha256.Sum256([]byte(fmt.Sprintf("REFRESHTOKEN%v%v%v", user.ID, randomdata.Number(0, 99999999), time.Now()))))
	token.TokenExpiresAt = time.Now().Add(tokenLife)
	token.RefreshTokenExpiresAt = time.Now().Add(refreshTokenLife)
	return token
}

// AuthUser 认证用户
func (s *Service) AuthUser(userName, password string) (*Token, error) {

	log.Infof("auth user for '%v'", userName)
	// log.Debugf("password = %v", password)
	user := s.userSrv.GetUserByLoginName(userName)
	passSha256 := utils.GetPass(password)

	if user == nil {
		log.Errorf("user '%v' not found", userName)
		return nil, errors.New("invalid username or password")
	} else if user.LoginPwd == passSha256 {
		// create token
		token := newToken(user)
		// save token
		key := fmt.Sprintf("LOGIN_TOKEN:%v", token.Token)
		err := s.client.Set(key, token, config.GetConfig().Token.RefreshTokenLife).Err()
		if err != nil {
			log.Errorf("创建令牌失败，请检查缓存服务是否正常")
		}
		return token, err
	} else {
		// log.Debugf("user.pass = %v , request pwd = %v", user.LoginPwd, passSha256)
		log.Errorf("invalid password for user '%v'", userName)
		return nil, errors.New("invalid username or password")
	}
}

// VerifyUserToken 校验用户令牌，返回用户ID、昵称和角色
func (s *Service) VerifyUserToken(token string) (*Token, error) {

	t, err := s.getUserToken(token)
	if err != nil {
		return nil, err
	}
	if t.Token != token {
		return nil, errors.New("invalid token")
	}
	if time.Now().After(t.TokenExpiresAt) {
		s.RemoveToken(token)
		return nil, errors.New("token expired")
	}
	// TODO: update token life
	return t, nil
}

// RefreshUserToken 更新用户令牌
func (s *Service) RefreshUserToken(token string, refreshToken string) (*Token, error) {

	t, err := s.getUserToken(token)
	if err != nil {
		return nil, err
	}
	if t.RefreshToken != refreshToken {
		return nil, errors.New("invalid refresh token")
	}
	if time.Now().After(t.RefreshTokenExpiresAt) {
		return nil, errors.New("refresh token expired")
	}
	var u user.User
	if err := s.userSrv.GetModelByID(&u, t.UserID); err != nil {
		return nil, err
	}
	t = newToken(&u)
	key := fmt.Sprintf("LOGIN_TOKEN:%v", t.Token)
	err = s.client.Set(key, t, config.GetConfig().Token.RefreshTokenLife).Err()

	return t, err
}

// RemoveToken 删除现有token
func (s *Service) RemoveToken(token string) {

	key := fmt.Sprintf("LOGIN_TOKEN:%v", token)
	s.client.Del(key)
}

// VerifyMemberToken 校验会员令牌
func (s *Service) VerifyMemberToken(token string) (uint, error) {

	m, err := s.memberSrv.MemberByToken(token)
	if err != nil {
		return 0, err
	}
	return m.ID, nil
}

// getUserToken 获取指定Token
func (s *Service) getUserToken(token string) (*Token, error) {

	t := Token{}
	key := fmt.Sprintf("LOGIN_TOKEN:%v", token)
	if err := s.client.Get(key).Scan(&t); err != nil {
		return nil, err
	}
	return &t, nil
}

func (s *Service) verifyApiKey(apikey string) (*ApikeyParams, error) {

	var t ApikeyParams
	key := fmt.Sprintf("APIKEY:%v", apikey)
	if err := s.client.Get(key).Scan(&t); err != nil {

		// query from db
		u := s.userSrv.GetUserByApiKey(apikey)
		if u == nil {
			log.Errorf("cannot find user with apikey = %v", apikey)
			return nil, errors.New("cannot find user")
		}
		t.UserID = u.ID
		t.ApiKey = u.ApiKey
		t.ApiSecret = u.ApiSecret
		t.UserName = u.Name
		t.UserRole = u.Role.Permission
		s.client.Set(key, &t, config.GetConfig().Token.RefreshTokenLife)
		return &t, nil
	}
	return &t, nil
}

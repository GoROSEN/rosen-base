package member

import (
	"errors"
	"fmt"

	"github.com/GoROSEN/rosen-apiserver/core/common"
	"github.com/GoROSEN/rosen-apiserver/core/config"
	"github.com/GoROSEN/rosen-apiserver/core/utils"
	"github.com/go-redis/redis/v7"
	"github.com/google/martian/log"
	"github.com/jinzhu/copier"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

// Service 服务层
type Service struct {
	common.CrudService
	SnsService

	client *redis.Client
}

// NewService 新建服务
func NewService(_db *gorm.DB, rds *redis.Client) *Service {
	return &Service{*common.NewCrudService(_db), *NewSnsService(_db, rds), rds}
}

// SignUp 会员注册
func (s *Service) SignUpMember(obj *Member) error {

	db := s.Db.Begin()
	if err := db.Save(obj).Error; err != nil {
		log.Errorf("save member error: %v", err)
		db.Rollback()
		return err
	}
	log.Infof("saved member %v, got member-id %v", obj.UserName, obj.ID)
	sns := &SnsSummary{MemberID: obj.ID}
	if err := db.Save(sns).Error; err != nil {
		log.Errorf("save sns error: %v", err)
		db.Rollback()
		return err
	}
	log.Infof("saved member sns summary")
	db.Commit()
	return nil
}

// AuthMember 会员鉴权
func (s *Service) AuthMember(loginName, loginPwd string) (string, *Member, error) {

	if len(loginName) == 0 {
		return "", nil, errors.New("invalid username")
	}
	// 密码加密规则（同user）：前端将密码哈希为md5后进行传输，入参loginPwd为md5(原始密码)；后端再进行sha256后进行比对
	passSha256 := utils.GetPass(loginPwd)

	var m Member
	if err := s.FindModelWhere(&m, "user_name = ? or cell_phone = ?", loginName, loginName); err != nil {
		return "", nil, err
	}
	// 验证登录名密码，成功后将token与会员ID存入redis
	if len(m.LoginPwd) == 0 {
		return "", nil, errors.New("please login with 3rd party")
	}
	if m.LoginPwd != passSha256 {
		return "", nil, errors.New("invalid password")
	}

	var mvo MemberFullVO
	copier.Copy(&mvo, &m)

	// 创建token并存到redis
	token := uuid.NewV4().String()
	tokenLife := config.GetConfig().Token.TokenLife
	err := s.client.Set(token, &mvo, tokenLife).Err()

	return token, &m, err
}

// AuthMemberBySMS 会员鉴权（短信验证码）
func (s *Service) AuthMemberBySMS(loginName, loginPwd string) (string, *Member, error) {

	var m Member
	if err := s.FindModelWhere(&m, "user_name = ? or cell_phone = ?", loginName, loginName); err != nil {
		return "", nil, err
	}
	// 验证登录名密码，成功后将token与会员ID存入redis
	// 暂时使用万能码
	if loginPwd != "999999" {
		return "", nil, errors.New("invalid password")
	}

	var mvo MemberFullVO
	copier.Copy(&mvo, &m)

	// 创建token并存到redis
	token := uuid.NewV4().String()
	tokenLife := config.GetConfig().Token.TokenLife
	err := s.client.Set(token, &mvo, tokenLife).Err()
	log.Infof("AuthMemberBySMS: created token: %v, life: %v, member-id: %v, err: %v", token, tokenLife, m.ID, err)

	return token, &m, err
}

// MemberByToken 获取登录令牌匹配的会员信息
func (s *Service) MemberByToken(token string) (*Member, error) {

	var mvo MemberFullVO
	if err := s.client.Get(token).Scan(&mvo); err != nil {
		return nil, err
	}

	log.Infof("MemberByToken: got member-id: %v", mvo.ID)

	var m Member
	if err := s.GetPreloadModelByID(&m, mvo.ID, []string{"SnsIDs"}); err != nil {
		return nil, err
	}

	return &m, nil
}

func (s *Service) MemberVoByToken(token string) (*MemberFullVO, error) {

	var mvo MemberFullVO
	if err := s.client.Get(token).Scan(&mvo); err != nil {
		return nil, err
	}

	return &mvo, nil
}

func (s *Service) MemberFollow(memberID uint, followingID uint) error {

	log.Infof("MemberFollow: %v -> %v", memberID, followingID)
	var member, m Member
	member.ID = memberID
	m.ID = followingID

	//检查是否已关注
	if s.Db.Model(&member).Where("following_id = ?", followingID).Association("Followings").Count() > 0 {
		log.Infof("already followed")
		return nil
	}

	member = Member{}
	member.ID = memberID
	// update follower
	tdb := s.Db.Begin()
	if err := tdb.Model(&member).Association("Followings").Append(&m); err != nil {
		tdb.Rollback()
		return err
	}

	// update following
	member = Member{}
	member.ID = memberID
	if err := tdb.Model(&m).Association("Followers").Append(&member); err != nil {
		tdb.Rollback()
		return err
	}
	if err := tdb.Commit().Error; err != nil {
		return err
	}

	// update summary
	var sns SnsSummary
	if err := s.Db.Find(&sns, "member_id = ?", memberID).Error; err != nil {
		return err
	}

	member = Member{}
	member.ID = memberID
	sns.FollowingsCount = uint(s.Db.Model(&member).Association("Followings").Count())
	if err := s.Db.Save(&sns).Error; err != nil {
		return err
	}

	var sns2 SnsSummary
	if err := s.Db.Find(&sns2, "member_id = ?", followingID).Error; err != nil {
		return err
	}

	sns2.FollowersCount = uint(s.Db.Model(&m).Association("Followers").Count())
	if err := s.Db.Save(&sns2).Error; err != nil {
		return err
	}

	return nil
}

func (s *Service) MemberUnfollow(memberID uint, followingID uint) error {

	log.Infof("MemberUnfollow: %v -x-> %v", memberID, followingID)
	var member, m Member
	member.ID = memberID
	m.ID = followingID

	//检查是否已关注
	if s.Db.Model(&member).Where("following_id = ?", followingID).Association("Followings").Count() == 0 {
		log.Infof("already unfollowed")
		return nil
	}

	// update follower
	member = Member{}
	member.ID = memberID
	tdb := s.Db.Begin()
	if err := tdb.Model(&member).Association("Followings").Delete(&m); err != nil {
		tdb.Rollback()
		return err
	}

	// update following
	member = Member{}
	member.ID = memberID
	if err := tdb.Model(&m).Association("Followers").Delete(&member); err != nil {
		tdb.Rollback()
		return err
	}
	if err := tdb.Commit().Error; err != nil {
		return err
	}

	// update summary
	var sns SnsSummary
	if err := s.Db.Find(&sns, "member_id = ?", memberID).Error; err != nil {
		return err
	}

	member = Member{}
	member.ID = memberID
	sns.FollowingsCount = uint(s.Db.Model(&member).Association("Followings").Count())
	if err := s.Db.Save(&sns).Error; err != nil {
		return err
	}

	var sns2 SnsSummary
	if err := s.Db.Find(&sns2, "member_id = ?", followingID).Error; err != nil {
		return err
	}

	sns2.FollowersCount = uint(s.Db.Model(&m).Association("Followers").Count())
	if err := s.Db.Save(&sns2).Error; err != nil {
		return err
	}

	return nil
}

func (s *Service) MemberBlock(memberID uint, blockId uint) error {

	log.Infof("MemberBlock: %v -> %v", memberID, blockId)
	var member, m Member
	member.ID = memberID
	m.ID = blockId

	//检查是否已拉黑
	if s.Db.Model(&member).Where("blocked_id = ?", blockId).Association("Blockeds").Count() > 0 {
		log.Infof("already blocked")
		return nil
	}

	member = Member{}
	member.ID = memberID
	// update blocked
	if err := s.Db.Model(&member).Association("Blockeds").Append(&m); err != nil {
		return err
	}

	return nil
}

func (s *Service) MemberUnblock(memberID uint, blockId uint) error {

	log.Infof("MemberUnblock: %v -x-> %v", memberID, blockId)
	var member, m Member
	member.ID = memberID
	m.ID = blockId

	//检查是否已关注
	if s.Db.Model(&member).Where("blocked_id = ?", blockId).Association("Blockeds").Count() == 0 {
		log.Infof("already unblocked")
		return nil
	}

	// update follower
	member = Member{}
	member.ID = memberID
	if err := s.Db.Model(&member).Association("Blockeds").Delete(&m); err != nil {
		return err
	}

	return nil
}

func (s *Service) CheckVCode(vtype, destination, vcode string) (bool, error) {
	redisKey := fmt.Sprintf("%v-code:%v", vtype, destination)
	cmd := s.client.Get(redisKey)
	if cmd.Err() != nil {
		log.Errorf("cannot get vcode from cache: %v")
		return false, cmd.Err()
	}
	if cmd.Val() != vcode {
		log.Errorf("invalid vcode: %v", vcode)
		return false, nil
	}
	if err := s.client.Del(redisKey).Err(); err != nil {
		log.Errorf("cannot delete vcode: %v", cmd.Val())
		return true, err
	}
	return true, nil
}

func (s *Service) ResetMemberPassword(email, phone, passwd string) error {
	var m Member
	if len(phone) > 0 {
		if err := s.db.Model(&Member{}).First(&m, "cell_phone = ?", phone).Error; err != nil {
			return err
		}
	} else if len(email) > 0 {
		if err := s.db.Model(&Member{}).First(&m, "email = ?", email).Error; err != nil {
			return err
		}
	} else {
		return errors.New("invalid account")
	}
	if len(m.LoginPwd) == 0 {
		// 三方授权用户，不能修改密码
		return errors.New("3rd-party user cannot modify password")
	}
	if err := s.db.Model(&m).Update("login_pwd", utils.GetPass(passwd)).Error; err != nil {
		return err
	}
	return nil
}

func (s *Service) MemberIsFollowing(memberID uint, followingID uint) bool {

	log.Infof("MemberIsFollowing: %v -> %v", memberID, followingID)
	var member Member
	member.ID = memberID

	//检查是否已关注
	if s.Db.Model(&member).Where("following_id = ?", followingID).Association("Followings").Count() > 0 {
		log.Infof("already followed")
		return true
	}

	return false
}

func (s *Service) MemberIsFriend(memberID uint, friendID uint) bool {

	log.Infof("MemberIsFriend: %v -> %v", memberID, friendID)
	var member Member
	member.ID = memberID

	//检查是否已好友
	if s.Db.Model(&member).Where("friend_id = ?", friendID).Association("Friends").Count() > 0 {
		log.Infof("already friends")
		return true
	}

	return false
}

func (s *Service) DeleteFriend(memberID uint, friendID uint) error {

	log.Infof("DeleteFriend: %v <-> %v", memberID, friendID)
	var member, m Member
	member.ID = memberID
	m.ID = friendID

	// 双方好友列表中均加入对方
	tdb := s.Db.Begin()
	if err := tdb.Model(&member).Association("Friends").Delete(&m); err != nil {
		tdb.Rollback()
		return err
	}

	if err := tdb.Model(&m).Association("Friends").Delete(&member); err != nil {
		tdb.Rollback()
		return err
	}
	if err := tdb.Commit().Error; err != nil {
		return err
	}

	return nil
}

func (s *Service) ProcessFriendRequest(request *SnsFriendRequest, approved bool) error {

	if approved {
		log.Infof("MemberFriend: %v <-> %v", request.SenderID, request.ReceiverID)
		var member, m Member
		member.ID = request.SenderID
		m.ID = request.ReceiverID

		//检查是否已好友
		if s.Db.Model(&member).Where("friend_id = ?", request.ReceiverID).Association("Friends").Count() > 0 {
			log.Infof("already friends")
		} else {
			// 双方好友列表中均加入对方
			tdb := s.Db.Begin()
			if err := tdb.Model(&member).Association("Friends").Append(&m); err != nil {
				tdb.Rollback()
				return err
			}

			if err := tdb.Model(&m).Association("Friends").Append(&member); err != nil {
				tdb.Rollback()
				return err
			}
			if err := tdb.Commit().Error; err != nil {
				return err
			}
		}
		request.Status = 1
	} else {
		request.Status = 2
	}

	if err := s.UpdateModel(&request, []string{"status"}, nil); err != nil {
		log.Errorf("ProcessFriendRequest: cannot update request: %v", err)
		return err
	}

	if request.Status == 1 {
		// 如果有未处理的另一方好友请求，一并处理掉
		if err := s.Db.Model(&SnsFriendRequest{}).Where("sender_id = ? and receiver_id = ? and status = 0", request.ReceiverID, request.SenderID).Update("status", 1).Error; err != nil {
			log.Errorf("ProcessFriendRequest: cannot update other requests: %v", err)
		}
	}

	return nil
}

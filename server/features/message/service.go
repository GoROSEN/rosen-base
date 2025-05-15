package message

import (
	"github.com/go-redis/redis/v7"
	"gorm.io/gorm"
)

type Service struct {
	db  *gorm.DB
	rds *redis.Client
}

func NewService(_db *gorm.DB, _rds *redis.Client) *Service {
	return &Service{_db, _rds}
}

// SendMemberSysMessage 发送系统消息
func (s *Service) SendMemberSysMessage(memberId uint, title, content string) error {

	msg := &Message{
		DestMemberID: memberId,
		Title:        title,
		Conent:       content,
		Unread:       true,
	}
	if err := s.db.Create(&msg).Error; err != nil {
		return err
	}
	return nil
}

func (s *Service) BindMemberDestination(memberId uint, channel, address string) error {

	// unbind device
	s.db.Delete(&Destination{}, "channel = ? and (member_id = ? or address = ?)", channel, memberId, address)
	// bind to new use
	d := Destination{
		MemberID: memberId,
		Channel:  channel,
		Address:  address,
	}
	return s.db.Create(&d).Error
}

func (s *Service) UnbindMemberDestination(memberId uint, channel, address string) error {

	return s.db.Unscoped().Delete(&Destination{}, "member_id = ? and channel = ? and address = ?", memberId, channel, address).Error
}

func (s *Service) GetDestinationForUser(userId uint, channel string) (*Destination, error) {

	var d Destination
	if err := s.db.First(&d, "user_id = ? and channel = ?", userId, channel).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (s *Service) GetDestinationForMember(memberId uint, channel string) (*Destination, error) {

	var d Destination
	if err := s.db.First(&d, "member_id = ? and channel = ?", memberId, channel).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (s *Service) GetDestinationsForUser(userId uint) ([]Destination, error) {

	d := []Destination{}
	if err := s.db.Find(&d, "user_id = ?", userId).Error; err != nil {
		return nil, err
	}
	return d, nil
}

func (s *Service) GetDestinationsForMember(memberId uint) ([]Destination, error) {

	d := []Destination{}
	if err := s.db.Find(&d, "member_id = ?", memberId).Error; err != nil {
		return nil, err
	}
	return d, nil
}

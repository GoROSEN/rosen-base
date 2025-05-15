package attachment

import (
	"fmt"

	"github.com/GoROSEN/rosen-apiserver/core/common"
	uuid "github.com/satori/go.uuid"
	"gorm.io/gorm"
)

type Service struct {
	common.CrudService
}

func NewService(_db *gorm.DB) *Service {

	return &Service{CrudService: *common.NewCrudService(_db)}
}

func (s *Service) NewPublicAttachment(moduleName, ossFilePathName, fileName, uuidStr string) (*Attachment, error) {

	var att Attachment
	var err error

	if len(uuidStr) > 0 {
		att.UUID = uuidStr
	} else {
		att.UUID = uuid.NewV4().String()
	}
	att.ModuleName = moduleName
	att.IsPublic = true
	att.OssFilePathName = ossFilePathName
	att.FileName = fileName

	if err := s.CreateModel(&att); err != nil {
		return nil, err
	}

	return &att, err
}

func (s *Service) NewMemberPrivateFile(memberId uint, moduleName, ossFilePathName, fileName, uuidStr string) (*Attachment, string, error) {

	var att Attachment
	var err error

	if len(uuidStr) > 0 {
		att.UUID = uuidStr
	} else {
		att.UUID = uuid.NewV4().String()
	}
	att.ModuleName = moduleName
	att.IsPublic = false
	att.OssFilePathName = ossFilePathName
	att.FileName = fileName

	var ownership MemberAttachment
	ownership.AttachmentID = att.ID
	ownership.Attachment = &att
	ownership.MemberID = uint(memberId)

	if err := s.CreateModel(&ownership); err != nil {
		return nil, "", err
	}

	preSignedUrl := fmt.Sprintf("/api/member/attachment/get/%v", att.UUID)
	return &att, preSignedUrl, err
}

func (s *Service) NewUserPrivateFile(userId uint, moduleName, ossFilePathName, fileName, uuidStr string) (*Attachment, string, error) {

	var att Attachment
	var err error

	if len(uuidStr) > 0 {
		att.UUID = uuidStr
	} else {
		att.UUID = uuid.NewV4().String()
	}
	att.ModuleName = moduleName
	att.IsPublic = false
	att.OssFilePathName = ossFilePathName
	att.FileName = fileName

	var ownership UserAttachment
	ownership.AttachmentID = att.ID
	ownership.Attachment = &att
	ownership.UserID = uint(userId)

	if err := s.CreateModel(&ownership); err != nil {
		return nil, "", err
	}

	preSignedUrl := fmt.Sprintf("/api/user/attachment/get/%v", att.UUID)
	return &att, preSignedUrl, err
}

func (s *Service) GetMemberPrivateFile(memberId uint, fileUUID string) (*MemberAttachment, error) {
	var memberAttachment MemberAttachment
	if err := s.FindPreloadJoinModelWhere(&memberAttachment, []string{"Attachment"}, []string{"left join attachment_attachments on attachment_attachments.id = attachment_member_attachments.attachment_id"}, "member_id = ? and attachment_attachments.uuid = ?", memberId, fileUUID); err != nil {
		return nil, err
	}

	return &memberAttachment, nil
}

func (s *Service) GetUserPrivateFile(userId uint, fileUUID string) (*UserAttachment, error) {
	var userAttach UserAttachment
	if err := s.FindPreloadJoinModelWhere(&userAttach, []string{"Attachment"}, []string{"left join attachment_attachments on attachment_attachments.id = attachment_user_attachments.attachment_id"}, "user_id = ? and attachment_attachments.uuid = ?", userId, fileUUID); err != nil {
		return nil, err
	}

	return &userAttach, nil
}

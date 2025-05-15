package message

import (
	"github.com/GoROSEN/rosen-apiserver/core/notification"
	"github.com/go-redis/redis/v7"
	"github.com/google/martian/log"
	"gorm.io/gorm"
)

type MsgMod struct {
	notifyService  *notification.Service
	messageService *Service
}

func (m *MsgMod) SetupMsgMod(db *gorm.DB, rds *redis.Client) {
	m.notifyService = notification.NewService(db)
	m.messageService = NewService(db, rds)
}

func (c *MsgMod) SendSimpleNotification(memberId uint, channel string, content string) {
	if dests, err := c.messageService.GetDestinationsForMember(memberId); err == nil {
		msg := &notification.MessageLog{
			Title: "",
			Body:  content,
		}
		for _, dest := range dests {
			if dest.Channel == channel {
				if err := c.notifyService.SendNotificationMsgWithoutLog(dest.Address, dest.Channel, msg, map[string]string{}); err != nil {
					log.Errorf("cannot send notification: %v", err)
				}
			}
		}
	} else {
		log.Errorf("cannot get message destination: %v", err)
	}
}

func (c *MsgMod) SendSimpleNotificationWithData(memberId uint, channel string, content string, data map[string]string, badge int) {
	if dests, err := c.messageService.GetDestinationsForMember(memberId); err == nil {
		msg := &notification.MessageLog{
			Title: "",
			Body:  content,
		}
		for _, dest := range dests {
			if dest.Channel == channel {
				if err := c.notifyService.SendNotificationMsgWithoutLogV2(dest.Address, dest.Channel, msg, data, badge); err != nil {
					log.Errorf("cannot send notification: %v", err)
				}
			}
		}
	} else {
		log.Errorf("cannot get message destination: %v", err)
	}
}

func (c *MsgMod) SendMessageV2(memberId uint, templateName, language string, params interface{}, sysMsg, pushNotification bool) {
	if dests, err := c.messageService.GetDestinationsForMember(memberId); err == nil {
		msg, err := c.notifyService.GenerateNotificationWithTemplate(templateName, language, params)
		if err != nil {
			log.Errorf("cannot generate msg: %v", err)
			return
		}
		if sysMsg {
			if err := c.messageService.SendMemberSysMessage(memberId, msg.Title, msg.Body); err != nil {
				log.Errorf("cannot send message: %v", err)
			}
		}
		if pushNotification {
			for _, dest := range dests {
				if err := c.notifyService.SendNotificationMsg(dest.Address, dest.Channel, msg, map[string]string{}); err != nil {
					log.Errorf("cannot send notification: %v", err)
				}
			}
		}
	} else {
		log.Errorf("cannot get message destination: %v", err)
	}
}

func (c *MsgMod) SendMessageWithDataV2(memberId uint, templateName, language string, params interface{}, data map[string]string, sysMsg, pushNotification bool) {
	if dests, err := c.messageService.GetDestinationsForMember(memberId); err == nil {
		msg, err := c.notifyService.GenerateNotificationWithTemplate(templateName, language, params)
		if err != nil {
			log.Errorf("cannot generate msg: %v", err)
			return
		}
		if sysMsg {
			if err := c.messageService.SendMemberSysMessage(memberId, msg.Title, msg.Body); err != nil {
				log.Errorf("cannot send message: %v", err)
			}
		}
		if pushNotification {
			for _, dest := range dests {
				if err := c.notifyService.SendNotificationMsg(dest.Address, dest.Channel, msg, data); err != nil {
					log.Errorf("cannot send notification: %v", err)
				}
			}
		}
	} else {
		log.Errorf("cannot get message destination: %v", err)
	}
}

func (c *MsgMod) SendMessage(memberId uint, templateName, language string, params interface{}) {

	c.SendMessageV2(memberId, templateName, language, params, true, true)
}

func (c *MsgMod) SendMessageWithData(memberId uint, templateName, language string, params interface{}, data map[string]string) {
	c.SendMessageWithDataV2(memberId, templateName, language, params, data, true, true)
}

func (c *MsgMod) SendSysMessage(memberId uint, templateName, language string, params interface{}) {

	msg, err := c.notifyService.GenerateNotificationWithTemplate(templateName, language, params)
	if err != nil {
		log.Errorf("cannot create message: %v", err)
	}
	if msg != nil {
		if err := c.messageService.SendMemberSysMessage(memberId, msg.Title, msg.Body); err != nil {
			log.Errorf("cannot send message: %v", err)
		}
	}
}

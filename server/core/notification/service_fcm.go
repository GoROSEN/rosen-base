package notification

import (
	"errors"

	"github.com/appleboy/go-fcm"
	"github.com/google/martian/log"
)

func (s *Service) sendFcm(udid, subject, content string, data map[string]string) error {

	if len(s.fcmkey) == 0 {
		return errors.New("invalid fcm config")
	}
	client, err := fcm.NewClient(s.fcmkey)
	if err != nil {
		return err
	}
	msg := &fcm.Message{
		To:   udid,
		Data: map[string]interface{}{},
		Notification: &fcm.Notification{
			Title: subject,
			Body:  content,
		},
	}
	res, err := client.Send(msg)
	if err != nil {
		return err
	}
	log.Infof("fcm replied: %v", res)
	return nil
}

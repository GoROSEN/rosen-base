package notification

import (
	"errors"

	"github.com/GoROSEN/rosen-apiserver/core/config"
	expo "github.com/oliveroneill/exponent-server-sdk-golang/sdk"
)

func (s *Service) sendExpo(udid, subject, content string, data map[string]string) error {

	return s.broadcastExpos([]string{udid}, subject, content, data, 0)
}

func (s *Service) sendExpoWithBadge(udid, subject, content string, data map[string]string, badge int) error {

	return s.broadcastExpos([]string{udid}, subject, content, data, badge)
}

func (s *Service) broadcastExpos(udids []string, subject, content string, data map[string]string, badge int) error {

	pushTokenes := make([]expo.ExponentPushToken, 0, len(udids))
	for _, udid := range udids {
		pushToken, err := expo.NewExponentPushToken(udid)
		if err == nil {
			pushTokenes = append(pushTokenes, pushToken)
		}
	}
	if len(pushTokenes) == 0 {
		return errors.New("invalid udids")
	}
	cfg := &config.GetConfig().Notification.Expo
	var client *expo.PushClient
	if len(cfg.AccessToken) > 0 {
		client = expo.NewPushClient(&expo.ClientConfig{
			AccessToken: cfg.AccessToken,
		})
	} else {
		client = expo.NewPushClient(nil)
	}
	if response, err := client.Publish(
		&expo.PushMessage{
			To:       pushTokenes,
			Body:     content,
			Data:     data,
			Sound:    "default",
			Title:    subject,
			Priority: expo.DefaultPriority,
			Badge:    badge,
		},
	); err != nil {
		return err
	} else {
		// Validate responses
		if err := response.ValidateResponse(); err != nil {
			return err
		}
	}
	return nil
}

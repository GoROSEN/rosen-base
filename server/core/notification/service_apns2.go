package notification

import (
	"errors"
	"fmt"

	"github.com/google/martian/log"
	"github.com/sideshow/apns2"
)

func (s *Service) sendApns(udid, subject, content string, data map[string]string) error {

	var client *apns2.Client
	if s.apns2.Cert != nil {
		// send with p12
		if s.production {
			client = apns2.NewClient(*s.apns2.Cert).Production()
		} else {
			client = apns2.NewClient(*s.apns2.Cert).Development()
		}
	} else if s.apns2.Token.AuthKey != nil {
		// send with jwt
		if s.production {
			client = apns2.NewTokenClient(s.apns2.Token).Production()
		} else {
			client = apns2.NewTokenClient(s.apns2.Token).Development()
		}
	} else {
		return errors.New("invalid apns2 config")
	}

	msg := &apns2.Notification{
		DeviceToken: udid,
		Topic:       s.apns2.BundleId,
		Payload: []byte(fmt.Sprintf(`{
			"aps" : {
					"alert" : {
							"title" : "%v",
							"body" : "%v",
					}
			}}`, subject, content)),
	}

	if res, err := client.Push(msg); err != nil {
		return err
	} else {
		log.Infof("apns reply: %v %v %v", res.StatusCode, res.ApnsID, res.Reason)
		if res.StatusCode != 200 {
			return errors.New(res.Reason)
		}
	}

	return nil
}

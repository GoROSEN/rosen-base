package notification

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"strings"
	"text/template"
	"time"

	"github.com/GoROSEN/rosen-apiserver/core/config"
	"github.com/google/martian/log"
	"github.com/sideshow/apns2/certificate"
	"github.com/sideshow/apns2/token"
	"gorm.io/gorm"
)

type Service struct {
	db    *gorm.DB
	apns2 struct {
		BundleId string
		Cert     *tls.Certificate
		Token    *token.Token
	}
	fcmkey     string
	production bool
}

func NewService(db *gorm.DB) *Service {
	s := &Service{db: db}
	cfg := config.GetConfig()
	s.production = cfg.Notification.Production
	// setting up apns
	s.apns2.BundleId = cfg.Notification.Apns2.BundleId
	if len(cfg.Notification.Apns2.CertFile) > 0 {
		if cert, err := certificate.FromP12File(cfg.Notification.Apns2.CertFile, ""); err == nil {
			s.apns2.Cert = &cert
		}
	}
	if len(cfg.Notification.Apns2.JwtToken.AuthKey) > 0 {
		if authKey, err := token.AuthKeyFromFile(cfg.Notification.Apns2.JwtToken.AuthKey); err == nil {
			s.apns2.Token.AuthKey = authKey
			s.apns2.Token.KeyID = cfg.Notification.Apns2.JwtToken.KeyID
			s.apns2.Token.TeamID = cfg.Notification.Apns2.JwtToken.TeamID
		}
	}
	// setting up fcm
	s.fcmkey = cfg.Notification.Fcm.ApiKey
	return s
}

func (s *Service) GenerateNotification(tempName, language string, params ...interface{}) (*MessageLog, error) {

	language = strings.Split(language, "_")[0]
	language = strings.Split(language, "-")[0]
	// find template
	var mt MessageTemplate
	sep := "-"
	if strings.Contains(language, "_") {
		sep = "_"
	}
	shortLang, _, _ := strings.Cut(language, sep)
	if err := s.db.Model(&MessageTemplate{}).First(&mt, "name = ? and lang like concat(?,'-%')", tempName, shortLang).Error; err != nil {
		// en-US as fallback
		if err = s.db.Model(&MessageTemplate{}).First(&mt, "name = ? and lang like concat(?,'-%')", tempName, "en").Error; err != nil {
			return nil, err
		}
	}
	// fill message
	title := mt.Title
	body := fmt.Sprintf(mt.Body, params...)

	// create log
	mlog := &MessageLog{
		TemplateID: mt.ID,
		Title:      title,
		Body:       body,
		Module:     mt.Module,
	}
	log.Debugf("GenerateNotification: <%v> %v", title, body)

	return mlog, nil
}

func (s *Service) GenerateNotificationWithTemplate(tempName, language string, params interface{}) (*MessageLog, error) {

	// find template
	var mt MessageTemplate
	sep := "-"
	if strings.Contains(language, "_") {
		sep = "_"
	}
	shortLang, _, _ := strings.Cut(language, sep)
	if err := s.db.Model(&MessageTemplate{}).First(&mt, "name = ? and lang like concat(?,'-%')", tempName, shortLang).Error; err != nil {
		// en-US as fallback
		shortLang = "en"
		if err = s.db.Model(&MessageTemplate{}).First(&mt, "name = ? and lang like concat(?,'-%')", tempName, shortLang).Error; err != nil {
			return nil, err
		}
	}
	// fill message
	var body string
	title := mt.Title
	funcMap := template.FuncMap{
		// The name "inc" is what the function will be called in the template text.
		"mulf": func(i, j float64) float64 {
			return i * j
		},
		"muli": func(i, j interface{}) int64 {
			var a, b int64
			switch t := i.(type) {
			case int:
				a = int64(i.(int))
			case uint:
				a = int64(i.(uint))
			case int32:
				a = int64(i.(int32))
			case uint32:
				a = int64(i.(uint32))
			case uint64:
				a = int64(i.(uint64))
			case float32:
				a = int64(i.(float32))
			case float64:
				a = int64(i.(float64))
			default:
				log.Debugf("%v", t)
			}
			switch t := j.(type) {
			case int:
				b = int64(j.(int))
			case uint:
				b = int64(j.(uint))
			case int32:
				b = int64(j.(int32))
			case uint32:
				b = int64(j.(uint32))
			case uint64:
				b = int64(j.(uint64))
			case float32:
				b = int64(j.(float32))
			case float64:
				b = int64(j.(float64))
			default:
				log.Debugf("%v", t)
			}
			return a * b
		},
		"unix_tm": func(i uint64) time.Time {
			return time.Unix(int64(i), 0)
		},
	}
	if templ, err := template.New(title).Funcs(funcMap).Parse(mt.Body); err != nil {
		log.Errorf("GenerateNotification: cannot parse template, err = %v", err)
		return nil, err
	} else {
		var buf bytes.Buffer
		templ.Execute(&buf, params)
		body = buf.String()
	}
	// create log
	mlog := &MessageLog{
		TemplateID: mt.ID,
		Title:      title,
		Body:       body,
		Module:     mt.Module,
	}
	log.Debugf("GenerateNotificationWithTemplate: <%v> %v", title, body)

	return mlog, nil
}

func (s *Service) SendNotificationMsgWithoutLog(dest, channel string, mlog *MessageLog, data map[string]string) error {

	return s.SendNotificationMsgWithoutLogV2(dest, channel, mlog, data, 0)
}

func (s *Service) SendNotificationMsgWithoutLogV2(dest, channel string, mlog *MessageLog, data map[string]string, badge int) error {
	// create log
	mlog.Destination = dest
	mlog.Channel = channel

	_, exists := data["channel"]
	if !exists {
		data["channel"] = mlog.Module
	}

	// send
	var err error
	switch channel {
	case "email":
		err = s.sendMail(dest, mlog.Title, mlog.Body)
	case "ios":
		err = s.sendApns(dest, mlog.Title, mlog.Body, data)
	case "android":
		err = s.sendFcm(dest, mlog.Title, mlog.Body, data)
	case "expo":
		err = s.sendExpoWithBadge(dest, mlog.Title, mlog.Body, data, badge)
	case "sms":
	}

	return err
}

func (s *Service) SendNotificationMsg(dest, channel string, mlog *MessageLog, data map[string]string) error {

	if err := s.SendNotificationMsgWithoutLog(dest, channel, mlog, data); err != nil {
		mlog.Result = fmt.Sprintf("%v", err)
		s.db.Save(mlog)
		return err
	}

	mlog.Result = "Success"
	s.db.Save(mlog)
	return nil
}

func (s *Service) SendNotification(dest, channel, tempName, language string, params interface{}) (*MessageLog, error) {
	return s.SendNotificationWithData(dest, channel, tempName, language, map[string]string{}, params)
}

func (s *Service) SendNotificationWithData(dest, channel, tempName, language string, data map[string]string, params interface{}) (*MessageLog, error) {

	mlog, err := s.GenerateNotificationWithTemplate(tempName, language, params)
	if err != nil {
		return nil, err
	}
	err = s.SendNotificationMsg(dest, channel, mlog, data)
	return mlog, err
}

func (s *Service) BroadcastNotification(dests []string, channel, tempName, language string, params interface{}) (*MessageLog, error) {
	return s.BroadcastNotificationWithData(dests, channel, tempName, language, map[string]string{}, params)
}

func (s *Service) BroadcastNotificationWithData(dests []string, channel, tempName, language string, data map[string]string, params interface{}) (*MessageLog, error) {

	msg, err := s.GenerateNotificationWithTemplate(tempName, language, params)
	if err != nil {
		return nil, err
	}

	_, exists := data["channel"]
	if !exists {
		data["channel"] = msg.Module
	}

	// send
	switch channel {
	case "email":
		// err = s.sendMail(dest, title, body)
	case "ios":
		// err = s.sendApns(dest, title, body, data)
	case "android":
		// err = s.sendFcm(dest, title, body, data)
	case "expo":
		err = s.broadcastExpos(dests, msg.Title, msg.Body, data, 0)
	case "sms":
	}

	var mlog *MessageLog
	for _, dest := range dests {
		mlog = &MessageLog{
			TemplateID:  msg.TemplateID,
			Destination: dest,
			Channel:     channel,
			Title:       msg.Title,
			Body:        msg.Body,
		}
		if err != nil {
			mlog.Result = fmt.Sprintf("%v", err)
			s.db.Save(mlog)
			return nil, err
		}
		mlog.Result = "Success"
		s.db.Save(mlog)
	}
	return mlog, nil
}

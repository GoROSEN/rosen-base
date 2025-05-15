package notification

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net/smtp"

	"github.com/GoROSEN/rosen-apiserver/core/config"
	"github.com/jordan-wright/email"
)

func (s *Service) sendMail(mailto, subject, content string) error {

	cfg := config.GetConfig()
	if cfg == nil {
		return errors.New("invalid config")
	}
	if len(cfg.Notification.Smtp.Host) == 0 || cfg.Notification.Smtp.Port == 0 || len(cfg.Notification.Smtp.User) == 0 || len(cfg.Notification.Smtp.Password) == 0 {
		return errors.New("invalid config")
	}
	// from := cfg.Notification.Smtp.From
	fromAddr := cfg.Notification.Smtp.From
	auth := smtp.PlainAuth("", cfg.Notification.Smtp.User, cfg.Notification.Smtp.Password, cfg.Notification.Smtp.Host)
	// to := []string{mailto}

	// content_type := "Content-Type: text/plain; charset=UTF-8"
	// msg := []byte("To: " + strings.Join(to, ",") + "\r\nFrom: " + from +
	// 	"<" + fromAddr + ">\r\nSubject: " + subject + "\r\n" + content_type + "\r\n\r\n" + content)

	e := email.NewEmail()
	e.From = fromAddr
	e.To = []string{mailto}
	e.Subject = subject
	e.Text = []byte(content)

	if cfg.Notification.Smtp.SSL {
		return e.SendWithTLS(fmt.Sprintf("%v:%v", cfg.Notification.Smtp.Host, cfg.Notification.Smtp.Port), auth, &tls.Config{ServerName: cfg.Notification.Smtp.Host})
	} else {
		return e.Send(fmt.Sprintf("%v:%v", cfg.Notification.Smtp.Host, cfg.Notification.Smtp.Port), auth)
	}
	// return smtp.SendMail(fmt.Sprintf("%v:%v", cfg.Notification.Smtp.Host, cfg.Notification.Smtp.Port), auth, fromAddr, to, msg)
}

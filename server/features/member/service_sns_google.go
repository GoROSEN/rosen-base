package member

import (
	"context"
	"errors"

	"github.com/google/martian/log"
	"golang.org/x/oauth2"
	goauth2 "google.golang.org/api/oauth2/v2"
	"google.golang.org/api/option"
)

type googleAuthorize struct {
	AccessToken string
}

func (a *googleAuthorize) Token() (*oauth2.Token, error) {
	return &oauth2.Token{
		AccessToken: a.AccessToken,
		TokenType:   "Bearer",
	}, nil
}

func (s *SnsService) fillGoogleUserInfo(userId, accessToken string, m *Member) error {
	cfg := &googleAuthorize{AccessToken: accessToken}
	ctx := context.Background()
	log.Infof("fillGoogleUserInfo: userId = %v, accessToken = %v", userId, accessToken)
	if oauth2Service, err := goauth2.NewService(ctx, option.WithTokenSource(cfg)); err != nil {
		log.Errorf("cannot get google oauth service: %v", err)
		return err
	} else {
		uis := goauth2.NewUserinfoService(oauth2Service)
		if uis == nil {
			log.Errorf("cannot get google user info service")
			return errors.New("cannot get google user info service")
		}
		if uip, err := uis.Get().Do(); uip == nil || err != nil {
			log.Errorf("cannot get google user info: %v", err)
			return err
		} else {
			m.DisplayName = uip.Name
			m.UserName = uip.Id
			m.Avatar = uip.Picture
			m.Email = uip.Email
		}
	}
	return nil
}

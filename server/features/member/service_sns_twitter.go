package member

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/g8rswimmer/go-twitter/v2"
	"github.com/google/martian/log"
)

type twitterAuthorize struct {
	Token string
}

func (a twitterAuthorize) Add(req *http.Request) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", a.Token))
}

func (s *SnsService) fillTwitterUserInfo(userId, accessToken string, m *Member) error {
	client := &twitter.Client{
		Authorizer: twitterAuthorize{
			Token: accessToken,
		},
		Client: http.DefaultClient,
		Host:   "https://api.twitter.com",
	}
	opts := twitter.UserLookupOpts{
		UserFields: []twitter.UserField{twitter.UserFieldID, twitter.UserFieldUserName, twitter.UserFieldProfileImageURL},
	}

	log.Infof("Callout to tweet lookup callout")

	if userResponse, err := client.AuthUserLookup(context.Background(), opts); err != nil {
		log.Errorf("cannot get twitter user info: %v", err)
		return err
	} else if len(userResponse.Raw.Users) > 0 {
		res := userResponse.Raw.Users[0]
		log.Infof("userResponse data = %v", res)
		if res != nil {
			m.DisplayName = res.UserName
			m.UserName = res.ID
			m.Avatar = res.ProfileImageURL
		} else {
			log.Errorf("got NULL user info")
			return errors.New("got NULL user info")
		}
	} else {
		log.Errorf("got 0 user info")
		return errors.New("got 0 user info")
	}
	return nil
}

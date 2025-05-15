package member

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/GoROSEN/rosen-apiserver/core/config"
	"github.com/Timothylock/go-signin-with-apple/apple"
	"github.com/gofrs/uuid"
	"github.com/google/martian/log"
)

func (s *SnsService) fillAppleUserInfo(userId, accessToken string, m *Member) error {

	var cfg *config.SnsConfig = &config.GetConfig().Sns

	p8, err := os.ReadFile(cfg.Apple.SecretFile)
	if err != nil {
		log.Errorf("cannot read secret file: %v", err)
		return err
	}

	secret, err := apple.GenerateClientSecret(string(p8), cfg.Apple.TeamID, cfg.Apple.ClientID, cfg.Apple.KeyID)
	if err != nil {
		log.Errorf("cannot generate client secret: %v", err)
		return err
	}

	log.Debugf("apple id loging: %v, %v, %v", cfg.Apple.ClientID, secret, accessToken)

	// verify access token
	vReq := apple.AppValidationTokenRequest{
		ClientID:     cfg.Apple.ClientID,
		ClientSecret: secret,
		Code:         accessToken,
	}

	var resp apple.ValidationResponse

	// Do the verification
	client := apple.New()
	if err := client.VerifyAppToken(context.Background(), vReq, &resp); err != nil {
		log.Errorf("cannot verify apple app token: %v", err)
		return err
	}

	if resp.Error != "" {
		fmt.Printf("apple returned an error: %s - %s\n", resp.Error, resp.ErrorDescription)
		return errors.New(resp.Error)
	}

	// jsonstr, _ := json.Marshal(&resp)
	// log.Debugf("response: %v", string(jsonstr))

	uniqueID, err := apple.GetUniqueID(resp.IDToken)
	if err != nil {
		log.Errorf("cannot get unique id: %v", err)
		// return err
	}

	claim, err := apple.GetClaims(resp.IDToken)

	if err != nil {
		log.Errorf("cannot get claims: %v", err)
		return err
	}
	log.Infof("res = %v", claim) //
	if _, ok := (*claim)["email_verified"]; ok {
		verified := (*claim)["email_verified"].(bool)
		log.Infof("email is verified: %v", verified)
	}
	if _, ok := (*claim)["email"]; ok {
		m.Email = (*claim)["email"].(string)
	}
	if _, ok := (*claim)["fullName"]; ok {
		m.DisplayName = (*claim)["fullName"].(string)
	}
	if _, ok := (*claim)["gender"]; ok {
		m.Gender = (*claim)["gender"].(string)
	}
	m.Avatar = "member/avatars/000000.png"
	m.UserName = uniqueID
	t, _ := uuid.NewV4()
	m.DisplayName = fmt.Sprintf("apple-user-%v", t)
	return nil
}

func (s *SnsService) fillAppleWebUserInfo(userId, accessToken string, m *Member) error {

	var cfg *config.SnsConfig = &config.GetConfig().Sns

	p8, err := os.ReadFile(cfg.Apple.SecretFile)
	if err != nil {
		log.Errorf("cannot read secret file: %v", err)
		return err
	}

	secret, err := apple.GenerateClientSecret(string(p8), cfg.Apple.TeamID, cfg.Apple.ServiceID, cfg.Apple.KeyID)
	if err != nil {
		log.Errorf("cannot generate client secret: %v", err)
		return err
	}

	log.Debugf("apple id loging: %v, %v, %v", cfg.Apple.ServiceID, secret, accessToken)

	// verify access token
	vReq := apple.WebValidationTokenRequest{
		ClientID:     cfg.Apple.ServiceID,
		ClientSecret: secret,
		Code:         accessToken,
		RedirectURI:  cfg.Apple.RedirectUrl,
	}

	var resp apple.ValidationResponse

	// Do the verification
	client := apple.New()
	if err := client.VerifyWebToken(context.Background(), vReq, &resp); err != nil {
		log.Errorf("cannot verify apple web token: %v", err)
		return err
	}

	if resp.Error != "" {
		fmt.Printf("apple returned an error: %s - %s\n", resp.Error, resp.ErrorDescription)
		return errors.New(resp.Error)
	}

	// jsonstr, _ := json.Marshal(&resp)
	// log.Debugf("response: %v", string(jsonstr))

	uniqueID, err := apple.GetUniqueID(resp.IDToken)
	if err != nil {
		log.Errorf("cannot get unique id: %v", err)
		// return err
	}

	claim, err := apple.GetClaims(resp.IDToken)

	if err != nil {
		log.Errorf("cannot get claims: %v", err)
		return err
	}
	log.Infof("res = %v", claim) //
	if _, ok := (*claim)["email_verified"]; ok {
		verified := (*claim)["email_verified"].(bool)
		log.Infof("email is verified: %v", verified)
	}
	if _, ok := (*claim)["email"]; ok {
		m.Email = (*claim)["email"].(string)
	}
	if _, ok := (*claim)["fullName"]; ok {
		m.DisplayName = (*claim)["fullName"].(string)
	}
	if _, ok := (*claim)["gender"]; ok {
		m.Gender = (*claim)["gender"].(string)
	}
	m.Avatar = "member/avatars/000000.png"
	m.UserName = uniqueID
	t, _ := uuid.NewV4()
	m.DisplayName = fmt.Sprintf("apple-user-%v", t)
	return nil
}

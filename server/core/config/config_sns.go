package config

type SnsConfig struct {
	Facebook struct {
		AppId       string `yaml:"appId"`
		AppSecret   string `yaml:"appSecret"`
		RedirectUrl string `yaml:"redirectUrl"`
	} `yaml:"facebook"`
	Apple struct {
		TeamID      string `yaml:"teamId"`
		ClientID    string `yaml:"clientId"`
		ServiceID   string `yaml:"serviceId"`
		KeyID       string `yaml:"keyId"`
		SecretFile  string `yaml:"secretFile"`
		RedirectUrl string `yaml:"redirectUrl"`
	} `yaml:"apple"`
}

package config

type Notification struct {
	Production bool `yaml:"production"`

	Smtp struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		From     string `yaml:"from"`
		SSL      bool   `yaml:"ssl"`
	} `yaml:"smtp"`

	Apns2 struct {
		BundleId string `yaml:"bundleId"`
		CertFile string `yaml:"cert"`
		JwtToken struct {
			AuthKey string `yaml:"key"`
			KeyID   string `yaml:"id"`
			TeamID  string `yaml:"team"`
		} `yaml:"jwt"`
	} `yaml:"apns2"`

	Fcm struct {
		CredentialsFile string `yaml:"credentials"`
		ApiKey          string `yaml:"apikey"`
	} `yaml:"fcm"`

	Expo struct {
		AccessToken string `yaml:"accessToken"`
	} `yaml:"expo"`
}

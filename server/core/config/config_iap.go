package config

// InAppPurchaseConfig 内购配置项
type InAppPurchaseConfig struct {
	Apple struct {
		RootKey  string `yaml:"root"`
		KeyFile  string `yaml:"keyfile"`
		KeyID    string `yaml:"keyId"`
		BundleID string `yaml:"bundleId"`
		Issuer   string `yaml:"issuer"`
		Sandbox  bool   `yaml:"sandbox"`
	} `yaml:"apple"`

	Google struct {
		Package string `yaml:"package"`
		KeyFile string `yaml:"keyfile"`
	} `yaml:"google"`
}

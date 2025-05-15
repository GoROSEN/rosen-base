package config

type TranslationConfig struct {
	ApiKey    string   `yaml:"apikey"`
	Endpoint  string   `yaml:"endpoint"`
	Region    string   `yaml:"region"`
	Languages []string `yaml:"languages"`
}

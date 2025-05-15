package config

// OssConfig OSS配置项
type OssConfig struct {
	Type              string `yaml:"type" json:"type"`
	Endpoint          string `yaml:"endpoint" json:"endpoint"`
	AccelEndpoint     string `yaml:"accelEndpoint" json:"accelEndpoint"`
	AccessKeyID       string `yaml:"accessKeyId" json:"accessKeyId"`
	AccessKeySecret   string `yaml:"accessKeySecret" json:"accessKeySecret"`
	PrivateBucket     string `yaml:"privateBucket" json:"privateBucket"`
	PublicBucket      string `yaml:"publicBucket" json:"publicBucket"`
	Region            string `yaml:"region" json:"region"`
	SSL               bool   `yaml:"ssl" json:"ssl"`
	PresignedDuration int64  `yaml:"presignedDuration" json:"-"`
}

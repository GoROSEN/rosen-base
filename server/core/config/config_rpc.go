package config

type RpcConfig struct {
	Enable bool `yaml:"enable"`
	Amqp   struct {
		Host     string `yaml:"host"`
		Port     uint16 `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
	} `yaml:"amqp"`
	Queues map[string]string `yaml:"queues"`
}

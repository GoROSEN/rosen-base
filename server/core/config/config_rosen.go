package config

type RosenCoinConfig struct {
	TokenName string `yaml:"name"`
	Decimals  uint64 `yaml:"decimals"`
}

type RosenConfig struct {
	VCode bool `yaml:"vcode"`
	MTE   struct {
		KeepAliveDurationInSec int `yaml:"keepalive"`
	} `yaml:"mte"`
	Ipfs struct {
		RpcAddr string `yaml:"rpc"`
		Gateway string `yaml:"gateway"`
	} `yaml:"ipfs"`
	Chains        []BlockchainConfig  `yaml:"chains"`
	CircleService CircleServiceConfig `yaml:"circle"`
	Energy        struct {
		Decimals uint64 `yaml:"decimals"`
	} `yaml:"energy"`
	Coin  RosenCoinConfig `yaml:"coin"`
	Coin2 RosenCoinConfig `yaml:"coin2"`
}

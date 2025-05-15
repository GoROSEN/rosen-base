// Package utils 工具库
package config

import (
	"errors"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v2"
)

// Config 配置信息
type Config struct {
	Db struct {
		Driver  string `yaml:"driver"`
		ConnStr string `yaml:"conn_str"`
	}
	Token struct {
		TokenLife        time.Duration `yaml:"token_life"`
		RefreshTokenLife time.Duration `yaml:"refresh_token_life"`
	}
	Redis struct {
		Host     string `yaml:"host"`
		Port     string `yaml:"port"`
		Password string `yaml:"password"`
		DB       int    `yaml:"db"`
	}
	Geoip struct {
		DB   string `yaml:"db"`
		Lang string `yaml:"lang"`
	}
	Content ContentConfig
	Cors    struct {
		Enable       bool     `yaml:"enable"`
		AllowOrigins []string `yaml:"allow_origins"`
		AllowHeaders string   `yaml:"allow_headers"`
	}
	Oss    OssConfig `yaml:"oss" json:"oss"`
	Wechat struct {
		AppID             string `yaml:"appId"`
		AppSecret         string `yaml:"appSecret"`
		AuthCallbackURL   string `yaml:"authCallbackURL"`
		VerifyFileURI     string `yaml:"verifyFileURI"`
		VerifyFileContent string `yaml:"verifyFileContent"`
	}
	Web struct {
		SessionToken string `yaml:"sessionToken"`
		CaptichaFont string `yaml:"capichaFont"`
	}
	Magick struct {
		Font string `yaml:"font"`
	}
	Logging struct {
		Console string `yaml:"console"`
		Level   string `yaml:"level"`
		File    string `yaml:"file"`
	}
	Notification  Notification `yaml:"notification"`
	ActiveModules string       `yaml:"modules"`
	EnableCronJob bool         `yaml:"enableCronjob"`

	Rosen         RosenConfig         `yaml:"rosen"`
	Sns           SnsConfig           `yaml:"sns"`
	BlockedMails  []string            `yaml:"blockedMails"`
	Rpc           RpcConfig           `yaml:"rpc"`
	InAppPurchase InAppPurchaseConfig `yaml:"in-app-purchase"`
	Translation   TranslationConfig   `yaml:"translation"`
}

var gConfig *Config

// GetConfig 获取全局配置
func GetConfig() *Config {

	if gConfig == nil {
		gConfig = &Config{}
	}
	return gConfig
}

// LoadFromFile 从文件加载配置
func (c *Config) LoadFromFile(filename string) error {

	yamlFile, err := os.ReadFile(filename)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(yamlFile, c)
	if err != nil {
		return err
	}

	return nil
}

// LoadFromEnv 从环境变量加载配置
func (c *Config) LoadFromEnv() error {

	// DB
	c.Db.Driver = os.Getenv("ARK_DB_DRIVER")
	c.Db.ConnStr = os.Getenv("ARK_DB_CONN")
	// Token
	c.Token.TokenLife, _ = time.ParseDuration(os.Getenv("ARK_TOKEN_LIFE"))
	c.Token.RefreshTokenLife, _ = time.ParseDuration(os.Getenv("ARK_REFRESH_TOKEN_LIFE"))
	// Redis
	c.Redis.Host = os.Getenv("ARK_REDIS_HOST")
	c.Redis.Port = os.Getenv("ARK_REDIS_PORT")
	c.Redis.Password = os.Getenv("ARK_REDIS_PASSWORD")
	c.Redis.DB, _ = strconv.Atoi(os.Getenv("ARK_REDIS_DB"))
	// Storage
	c.Content.LocalPrefix = os.Getenv("ARK_CONTENT_LOCAL_PREFIX")
	c.Content.RemotePrefix = os.Getenv("ARK_CONTENT_REMOTE_PREFIX")
	// CORS
	c.Cors.Enable, _ = strconv.ParseBool(os.Getenv("ARK_CORS_ENABLE"))
	c.Cors.AllowOrigins = strings.Split(os.Getenv("ARK_CORS_ALLOW_ORIGINS"), ",")
	c.Cors.AllowHeaders = os.Getenv("ARK_CORS_ALLOW_HEADERS")
	// OSS
	c.Oss.Type = os.Getenv("ARK_OSS_TYPE")
	c.Oss.AccessKeyID = os.Getenv("ARK_OSS_ACCESS_KEY_ID")
	c.Oss.AccessKeySecret = os.Getenv("ARK_OSS_ACCESS_KEY_SECRET")
	c.Oss.PrivateBucket = os.Getenv("ARK_OSS_PRIVATE_BUCKET")
	c.Oss.PublicBucket = os.Getenv("ARK_OSS_PUBLIC_BUCKET")
	c.Oss.Region = os.Getenv("ARK_OSS_REGION")
	c.Oss.SSL, _ = strconv.ParseBool(os.Getenv("ARK_OSS_SSL"))
	c.Oss.PresignedDuration, _ = strconv.ParseInt(os.Getenv("ARK_OSS_PRESIGNED_DURATION"), 10, 64)
	c.Oss.Endpoint = os.Getenv("ARK_OSS_ENDPOINT")
	c.Oss.AccelEndpoint = os.Getenv("ARK_OSS_ACCEL_ENDPOINT")
	// Wechat
	c.Wechat.AppID = os.Getenv("ARK_WECHAT_APP_ID")
	c.Wechat.AppSecret = os.Getenv("ARK_WECHAT_APP_SECRET")
	c.Wechat.AuthCallbackURL = os.Getenv("ARK_WECHAT_AUTH_CALLBACK_URL")
	c.Wechat.VerifyFileURI = os.Getenv("ARK_WECHAT_VERIFY_FILE_URI")
	c.Wechat.VerifyFileContent = os.Getenv("ARK_WECHAT_VERIFY_FILE_CONTENT")
	// web
	c.Web.SessionToken = os.Getenv("ARK_WEB_SESSION_TOKEN")
	c.Web.CaptichaFont = os.Getenv("ARK_WEB_CAPICHA_FONT")
	// imagemagick
	c.Magick.Font = os.Getenv("ARK_MAGICK_FONT")
	// geoip
	c.Geoip.DB = os.Getenv("ARK_GEOIP_DB")
	c.Geoip.Lang = os.Getenv("ARK_GEOIP_LANG")
	// active module
	c.ActiveModules = os.Getenv("ARK_MODULES")
	// cron jobs
	c.EnableCronJob, _ = strconv.ParseBool(os.Getenv("ARK_ENABLE_CRONJOB"))
	// notification-smtp
	c.Notification.Smtp.Host = os.Getenv("ARK_NOTIFICATION_SMTP_HOST")
	c.Notification.Smtp.Port, _ = strconv.Atoi(os.Getenv("ARK_NOTIFICATION_SMTP_PORT"))
	c.Notification.Smtp.User = os.Getenv("ARK_NOTIFICATION_SMTP_USERNAME")
	c.Notification.Smtp.Password = os.Getenv("ARK_NOTIFICATION_SMTP_PASSWORD")
	c.Notification.Smtp.From = os.Getenv("ARK_NOTIFICATION_SMTP_FROM")
	c.Notification.Smtp.SSL, _ = strconv.ParseBool(os.Getenv("ARK_NOTIFICATION_SMTP_SSL"))
	// notification-apns/2
	// notification-fcm
	// notification-sms
	// notification-rosen
	c.Rosen.Ipfs.RpcAddr = os.Getenv("ROSEN_IPFS_RPC")
	c.Rosen.Ipfs.Gateway = os.Getenv("ROSEN_IPFS_GATEWAY")
	c.Rosen.MTE.KeepAliveDurationInSec, _ = strconv.Atoi(os.Getenv("ROSEN_MTE_KEEPALIVE"))
	// sns
	c.Sns.Facebook.AppId = os.Getenv("SNS_FACEBOOK_APP_ID")
	c.Sns.Facebook.AppSecret = os.Getenv("SNS_FACEBOOK_APP_SECRET")
	c.Sns.Facebook.RedirectUrl = os.Getenv("SNS_FACEBOOK_REDIRECT_URL")
	// blocked mails
	c.BlockedMails = []string{}
	// logging
	c.Logging.File = os.Getenv("ARK_LOGGING_FILE")
	c.Logging.Level = os.Getenv("ARK_LOGGING_LEVEL")

	if c.Db.Driver == "" || c.Db.ConnStr == "" {
		return errors.New("blank env")
	}
	return nil
}

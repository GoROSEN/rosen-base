package common

import (
	"io"
	"net/url"
	"time"

	"github.com/GoROSEN/rosen-apiserver/core/config"
	"github.com/google/martian/log"
)

// OssFileInfo oss文件信息
type OssFileInfo struct {
	Size        int64
	ContentType string
	Stream      io.ReadCloser // 需要调用方关闭流
}

// OssService OSS抽象接口
type OssService interface {
	UploadFile(file io.Reader, bucketName, name, contentType string, size int64) error
	PresignedUploadUrl(bucketName, filename string, expires time.Duration) (*url.URL, error)
	DownloadFile(bucketName, filename string) (*OssFileInfo, error)
	PresignedDownloadUrl(bucketName, filename string, expires time.Duration) (*url.URL, error)
	DownloadUrl(bucketName, filename string) (*url.URL, error)
}

// NewOssService 根据配置创建新的OSS服务
func NewOssService(cfg *config.OssConfig) OssService {
	log.Debugf("creating oss service for %v", cfg.Type)
	if cfg.Type == "aliyun" {
		return NewAliyunOssService(cfg)
	} else if cfg.Type == "s3" {
		return NewS3OssService(cfg)
	}
	return nil
}

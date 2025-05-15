package common

import (
	"errors"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/GoROSEN/rosen-apiserver/core/config"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/google/martian/log"
)

// AliyunOssService 阿里云OSS
type AliyunOssService struct {
	client *oss.Client
	config *config.OssConfig
}

// NewAliyunOssService 新建OSS服务
func NewAliyunOssService(cfg *config.OssConfig) *AliyunOssService {

	svc := &AliyunOssService{}
	client, err := oss.New(cfg.Endpoint, cfg.AccessKeyID, cfg.AccessKeySecret)
	if err != nil {
		log.Errorf("cannot init aliyun oss client: %v", err)
		return nil
	}
	svc.client = client
	svc.config = cfg
	return svc
}

func (svc *AliyunOssService) UploadFile(file io.Reader, bucketName, name, contentType string, size int64) error {

	if svc.client == nil {
		return errors.New("invalid client")
	}
	bucket, err := svc.client.Bucket(bucketName)
	if err != nil {
		log.Errorf("UploadFile: cannot get bucket: %v", err)
		return err
	}
	if err := bucket.PutObject(name, file); err != nil {
		return err
	}

	return nil
}

func (svc *AliyunOssService) PresignedUploadUrl(bucketName, filename string, expires time.Duration) (*url.URL, error) {

	if svc.client == nil {
		return nil, errors.New("invalid client")
	}

	bucket, err := svc.client.Bucket(bucketName)
	if err != nil {
		log.Errorf("PresignedUploadUrl: cannot get bucket: %v", err)
		return nil, err
	}

	signedURL, err := bucket.SignURL(filename, oss.HTTPPut, (int64)(expires.Seconds()))
	if err != nil {
		log.Errorf("PresignedUploadUrl: cannot get url: %v", err)
		return nil, err
	}
	return url.Parse(signedURL)
}

func (svc *AliyunOssService) DownloadFile(bucketName, filename string) (*OssFileInfo, error) {

	if svc.client == nil {
		return nil, errors.New("invalid client")
	}

	bucket, err := svc.client.Bucket(bucketName)
	if err != nil {
		log.Errorf("DownloadFile: cannot get bucket: %v", err)
		return nil, err
	}

	props, err := bucket.GetObjectDetailedMeta(filename)
	if err != nil {
		log.Errorf("DownloadFile: cannot get file meta: %v", err)
		return nil, err
	}

	body, err := bucket.GetObject(filename)
	if err != nil {
		log.Errorf("DownloadFile: cannot get object: %v", err)
		return nil, err
	}

	size, err := strconv.ParseInt(props.Get("Content-Length"), 10, 64)
	if err != nil {
		log.Errorf("DownloadFile: cannot get convert content-length to int: %v", err)
		return nil, err
	}

	return &OssFileInfo{Size: size, ContentType: props.Get("Content-Type"), Stream: body}, nil
}

func (svc *AliyunOssService) PresignedDownloadUrl(bucketName, filename string, expires time.Duration) (*url.URL, error) {

	if svc.client == nil {
		return nil, errors.New("invalid client")
	}

	bucket, err := svc.client.Bucket(bucketName)
	if err != nil {
		log.Errorf("PresignedDownloadUrl: cannot get bucket: %v", err)
		return nil, err
	}

	signedURL, err := bucket.SignURL(filename, oss.HTTPGet, (int64)(expires.Seconds()))
	if err != nil {
		log.Errorf("PresignedDownloadUrl: cannot get url: %v", err)
		return nil, err
	}
	return url.Parse(signedURL)
}

func (svc *AliyunOssService) DownloadUrl(bucketName, filename string) (*url.URL, error) {

	if svc.client == nil {
		return nil, errors.New("invalid client")
	}

	_, err := svc.client.Bucket(bucketName)
	if err != nil {
		log.Errorf("PresignedDownloadUrl: cannot get bucket: %v", err)
		return nil, err
	}

	if len(svc.config.AccelEndpoint) > 0 {
		if !strings.Contains(filename, svc.config.AccelEndpoint) {
			return url.Parse(fmt.Sprintf("%v/%v", svc.config.AccelEndpoint, filename))
		} else {
			return url.Parse(filename)
		}
	} else {
		if !strings.Contains(filename, svc.client.Config.Endpoint) {
			return url.Parse(fmt.Sprintf("https://%v/%v", svc.client.Config.Endpoint, filename))
		} else {
			return url.Parse(filename)
		}
	}
}

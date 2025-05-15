package common

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/GoROSEN/rosen-apiserver/core/config"
	"github.com/google/martian/log"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// S3OssService S3接口OSS客户端
type S3OssService struct {
	minioClient *minio.Client
	config      *config.OssConfig
}

// NewS3OssService 根据配置新建OSS服务
func NewS3OssService(cfg *config.OssConfig) *S3OssService {
	svc := &S3OssService{}
	// Initialize minio client object.
	minioClient, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.AccessKeySecret, ""),
		Secure: cfg.SSL,
	})
	if err != nil {
		log.Errorf("cannot init minio client: %v", err)
		return nil
	}
	svc.minioClient = minioClient
	svc.config = cfg

	return svc
}

func (svc *S3OssService) UploadFile(file io.Reader, bucketName, name, contentType string, size int64) error {

	// 上传至oss
	if svc.minioClient == nil {
		return errors.New("invalid client")
	}

	// Initialize minio client object.
	err := svc.minioClient.MakeBucket(context.Background(), bucketName, minio.MakeBucketOptions{Region: svc.config.Region})
	if err != nil {
		// Check to see if we already own this bucket (which happens if you run this twice)
		exists, errBucketExists := svc.minioClient.BucketExists(context.Background(), bucketName)
		if errBucketExists == nil && exists {
			log.Infof("We already own %v", bucketName)
		} else {
			log.Errorf("access bucket failed: %v", err)
			return err
		}
	} else {
		log.Infof("Successfully created %v", bucketName)
	}
	// Upload the file with FPutObject
	n, err := svc.minioClient.PutObject(context.Background(), bucketName, name, file, size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		log.Errorf("put object error: ", err)
		return err
	}
	log.Infof("Successfully uploaded %v of size %v", name, n)

	return nil
}

func (svc *S3OssService) PresignedUploadUrl(bucketName, filename string, expires time.Duration) (*url.URL, error) {

	if svc.minioClient == nil {
		return nil, errors.New("invalid client")
	}

	// Initialize minio client object.
	presignedURL, err := svc.minioClient.PresignedPutObject(context.Background(), bucketName, filename, expires)
	if err != nil {
		log.Errorf("get presigned url error: %v", err)
		return nil, err
	}

	return presignedURL, err
}

func (svc *S3OssService) DownloadFile(bucketName, filename string) (*OssFileInfo, error) {

	if svc.minioClient == nil {
		return nil, errors.New("invalid client")
	}

	objInfo, err := svc.minioClient.StatObject(context.Background(), bucketName, filename, minio.StatObjectOptions{})
	if err != nil {
		log.Errorf("state file error: %v", err)
		return nil, err
	}

	log.Infof("file size = %v, content type = %v", objInfo.Size, objInfo.ContentType)

	object, err := svc.minioClient.GetObject(context.Background(), bucketName, filename, minio.GetObjectOptions{})
	if err != nil {
		log.Errorf("download file error: %v", err)
		return nil, err
	}
	log.Infof("sending file")
	return &OssFileInfo{Size: objInfo.Size, ContentType: objInfo.ContentType, Stream: object}, nil
}

func (svc *S3OssService) PresignedDownloadUrl(bucketName, filename string, expires time.Duration) (*url.URL, error) {

	if svc.minioClient == nil {
		return nil, errors.New("invalid client")
	}

	log.Infof("stating object %v/%v", bucketName, filename)
	objInfo, err := svc.minioClient.StatObject(context.Background(), bucketName, filename, minio.StatObjectOptions{})
	if err != nil {
		log.Errorf("state file error: %v", err)
		return nil, err
	}

	log.Infof("file size = %v, content type = %v", objInfo.Size, objInfo.ContentType)
	reqParams := make(url.Values)
	reqParams.Set("response-content-disposition", fmt.Sprintf("attachment; filename=\"%v\"", filename))
	return svc.minioClient.PresignedGetObject(context.Background(), bucketName, filename, expires, reqParams)
}

func (svc *S3OssService) DownloadUrl(bucketName, filename string) (*url.URL, error) {

	if svc.minioClient == nil {
		return nil, errors.New("invalid client")
	}

	log.Infof("stating object %v/%v", bucketName, filename)
	_, err := svc.minioClient.StatObject(context.Background(), bucketName, filename, minio.StatObjectOptions{})
	if err != nil {
		log.Errorf("state file error: %v", err)
		return nil, err
	}

	if len(svc.config.AccelEndpoint) > 0 {
		if !strings.Contains(filename, svc.config.AccelEndpoint) {
			return url.Parse(fmt.Sprintf("%v/%v/%v", svc.config.AccelEndpoint, bucketName, filename))
		} else {
			return url.Parse(fmt.Sprintf("%v/%v", bucketName, filename))
		}
	} else {
		return svc.minioClient.EndpointURL().Parse(fmt.Sprintf("%v/%v", bucketName, filename))
	}
}

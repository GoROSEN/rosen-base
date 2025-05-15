package common

import (
	"archive/zip"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"io/ioutil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/GoROSEN/rosen-apiserver/core/config"
	"github.com/GoROSEN/rosen-apiserver/core/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/martian/log"
	uuid "github.com/satori/go.uuid"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// OssController 文件处理控制器
type OssController struct {
	Oss                  OssService
	privateBucketName    string
	publicBucketName     string
	ossFilePrefix        string
	ossPresignedDuration time.Duration
}

// SetupOSS 初始化Oss配置
func (c *OssController) SetupOSS(filePrefix string) {
	cfg := config.GetConfig().Oss
	c.SetupOSSWithBucketName(filePrefix, cfg.PublicBucket, cfg.PrivateBucket)
}

// SetupOSS 初始化Oss配置
func (c *OssController) SetupOSSWithBucketName(filePrefix, publicBucket, privateBucket string) {
	log.Infof("setting up oss controller, filePrefix = %v", filePrefix)
	cfg := config.GetConfig()
	c.privateBucketName = privateBucket
	c.publicBucketName = publicBucket
	c.ossFilePrefix = filePrefix
	c.ossPresignedDuration = time.Duration(cfg.Oss.PresignedDuration) * time.Second
	c.Oss = NewOssService(&cfg.Oss)
}

// UploadPrivateFile 上传文件并返回文件路径名，需使用get presigned url获取下载路径
func (c *OssController) UploadPrivateFile(ctx *gin.Context) {

	filePathName, _, err := c.SaveFormFileToOSS(ctx, "file", uuid.NewV4().String(), false)
	if err != nil {
		log.Errorf("failed to save file: %v", err)
		utils.SendFailureResponse(ctx, 500, "message.common.system-error")
		return
	}

	presignedUrl, _ := c.PresignedOssDownloadURLWithoutPrefix(filePathName)

	utils.SendSuccessResponse(ctx, gin.H{"path": filePathName, "url": presignedUrl.String()})
}

// GetPrivateFilePresignedURL 获取:file(base64)指定的文件的签名下载路径
func (c *OssController) GetPrivateFilePresignedURL(ctx *gin.Context) {
	str := ctx.Param("file")
	filePathName, err := base64.URLEncoding.DecodeString(str)
	if err != nil {
		log.Errorf("failed to decode file name: %v", err)
		utils.SendFailureResponse(ctx, 500, "message.common.system-error")
		return
	}

	url, err := c.Oss.PresignedDownloadUrl(c.privateBucketName, string(filePathName), c.ossPresignedDuration)
	if err != nil {
		log.Errorf("failed to get file url: %v", err)
		utils.SendFailureResponse(ctx, 500, "message.common.system-error")
		return
	}
	utils.SendSuccessResponse(ctx, url.String())
}

// UploadPublicFile 上传文件并返回文件URL
func (c *OssController) UploadPublicFile(ctx *gin.Context) {

	fileName := uuid.NewV4().String()
	pathname, _, err := c.SaveFormFileToOSS(ctx, "file", fileName, true)
	if err != nil {
		log.Errorf("failed to save file: %v", err)
		utils.SendFailureResponse(ctx, 500, "message.common.system-error")
		return
	}

	url, err := c.OssDownloadURLWithoutPrefix(pathname)
	if err != nil {
		log.Errorf("failed to get file url: %v", err)
		utils.SendFailureResponse(ctx, 500, "message.common.system-error")
		return
	}

	utils.SendSuccessResponse(ctx, gin.H{"url": url.String()})
}

/// Aux

// GetFormFileName 获取表单文件名
func (*OssController) GetFormFileName(ctx *gin.Context, fieldName string) (string, error) {
	_, header, err := ctx.Request.FormFile(fieldName)
	if err != nil {
		log.Errorf("failed to get file name: %v", err)
		return "", err
	}
	return header.Filename, nil
}

// SaveFormFile 将表单文件保存到指定位置，返回原始文件名
func (*OssController) SaveFormFile(ctx *gin.Context, fieldName, toFileName string) (string, string, error) {

	file, header, err := ctx.Request.FormFile(fieldName)
	if err != nil {
		log.Errorf("failed to get file: %v", err)
		utils.SendFailureResponse(ctx, 500, "message.common.system-error")
		return "", "", err
	}

	filename := toFileName + filepath.Ext(header.Filename)
	log.Debugf("saved to %v", filename)

	out, err := os.Create(filename)
	if err != nil {
		log.Errorf("create file failed: %v", err)
		utils.SendFailureResponse(ctx, 500, "message.common.system-error")
		return "", header.Filename, err
	}
	defer out.Close()

	_, err = io.Copy(out, file)
	if err != nil {
		log.Errorf("copy file failed: %v", err)
		utils.SendFailureResponse(ctx, 500, "message.common.system-error")
		return "", header.Filename, err
	}

	// 原始文件名：header.Filename，先不要了
	return filename, header.Filename, nil
}

// SaveFormFileToOSS 将表单文件保存到OSS指定位置，返回OSS文件名
func (c *OssController) SaveFormFileToOSS(ctx *gin.Context, fieldName string, toFileName string, isPublic bool) (string, string, error) {

	file, header, err := ctx.Request.FormFile(fieldName)
	if err != nil {
		return "", "", err
	}
	tfilename := ctx.Request.FormValue("filename")

	var bucketName string
	if isPublic {
		bucketName = c.publicBucketName
	} else {
		bucketName = c.privateBucketName
	}
	if len(tfilename) == 0 {
		tfilename = header.Filename
	}
	filename := path.Join(c.ossFilePrefix, toFileName+filepath.Ext(tfilename))
	log.Debugf("file name: %v", filename)
	var contentType string
	if strings.HasSuffix(filename, ".jpg") || strings.HasSuffix(filename, ".jpeg") {
		contentType = "image/jpeg"
	} else if strings.HasSuffix(filename, ".png") {
		contentType = "image/png"
	}
	log.Infof("contentType: %v", contentType)
	if err := c.Oss.UploadFile(file, bucketName, filename, contentType, header.Size); err != nil {
		return "", header.Filename, err
	}
	return filename, header.Filename, nil
}

func (c *OssController) SaveFileToOSS(data []byte, fileName, contentType, toFileName string, isPublic bool) (string, error) {

	var bucketName string
	if isPublic {
		bucketName = c.publicBucketName
	} else {
		bucketName = c.privateBucketName
	}
	fs := bytes.NewReader(data)
	filename := path.Join(c.ossFilePrefix, toFileName+filepath.Ext(fileName))
	log.Infof("contentType: %v", contentType)
	if err := c.Oss.UploadFile(fs, bucketName, filename, contentType, int64(len(data))); err != nil {
		return "", err
	}
	return filename, nil
}

func (c *OssController) SaveStreamToOSS(fs io.Reader, fileName, contentType, toFileName string, isPublic bool, fileSize uint64) (string, error) {

	var bucketName string
	if isPublic {
		bucketName = c.publicBucketName
	} else {
		bucketName = c.privateBucketName
	}
	filename := path.Join(c.ossFilePrefix, toFileName+filepath.Ext(fileName))
	log.Infof("contentType: %v", contentType)
	if err := c.Oss.UploadFile(fs, bucketName, filename, contentType, int64(fileSize)); err != nil {
		return "", err
	}
	return filename, nil
}

// ExtractFormZipFileToOss 返回：map[uuid][]string{ossFileName, originFileName}
func (c *OssController) ExtractFormZipFileToOss(ctx *gin.Context, fieldName, toFolder string, isPublic bool) (map[string][]string, error) {
	tmpFileName := utils.TempFilePathName()
	src, _, err := c.SaveFormFile(ctx, "file", tmpFileName)
	if err != nil {
		log.Errorf("failed to save zip file: %v", err)
		return nil, err
	}
	zr, err := zip.OpenReader(src)
	if err != nil {
		log.Errorf("failed to open zip file: %v", err)
		return nil, err
	}
	defer zr.Close()
	if err != nil {
		log.Errorf("failed to unzip file: %v", err)
		return nil, err
	}
	results := map[string][]string{}
	// 遍历 zr ，将文件写入到oss
	for _, file := range zr.File {

		// 如果是目录，就跳过
		if file.FileInfo().IsDir() {
			continue
		}

		// 获取到 Reader
		fr, err := file.Open()
		if err != nil {
			log.Errorf("failed to read archived file: %v", err)
			continue
		}

		localFileName := file.Name
		if file.Flags == 0 {
			// 文件名为GBK，需要转换
			i := bytes.NewReader([]byte(file.Name))
			decoder := transform.NewReader(i, simplifiedchinese.GB18030.NewDecoder())
			content, _ := ioutil.ReadAll(decoder)
			localFileName = string(content)
		}

		uuidStr := uuid.NewV4().String()
		ossFileName := fmt.Sprintf("%v/%v", toFolder, uuidStr)
		log.Debugf("saving file '%v' to oss: %v (%v bytes)", localFileName, ossFileName, file.UncompressedSize64)
		ossFilePathName, err := c.SaveStreamToOSS(fr, localFileName, "", ossFileName, false, file.UncompressedSize64)
		if err != nil {
			log.Errorf("failed to save oss file: %v", err)
			continue
		}
		results[uuidStr] = []string{ossFilePathName, localFileName}

		// 因为是在循环中，无法使用 defer ，直接放在最后
		// 不过这样也有问题，当出现 err 的时候就不会执行这个了，
		// 可以把它单独放在一个函数中，这里是个实验，就这样了
		fr.Close()
	}
	return results, nil
}

// StreamOutOssFile 找到并将oss文件以流形式输出
func (c *OssController) StreamOutOssFile(ctx *gin.Context, fileName string, isPublic bool) error {

	return c.StreamOutOssFileWithoutPrefix(ctx, path.Join(c.ossFilePrefix, fileName), isPublic)
}

// StreamOutOssFile 找到并将oss文件以流形式输出
func (c *OssController) StreamOutOssFileWithoutPrefix(ctx *gin.Context, fileName string, isPublic bool) error {

	var bucketName string
	if isPublic {
		bucketName = c.publicBucketName
	} else {
		bucketName = c.privateBucketName
	}
	objInfo, err := c.Oss.DownloadFile(bucketName, fileName)
	if err != nil {
		log.Errorf("state file error: %v", err)
		return err
	}

	log.Infof("file size = %v, content type = %v", objInfo.Size, objInfo.ContentType)

	log.Infof("sending file")
	defer objInfo.Stream.(io.Closer).Close()
	ctx.DataFromReader(200, objInfo.Size, objInfo.ContentType, objInfo.Stream, nil)
	return nil
}

// PresignedOssDownloadURL 获取预签名URL
func (c *OssController) PresignedOssDownloadURL(filePathName string) (*url.URL, error) {
	filepath := path.Join(c.ossFilePrefix, filePathName)
	log.Infof("presigning %v/%v with %v", c.privateBucketName, filepath, c.ossPresignedDuration)
	return c.Oss.PresignedDownloadUrl(c.privateBucketName, filepath, c.ossPresignedDuration)
}

// PresignedOssDownloadURLWithoutPrefix 获取预签名URL
func (c *OssController) PresignedOssDownloadURLWithoutPrefix(filePathName string) (*url.URL, error) {
	log.Infof("presigning %v/%v with %v", c.privateBucketName, filePathName, c.ossPresignedDuration)
	return c.Oss.PresignedDownloadUrl(c.privateBucketName, filePathName, c.ossPresignedDuration)
}

// OssDownloadURLWithoutPrefix 获取URL
func (c *OssController) OssDownloadURLWithoutPrefix(filePathName string) (*url.URL, error) {
	// log.Infof("get url for %v/%v", c.bucketName, filePathName)
	if strings.HasPrefix(filePathName, "http") {
		return url.Parse(filePathName)
	}
	return c.Oss.DownloadUrl(c.publicBucketName, filePathName)
}

// OssDownloadURL 获取URL
func (c *OssController) OssDownloadURL(filePathName string) (*url.URL, error) {
	// log.Infof("get url for %v/%v/%v", c.bucketName, c.ossFilePrefix, filePathName)
	if strings.HasPrefix(filePathName, "http") {
		return url.Parse(filePathName)
	}
	return c.Oss.DownloadUrl(c.publicBucketName, path.Join(c.ossFilePrefix, filePathName))
}

// DownloadFile 下载OSS文件
func (c *OssController) DownloadFile(filePathName string, isPublic, withPrefix bool) (*OssFileInfo, error) {
	var bucket, prefix string
	if isPublic {
		bucket = c.publicBucketName
	} else {
		bucket = c.privateBucketName
	}
	if withPrefix {
		prefix = c.ossFilePrefix
	}
	log.Debugf("downloading oss file: %v%v%v", bucket, prefix, filePathName)
	return c.Oss.DownloadFile(bucket, path.Join(prefix, filePathName))
}

func (c *OssController) PresignedUploadURL(filePathName string) (*url.URL, error) {
	if strings.HasPrefix(filePathName, "http") {
		return url.Parse(filePathName)
	}
	return c.Oss.PresignedUploadUrl(c.privateBucketName, path.Join(c.ossFilePrefix, filePathName), c.ossPresignedDuration)
}

package auth

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/GoROSEN/rosen-apiserver/core/utils"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v7"
	"github.com/google/martian/log"
	"gorm.io/gorm"
)

// Interceptor 认证拦截器
type Interceptor struct {
	service *Service
}

// NewInterceptor 初始化拦截器
func NewInterceptor(redisClient *redis.Client, db *gorm.DB) *Interceptor {

	i := &Interceptor{}
	i.service = NewAuthService(redisClient, db)
	return i
}

// AuthInterceptor 认证拦截器
func (c *Interceptor) AuthInterceptor(ctx *gin.Context) {

	token, result := ctx.GetQuery("token")
	if !result {
		token = ctx.GetHeader("X-Token")
		result = len(token) > 0
	}
	if !result {
		log.Infof("token not detected, set empty user info")
		ctx.Set("userid", 0)
		ctx.Set("member-id", 0)
		ctx.Set("companyid", 0)
		ctx.Set("userrole", []string{})
		ctx.Next()
		return
	}

	// member
	uid, err := c.service.VerifyMemberToken(token)
	if err == nil {
		log.Debugf("AuthInterceptor: member-id = %v", uid)
		log.Infof("set member-id to context")
		ctx.Set("member-id", int(uid))
		ctx.Set("member-token", token)
	}

	// admin
	t, err := c.service.VerifyUserToken(token)
	if err == nil {
		log.Debugf("AuthInterceptor: uid = %v, role = %v", t.UserID, t.UserRole)
		log.Infof("set userid to context")
		roles := strings.Split(t.UserRole, ",")
		ctx.Set("userid", int(t.UserID))
		ctx.Set("userrole", roles)
		ctx.Set("token", token)
	}

	ctx.Next()
}

func (c *Interceptor) ApiKeyAuthInterceptor(ctx *gin.Context) {

	/*
		算法：signature = md5(md5(params...+ts)+APPSECRET)
		params:
		   1. query, form: kv对以字典排序后拼接，如k1=v1&k2=v2&k3=v3
			 2. json: json字符串
			 3. 文件字段不参与校验
	*/
	apikey := ctx.GetHeader("X-APIKey")
	if len(apikey) > 0 {
		params, err := c.service.verifyApiKey(apikey)
		if err == nil {
			log.Infof("apikey = %v", params.ApiKey)
			signature := ctx.Request.FormValue("signature")
			timestamp, _ := strconv.Atoi(ctx.Request.FormValue("ts"))
			curTm := time.Now().Unix()
			if int64(timestamp) < curTm-600 || int64(timestamp) > curTm+10 {
				// 时间戳若为10分钟之前或大于当前时间，判定无效
				log.Errorf("invalid timestamp: %v , current timestamp = %v", timestamp, curTm)
				ctx.Next()
				return
			}
			// 至少一个ts、一个signature参数
			// 对query中的参数进行校验
			keys := make([]string, 0, len(ctx.Request.Form))
			for k := range ctx.Request.Form {
				if k != "signature" {
					keys = append(keys, k)
				}
			}
			sort.Strings(keys)
			var paramToEncode string
			for _, k := range keys {
				paramToEncode += fmt.Sprintf("%v=%v&", k, ctx.Request.Form.Get(k))
			}
			if len(paramToEncode) > 0 {
				paramToEncode = paramToEncode[0 : len(paramToEncode)-1]
			}
			s := utils.GetSignature(paramToEncode, params.ApiSecret)
			log.Infof("data: %v\ncalculated signature: %v", paramToEncode, s)
			if s == signature {
				// 检验通过
				ctx.Set("apikey", params.ApiKey)
				ctx.Set("userid", int(params.UserID))
				roles := strings.Split(params.UserRole, ",")
				ctx.Set("userrole", roles)
			}
			// }
		} else {
			log.Errorf("invalid apikey")
		}
	}
	ctx.Next()
}

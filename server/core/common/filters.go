package common

import (
	"github.com/GoROSEN/rosen-apiserver/core/utils"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/martian/log"
	uuid "github.com/satori/go.uuid"
)

// UserFilter 用户登录过滤器
func UserFilter(ctx *gin.Context) {

	uid := ctx.GetInt("userid")
	if uid == 0 {
		log.Infof("UserFilter user login is required")
		utils.AbortResponse(ctx, 50018, "login is required")
		return
	}
	ctx.Next()
}

// AdminFilter 管理员登录过滤器
func AdminFilter(ctx *gin.Context) {

	uid := ctx.GetInt("userid")
	if uid == 0 {
		log.Infof("UserFilter user login is required")
		utils.AbortResponse(ctx, 50018, "login is required")
		return
	}
	roles, _ := ctx.Get("userrole")
	if roles == nil {
		log.Infof("AdminFilter user roles is empty")
		utils.AbortFailureResponse(ctx, 403, "permission denied")
		return
	}
	for _, t := range roles.([]string) {
		if t == "*" || t == "admin" {
			ctx.Next()
			return
		}
	}
	log.Infof("AdminFilter user roles is not admin")
	utils.AbortFailureResponse(ctx, 403, "permission denied")
}

// UserPermissionFilter 权限过滤器
func UserPermissionFilter(perm string) func(ctx *gin.Context) {

	return func(ctx *gin.Context) {
		uid := ctx.GetInt("userid")
		if uid == 0 {
			log.Infof("UserFilter user login is required")
			utils.AbortResponse(ctx, 50018, "login is required")
			return
		}
		roles, _ := ctx.Get("userrole")
		if roles == nil {
			log.Infof("PermissionFilter user roles is empty")
			utils.AbortFailureResponse(ctx, 403, "permission denied")
			return
		}
		for _, t := range roles.([]string) {
			if t == "*" || t == perm {
				ctx.Next()
				return
			}
		}
		log.Infof("AdminFilter user permission %v is not granted", perm)
		utils.AbortFailureResponse(ctx, 403, "permission denied")
		return
	}
}

// ClientIDFilter 生成客户端ID信息
func ClientIDFilter(ctx *gin.Context) {

	session := sessions.Default(ctx)

	clientUA := ctx.GetHeader("User-Agent")
	var clientID string
	// handle user id
	scid := session.Get("client_id")
	if scid == nil {
		clientID = uuid.NewV4().String()
		session.Set("client_id", clientID)
		if err := session.Save(); err != nil {
			log.Errorf("save user session failed: %v", err)
		}
	} else {
		clientID = scid.(string)
	}

	ctx.Set("ClientUA", clientUA)
	ctx.Set("ClientID", clientID)

}

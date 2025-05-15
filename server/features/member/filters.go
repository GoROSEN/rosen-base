package member

import (
	"github.com/GoROSEN/rosen-apiserver/core/utils"
	"github.com/gin-gonic/gin"
)

// MemberFilter 会员登录过滤器
func MemberFilter(ctx *gin.Context) {

	uid := ctx.GetInt("member-id")
	if uid == 0 {
		utils.AbortResponse(ctx, 50018, "login is required")
		return
	}
	ctx.Next()
}

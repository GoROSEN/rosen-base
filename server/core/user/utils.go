package user

import "github.com/gin-gonic/gin"

func isAdmin(ctx *gin.Context) bool {
	role, _ := ctx.Get("userrole")
	if role == nil {
		return false
	}
	roles := role.([]string)
	for _, t := range roles {
		if t == "admin" {
			return true
		}
	}
	return false
}

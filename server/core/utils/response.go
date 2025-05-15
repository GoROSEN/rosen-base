package utils

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SendResponse 发送响应
func SendResponse(ctx *gin.Context, code int, msg string, data interface{}) {

	if data == nil {
		ctx.JSON(http.StatusOK, gin.H{"code": code, "msg": msg})
	} else {
		ctx.JSON(http.StatusOK, gin.H{"code": code, "msg": msg, "data": data})
	}
}

// SendFailureResponse 发送失败响应
func SendFailureResponse(ctx *gin.Context, code int, msg string) {

	SendResponse(ctx, code, msg, nil)
}

// SendSuccessResponse 发送成功响应
func SendSuccessResponse(ctx *gin.Context, data interface{}) {
	SendResponse(ctx, 20000, "OK", data)
}

// SendSuccessMsgResponse 发送成功响应
func SendSuccessMsgResponse(ctx *gin.Context, msg string, data interface{}) {
	SendResponse(ctx, 20000, msg, data)
}

// SendPagedSuccessResponse 发送成功响应
func SendPagedSuccessResponse(ctx *gin.Context, page, pageSize int, count int64, items interface{}) {
	pageCount := count / int64(pageSize)
	if count%int64(pageSize) > 0 {
		pageCount++
	} else if pageCount == 0 {
		pageCount = 1
	}
	if items == nil {
		SendResponse(ctx, 20000, "OK", gin.H{
			"pager": gin.H{"page": page, "pageCount": pageCount, "pageSize": pageSize},
			"items": []any{}})
	} else {
		SendResponse(ctx, 20000, "OK", gin.H{
			"pager": gin.H{"page": page, "pageCount": pageCount, "pageSize": pageSize},
			"items": items})
	}
}

// SendSimpleSuccessResponse 发送无数据成功响应
func SendSimpleSuccessResponse(ctx *gin.Context) {
	SendResponse(ctx, 20000, "OK", nil)
}

// AbortFailureResponse 中止响应
func AbortFailureResponse(ctx *gin.Context, code int, msg string) {
	ctx.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"code": code, "msg": msg})
}

// AbortResponse 中止响应
func AbortResponse(ctx *gin.Context, code int, msg string) {
	ctx.AbortWithStatusJSON(http.StatusOK, gin.H{"code": code, "msg": msg})
}

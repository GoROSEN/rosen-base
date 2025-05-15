package utils

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

func GetParamInt(ctx *gin.Context, key string, defaultVal int) int {
	val, err := strconv.Atoi(ctx.Param(key))
	if err != nil {
		return defaultVal
	}
	return val
}

func GetParamFloat(ctx *gin.Context, key string, defaultVal float64) float64 {
	val, err := strconv.ParseFloat(ctx.Param(key), 64)
	if err != nil {
		return defaultVal
	}
	return val
}

func GetParam(ctx *gin.Context, key string, defaultVal string) string {
	val := ctx.Param(key)
	if len(val) == 0 {
		val = defaultVal
	}
	return val
}

func GetQueryInt(ctx *gin.Context, key string, defaultVal int) int {
	val, err := strconv.Atoi(ctx.Query(key))
	if err != nil {
		return defaultVal
	}
	return val
}

func GetQueryFloat(ctx *gin.Context, key string, defaultVal float64) float64 {
	val, err := strconv.ParseFloat(ctx.Query(key), 64)
	if err != nil {
		return defaultVal
	}
	return val
}

func GetQuery(ctx *gin.Context, key string, defaultVal string) string {
	val := ctx.Query(key)
	if len(val) == 0 {
		val = defaultVal
	}
	return val
}

func GetFormInt(ctx *gin.Context, key string, defaultVal int) int {
	val, err := strconv.Atoi(ctx.Request.FormValue(key))
	if err != nil {
		return defaultVal
	}
	return val
}

func GetFormFloat(ctx *gin.Context, key string, defaultVal float64) float64 {
	val, err := strconv.ParseFloat(ctx.Request.FormValue(key), 64)
	if err != nil {
		return defaultVal
	}
	return val
}

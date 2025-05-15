package common

import (
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"
)

// SetupCrudV3 快速生成如下CRUD接口：
// - GET /<modelName>/page/:page
// - GET /<modelName>/list
// - GET /<modelName>/get/:id
// - POST /<modelName>/save
// - POST /<modelName>/delete/:id
// 注意：model类必须是新款DM和VM二合一的那种。
func SetupCrudV3[T any](r *gin.RouterGroup, modelName string, crud *CrudController, preloads []string, joins []string, filterFuncs ...func(*gin.Context)) {

	for _, fn := range filterFuncs {
		if fn != nil {
			r.Use(fn)
		}
	}

	r.GET(fmt.Sprintf("/%v/page/:page", modelName), func(ctx *gin.Context) {
		var obj T
		orderby := ctx.Query("sort")
		if len(orderby) == 0 {
			orderby = "id desc"
		} else {
			if orderby[0] == '-' {
				orderby = fmt.Sprintf("%v desc", orderby[1:])
			} else {
				orderby = strings.TrimPrefix(orderby, "+")
			}
		}
		crud.PageFullV2(ctx, &[]T{}, &obj, preloads, joins, orderby, "")
	})

	r.GET(fmt.Sprintf("/%v/list", modelName), func(ctx *gin.Context) {
		orderby := ctx.Query("sort")
		if len(orderby) == 0 {
			orderby = "id desc"
		} else {
			if orderby[0] == '-' {
				orderby = fmt.Sprintf("%v desc", orderby[1:])
			} else {
				orderby = strings.TrimPrefix(orderby, "+")
			}
		}
		crud.ListAllV2(ctx, &[]T{}, preloads, joins, orderby, "")
	})

	r.GET(fmt.Sprintf("/%v/get/:id", modelName), func(ctx *gin.Context) {
		var obj T
		crud.GetModelV2(ctx, &obj, preloads)
	})

	r.POST(fmt.Sprintf("/%v/save", modelName), func(ctx *gin.Context) {
		var obj T
		crud.SaveModelV2(ctx, &obj)
	})

	r.POST(fmt.Sprintf("/%v/delete/:id", modelName), func(ctx *gin.Context) {
		var obj T
		crud.DeleteModelByID(ctx, &obj)
	})
}

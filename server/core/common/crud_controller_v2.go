package common

import (
	"reflect"
	"strconv"

	"github.com/GoROSEN/rosen-apiserver/core/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/martian/log"
)

// V2使用数据模型和视图模型统一方式，不需要copier额外复制

// Page 控制器分页逻辑
func (c *CrudController) PageV2(ctx *gin.Context, list interface{}, obj interface{}, orderby string) {

	c.PreloadPageV2(ctx, list, obj, []string{}, orderby)
}

// PageWhere 控制器分页逻辑
func (c *CrudController) PageWhereV2(ctx *gin.Context, list interface{}, obj interface{}, orderby string, where string, params ...interface{}) {

	c.PageFullV2(ctx, list, obj, nil, nil, orderby, where, params...)
}

// PreloadPage 控制器分页逻辑
func (c *CrudController) PreloadPageV2(ctx *gin.Context, list interface{}, obj interface{}, preloads []string, orderby string) {
	c.PageFullV2(ctx, list, obj, preloads, nil, orderby, "")
}

func (c *CrudController) JoinsPageV2(ctx *gin.Context, list interface{}, obj interface{}, joins []string, orderby string) {
	c.PageFullV2(ctx, list, obj, nil, joins, orderby, "")
}

// PageFull 控制器分页逻辑
func (c *CrudController) PageFullV2(ctx *gin.Context, list interface{}, obj interface{}, preloads []string, joins []string, orderby, where string, params ...interface{}) {

	page, err := strconv.Atoi(ctx.Param("page"))
	if err != nil || page < 1 {
		if page, err = strconv.Atoi(ctx.Query("page")); err != nil || page < 1 {
			page = 1
		}
	}
	pageSize, err := strconv.Atoi(ctx.Query("pageSize"))
	if err != nil || pageSize <= 0 {
		pageSize = 10
	}

	count, err := c.Crud.PagedV2(list, obj, preloads, joins, orderby, page-1, pageSize, where, params...)
	if err != nil {
		utils.SendFailureResponse(ctx, 500, "message.common.system-error")
		return
	}

	utils.SendPagedSuccessResponse(ctx, page, pageSize, count, list)
}

// ListAll 获取所有模型
func (c *CrudController) ListAllV2(ctx *gin.Context, list interface{}, preloads []string, joins []string, orderby, where string, params ...interface{}) {

	if err := c.Crud.ListV2(list, preloads, joins, orderby, where, params...); err != nil {
		utils.SendFailureResponse(ctx, 500, "message.common.system-error")
		return
	}
	utils.SendSuccessResponse(ctx, list)

}

// GetModel 获取模型数据, id来自params
func (c *CrudController) GetModelV2(ctx *gin.Context, obj interface{}, preloads []string) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id < 1 {
		utils.SendFailureResponse(ctx, 400, "message.common.system-error")
		return
	}
	if len(preloads) == 0 {
		if err := c.Crud.GetModelByID(obj, uint(id)); err != nil {
			log.Errorf("get model error: %v", err)
			utils.SendFailureResponse(ctx, 500, "message.common.system-error")
			return
		}
	} else {
		if err := c.Crud.GetPreloadModelByID(obj, uint(id), preloads); err != nil {
			log.Errorf("get model error: %v", err)
			utils.SendFailureResponse(ctx, 500, "message.common.system-error")
			return
		}
	}
	utils.SendSuccessResponse(ctx, obj)
}

// SaveModel 保存模型数据
func (c *CrudController) SaveModelV2(ctx *gin.Context, obj interface{}) {
	c.SaveModelOptV2(ctx, obj, nil, nil)
}

func (c *CrudController) CreateOrUpdateModelV2(ctx *gin.Context, obj interface{}, selects []string, omits []string, createOmits []string) {

	if err := ctx.Bind(obj); err != nil {
		log.Errorf("cannot bind vo: %v", err)
		utils.SendFailureResponse(ctx, 400, "message.common.system-error")
		return
	}
	val := reflect.ValueOf(obj).Elem()
	objId := val.FieldByName("ID").Interface().(uint)
	if objId == 0 {
		omits2 := omits
		if createOmits != nil {
			omits2 = createOmits
		}
		if err := c.Crud.CreateModelOpt(obj, selects, omits2); err != nil {
			log.Errorf("save model error: %v", err)
			utils.SendFailureResponse(ctx, 500, "message.common.system-error")
			return
		}
	} else {
		if err := c.Crud.UpdateModel(obj, selects, omits); err != nil {
			log.Errorf("save model error: %v", err)
			utils.SendFailureResponse(ctx, 500, "message.common.system-error")
			return
		}
	}
	utils.SendSimpleSuccessResponse(ctx)
}

func (c *CrudController) UpdateModelV2(ctx *gin.Context, obj interface{}, selects []string, omits []string) {

	if err := ctx.Bind(obj); err != nil {
		log.Errorf("cannot bind vo: %v", err)
		utils.SendFailureResponse(ctx, 400, "message.common.system-error")
		return
	}
	val := reflect.ValueOf(obj).Elem()
	objId := val.FieldByName("ID").Interface().(uint)
	if objId == 0 {
		log.Errorf("cannot find object")
		utils.SendFailureResponse(ctx, 500, "message.common.system-error")
		return
	} else {
		if err := c.Crud.UpdateModel(obj, selects, omits); err != nil {
			log.Errorf("save model error: %v", err)
			utils.SendFailureResponse(ctx, 500, "message.common.system-error")
			return
		}
	}
	utils.SendSimpleSuccessResponse(ctx)
}

// SaveModelOptV2 保存模型数据
func (c *CrudController) SaveModelOptV2(ctx *gin.Context, obj interface{}, selects []string, omits []string) {

	if err := ctx.Bind(obj); err != nil {
		log.Errorf("could not bind vo: %v", err)
		utils.SendFailureResponse(ctx, 400, "message.common.system-error")
		return
	}
	if selects == nil {
		selects = []string{"*"}
	}
	log.Infof("obj: %v", obj)
	val := reflect.ValueOf(obj).Elem()
	objId := val.FieldByName("ID").Interface().(uint)
	if objId == 0 {
		if omits == nil {
			omits = []string{}
		}
		if err := c.Crud.SaveModelOpt(obj, selects, omits); err != nil {
			log.Errorf("save model error: %v", err)
			utils.SendFailureResponse(ctx, 500, "message.common.system-error")
			return
		}
	} else {
		if omits == nil {
			omits = []string{"created_at"}
		} else {
			omits = append(omits, "created_at")
		}
		if err := c.Crud.UpdateModel(obj, selects, omits); err != nil {
			log.Errorf("save model error: %v", err)
			utils.SendFailureResponse(ctx, 500, "message.common.system-error")
			return
		}
	}
	utils.SendSimpleSuccessResponse(ctx)
}

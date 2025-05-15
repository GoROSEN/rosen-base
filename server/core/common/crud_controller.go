package common

import (
	"reflect"
	"strconv"

	"github.com/GoROSEN/rosen-apiserver/core/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/martian/log"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
)

// CrudController Crud控制器
type CrudController struct {
	Crud *CrudService
}

// SetupCrud 初始化Crud
func (c *CrudController) SetupCrud(db *gorm.DB) {
	c.Crud = NewCrudService(db)
}

// Page 控制器分页逻辑
func (c *CrudController) Page(ctx *gin.Context, list interface{}, volist interface{}, obj interface{}, orderby string) {

	c.PageCvt(ctx, list, volist, obj, nil, orderby)
}

// PageWhere 控制器分页逻辑
func (c *CrudController) PageWhere(ctx *gin.Context, list interface{}, volist interface{}, obj interface{}, orderby string, where string, params ...interface{}) {

	c.PageFull(ctx, list, volist, obj, nil, nil, nil, orderby, where, params...)
}

// PageCvt 带自定义转换的分页列表
func (c *CrudController) PageCvt(ctx *gin.Context, list interface{}, volist interface{}, obj interface{}, cvt func(interface{}) interface{}, orderby string) {

	c.PreloadPageCvt(ctx, list, volist, obj, []string{}, cvt, orderby)
}

// PreloadPage 控制器分页逻辑
func (c *CrudController) PreloadPage(ctx *gin.Context, list interface{}, volist interface{}, obj interface{}, preloads []string, orderby string) {
	c.PreloadPageCvt(ctx, list, volist, obj, preloads, nil, orderby)
}

func (c *CrudController) JoinsPage(ctx *gin.Context, list interface{}, volist interface{}, obj interface{}, joins []string, orderby string) {
	c.JoinsPageCvt(ctx, list, volist, obj, joins, nil, orderby)
}

// PreloadPageCvt 自定义转换器控制器分页逻辑
func (c *CrudController) PreloadPageCvt(ctx *gin.Context, list interface{}, volist interface{}, obj interface{}, preloads []string, cvt func(interface{}) interface{}, orderby string) {
	c.PageFull(ctx, list, volist, obj, preloads, nil, cvt, orderby, "")
}

func (c *CrudController) JoinsPageCvt(ctx *gin.Context, list interface{}, volist interface{}, obj interface{}, joins []string, cvt func(interface{}) interface{}, orderby string) {
	c.PageFull(ctx, list, volist, obj, nil, joins, cvt, orderby, "")
}

// PageFull 控制器分页逻辑
func (c *CrudController) PageFull(ctx *gin.Context, list interface{}, volist interface{}, obj interface{}, preloads []string, joins []string, cvt func(interface{}) interface{}, orderby, where string, params ...interface{}) {

	c.PageFullV1_1(ctx, list, volist, obj, preloads, joins, nil, cvt, orderby, where, params...)
}

func (c *CrudController) PageFullV1_1(ctx *gin.Context, list interface{}, volist interface{}, obj interface{}, preloads []string, joins []string, innerJoins []string, cvt func(interface{}) interface{}, orderby, where string, params ...interface{}) {

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

	count, err := c.Crud.PagedV3(list, obj, preloads, joins, innerJoins, orderby, page-1, pageSize, where, params...)
	if err != nil {
		utils.SendFailureResponse(ctx, 500, "message.common.system-error")
		return
	}
	if volist != nil && volist != list {
		if cvt != nil {
			volist = cvt(list)
		} else {
			copier.Copy(volist, list)
		}
		utils.SendPagedSuccessResponse(ctx, page, pageSize, count, volist)
	} else {
		utils.SendPagedSuccessResponse(ctx, page, pageSize, count, list)
	}
}

// ListAll 获取所有模型
func (c *CrudController) ListAll(ctx *gin.Context, list interface{}, volist interface{}, obj interface{}, preloads []string, joins []string, cvt func(interface{}) interface{}, orderby, where string, params ...interface{}) {

	c.ListAllV1_1(ctx, list, volist, obj, preloads, joins, nil, cvt, orderby, where, params...)
}

// ListAll 获取所有模型
func (c *CrudController) ListAllV1_1(ctx *gin.Context, list interface{}, volist interface{}, obj interface{}, preloads []string, joins []string, innerJoins []string, cvt func(interface{}) interface{}, orderby, where string, params ...interface{}) {

	if err := c.Crud.ListV2_1(list, preloads, joins, innerJoins, orderby, where, params...); err != nil {
		utils.SendFailureResponse(ctx, 500, "message.common.system-error")
		return
	}
	if cvt != nil {
		volist = cvt(list)
	} else {
		copier.Copy(volist, list)
	}
	utils.SendSuccessResponse(ctx, volist)

}

// ListTop 获取前N个模型
func (c *CrudController) ListTop(ctx *gin.Context, count int, list interface{}, volist interface{}, obj interface{}, preloads []string, joins []string, cvt func(interface{}) interface{}, orderby, where string, params ...interface{}) {

	if err := c.Crud.ListV3(list, count, preloads, joins, orderby, where, params...); err != nil {
		utils.SendFailureResponse(ctx, 500, "message.common.system-error")
		return
	}
	if volist != nil && volist != list {
		if cvt != nil {
			volist = cvt(list)
		} else {
			copier.Copy(volist, list)
		}
		utils.SendSuccessResponse(ctx, volist)
	} else {
		utils.SendSuccessResponse(ctx, list)
	}

}

// GetModel 获取模型数据, id来自params
func (c *CrudController) GetModel(ctx *gin.Context, obj interface{}, vo interface{}, preloads []string) {
	c.GetModelCvt(ctx, obj, vo, preloads, nil)
}

// GetModelCvt 获取模型数据, id来自params
func (c *CrudController) GetModelCvt(ctx *gin.Context, obj interface{}, vo interface{}, preloads []string, cvt func(interface{}) interface{}) {
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
	if vo != nil && vo != obj {
		if cvt != nil {
			vo = cvt(obj)
		} else {
			if err := copier.Copy(vo, obj); err != nil {
				log.Errorf("copy to vo error: %v", err)
				utils.SendFailureResponse(ctx, 500, "message.common.system-error")
				return
			}
		}
		utils.SendSuccessResponse(ctx, vo)
	} else {
		utils.SendSuccessResponse(ctx, obj)
	}
}

// SaveModel 保存模型数据
func (c *CrudController) SaveModel(ctx *gin.Context, vo interface{}, obj interface{}) {
	c.SaveModelCvt(ctx, vo, obj, nil)
}

func (c *CrudController) CreateOrUpdateModel(ctx *gin.Context, vo interface{}, obj interface{}, selects []string, omits []string, createOmits []string) {

	if err := ctx.ShouldBind(vo); err != nil {
		log.Errorf("cannot bind vo: %v", err)
		utils.SendFailureResponse(ctx, 400, "message.common.system-error")
		return
	}
	if vo != obj {
		if err := copier.Copy(obj, vo); err != nil {
			log.Errorf("copy to model error: %v", err)
			utils.SendFailureResponse(ctx, 500, "message.common.system-error")
			return
		}
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

func (c *CrudController) CreateOrUpdateModelCvt(ctx *gin.Context, vo interface{}, obj interface{}, selects []string, omits []string, createOmits []string, cvt func(interface{}) interface{}) {

	if err := ctx.ShouldBind(vo); err != nil {
		log.Errorf("cannot bind vo: %v", err)
		utils.SendFailureResponse(ctx, 400, "message.common.system-error")
		return
	}
	if vo != obj {
		log.Infof("vo: %v", vo)
		obj = cvt(vo)
		log.Infof("obj: %v", obj)
	}
	val := reflect.ValueOf(obj).Elem()
	objId := val.FieldByName("ID").Interface().(uint)
	if objId == 0 {
		omits2 := omits
		if createOmits != nil {
			omits2 = createOmits
		}
		log.Infof("create with selects: %v", selects)
		if err := c.Crud.CreateModelOpt(obj, selects, omits2); err != nil {
			log.Errorf("save model error: %v", err)
			utils.SendFailureResponse(ctx, 500, "message.common.system-error")
			return
		}
	} else {
		log.Infof("update with selects: %v", selects)
		if err := c.Crud.UpdateModel(obj, selects, omits); err != nil {
			log.Errorf("save model error: %v", err)
			utils.SendFailureResponse(ctx, 500, "message.common.system-error")
			return
		}
	}
	utils.SendSimpleSuccessResponse(ctx)
}

func (c *CrudController) UpdateModel(ctx *gin.Context, vo interface{}, obj interface{}, selects []string, omits []string) {

	if err := ctx.ShouldBind(vo); err != nil {
		log.Errorf("cannot bind vo: %v", err)
		utils.SendFailureResponse(ctx, 400, "message.common.system-error")
		return
	}
	if vo != obj {
		if err := copier.Copy(obj, vo); err != nil {
			log.Errorf("copy to model error: %v", err)
			utils.SendFailureResponse(ctx, 500, "message.common.system-error")
			return
		}
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

// SaveModelCvt 保存模型数据
func (c *CrudController) SaveModelCvt(ctx *gin.Context, vo interface{}, obj interface{}, cvt func(interface{}) interface{}) {

	if vo == nil {
		vo = obj
	}
	if err := ctx.ShouldBind(vo); err != nil {
		log.Errorf("could not bind vo: %v", err)
		utils.SendFailureResponse(ctx, 400, "message.common.system-error")
		return
	}
	if vo != obj {
		log.Infof("vo: %v", vo)
		if cvt != nil {
			obj = cvt(vo)
		} else {
			if err := copier.Copy(obj, vo); err != nil {
				log.Errorf("copy to model error: %v", err)
				utils.SendFailureResponse(ctx, 500, "message.common.system-error")
				return
			}
		}
	}
	log.Infof("obj: %v", obj)
	val := reflect.ValueOf(obj).Elem()
	objId := val.FieldByName("ID").Interface().(uint)
	if objId == 0 {
		if err := c.Crud.SaveModel(obj); err != nil {
			log.Errorf("save model error: %v", err)
			utils.SendFailureResponse(ctx, 500, "message.common.system-error")
			return
		}
	} else {
		if err := c.Crud.UpdateModel(obj, []string{"*"}, []string{"created_at"}); err != nil {
			log.Errorf("save model error: %v", err)
			utils.SendFailureResponse(ctx, 500, "message.common.system-error")
			return
		}
	}
	utils.SendSimpleSuccessResponse(ctx)
}

// DeleteModelByID 删除模型
func (c *CrudController) DeleteModelByID(ctx *gin.Context, obj interface{}) {

	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil || id < 1 {
		utils.SendFailureResponse(ctx, 400, "message.common.system-error")
		return
	}
	val := reflect.ValueOf(obj).Elem()
	val.FieldByName("ID").SetUint(uint64(id))
	if err := c.Crud.DeleteModel(obj); err != nil {
		log.Errorf("delete model error: %v", err)
		utils.SendFailureResponse(ctx, 500, "message.common.system-error")
		return
	}
	utils.SendSimpleSuccessResponse(ctx)
}

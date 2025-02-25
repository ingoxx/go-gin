package operatelog

import (
	"net/http"

	"github.com/ingoxx/go-gin/project/logic/operate"
	"github.com/ingoxx/go-gin/project/model"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func LogListController(ctx *gin.Context) {
	var od operate.OperateLogListQuery
	var u model.OperateLogModel
	if err := ctx.ShouldBindQuery(&od); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"code":    10001,
		})
		return
	}

	validate := validator.New()
	vd := NewValidateData(validate)
	if err := od.PaginateLogic(u, vd); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"code":    10002,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":     od.PageData.ModelSlice,
		"total":    od.PageData.Total,
		"pageSize": od.PageData.PageSize,
		"code":     10000,
	})
}

func GetLoginNumDataController(ctx *gin.Context) {
	var op model.OperateLogModel
	data, err := op.GetLoginNum()
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"code":    10002,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": data,
		"code": 10000,
	})

}

func GetLinuxCmdDataController(ctx *gin.Context) {
	var op model.OperateLogModel
	data, err := op.GetRunLinuxCmdNum()
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"code":    10002,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": data,
		"code": 10000,
	})

}

func GetUserLoginNumController(ctx *gin.Context) {
	var op model.OperateLogModel
	data, err := op.GetUserLoginNum()
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"code":    10002,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": data,
		"code": 10000,
	})

}

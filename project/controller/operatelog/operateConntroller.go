package operatelog

import (
	"net/http"

	"github.com/Lxb921006/Gin-bms/project/logic/operate"
	"github.com/Lxb921006/Gin-bms/project/model"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func OperateLogList(ctx *gin.Context) {
	var od operate.OperateLogListQuery
	var u model.OperateLog
	if err := ctx.ShouldBindQuery(&od); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"code":    20001,
		})
		return
	}

	validate := validator.New()
	vd := NewValidateData(validate)
	if err := od.PaginateLogic(u, vd); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"code":    20002,
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

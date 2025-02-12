package assets

import (
	"fmt"
	"github.com/Lxb921006/Gin-bms/project/logger"
	"github.com/Lxb921006/Gin-bms/project/logic/assets"
	"github.com/Lxb921006/Gin-bms/project/model"
	"github.com/Lxb921006/Gin-bms/project/service"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/mitchellh/mapstructure"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

var (
	upGrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

// RunProgramController 程序更新-在新的页面, 查看系统日志, 执行linux命令
func RunProgramController(ctx *gin.Context) {
	conn, err := upGrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		return
	}

	defer conn.Close()

	var aprm model.AssetsProgramUpdateRecordModel
	if err = service.NewWs(conn, &aprm).Run(); err != nil {
		return
	}
}

// SyncFileController 分发文件
func SyncFileController(ctx *gin.Context) {
	conn, err := upGrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		return
	}

	defer conn.Close()

	if err = service.NewSendFileWs(conn).Send(); err != nil {
		return
	}
}

// RunProgramApiController 程序更新-当前页面
func RunProgramApiController(ctx *gin.Context) {
	var ps RunProgramApiForm
	if err := ps.Run(ctx); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"code":    10001,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "提交成功",
		"code":    10000,
	})

}

// GetMissionStatusController 废弃
func GetMissionStatusController(ctx *gin.Context) {
	var ps GetMissionStatusForm
	data, err := ps.GetProgress(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"code":    10001,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":    data,
		"message": "ok",
		"code":    10000,
	})
}

func CreateUpdateProgramRecordController(ctx *gin.Context) {
	var create CreateUpdateProgramRecordForm

	if err := create.Create(ctx); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"code":    10002,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "更新已提交",
		"code":    10000,
	})
}

func ProgramUpdateListController(ctx *gin.Context) {
	var apul ProgramUpdateListForm
	data, err := apul.List(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"code":    10001,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":     data.ModelSlice,
		"total":    data.Total,
		"pageSize": data.PageSize,
		"code":     10000,
	})
}

func UploadController(ctx *gin.Context) {
	auf := NewUploadForm()
	data, err := auf.UploadFiles(ctx)
	if err != nil {
		ctx.SecureJSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"code":    10001,
		})
	} else {
		ctx.SecureJSON(http.StatusOK, gin.H{
			"message": "upload ok",
			"data":    data,
			"code":    10000,
		})
	}
}

// ListController 服务器列表
func ListController(ctx *gin.Context) {
	var alc ListForm
	data, err := alc.List(ctx)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"code":    10001,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data":     data.ModelSlice,
		"total":    data.Total,
		"pageSize": data.PageSize,
		"config":   NewProgramConfig(),
		"code":     10000,
		"message":  "ok",
	})
}

// CreateController 创建服务器
func CreateController(ctx *gin.Context) {
	var nca = NewCreateUpdateAssetsForm(ctx)
	if err := nca.VerifyFrom(); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"code":    10001,
		})
		return
	}

	if err := nca.Create(); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("%s 创建失败, errMsg: %v", nca.Ip, err.Error()),
			"code":    10002,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("%s 创建成功", nca.Ip),
		"code":    10000,
	})

	return

}

// UpdateController 更新服务器信息
func UpdateController(ctx *gin.Context) {
	var nca = NewCreateUpdateAssetsForm(ctx)
	if err := nca.VerifyFrom(); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"code":    10001,
		})
		return
	}

	if err := nca.Update(); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("%s 更新失败, errMsg: %v", nca.Ip, err.Error()),
			"code":    10002,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("%s 更新成功", nca.Ip),
		"code":    10000,
	})

	return
}

// DeleteController 删除服务器
func DeleteController(ctx *gin.Context) {
	var adf DelForm
	if err := adf.Del(ctx); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"code":    10001,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("%v 删除成功", adf.Ips),
		"code":    10000,
	})
}

func AddProgramOperateController(ctx *gin.Context) {
	var pf ProgramAddForm
	var adp model.AssetsProgramModel

	if err := ctx.ShouldBind(&pf); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"code":    10001,
		})
		return
	}

	if err := mapstructure.Decode(pf, &adp); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("添加%v失败, errMsg: %v", pf.CnName, err.Error()),
			"code":    10002,
		})
		return
	}

	if err := adp.Create(adp); err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("添加%v失败, errMsg: %v", pf.CnName, err.Error()),
			"code":    10003,
		})
		return
	}

	data, err := assets.NewProgramOperate().ProgramData(adp)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("获取程序操作列表失败, errMsg: %v", err.Error()),
			"code":    10004,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": fmt.Sprintf("添加%v成功", pf.CnName),
		"code":    10000,
		"data":    data,
	})
}

func ProgramListController(ctx *gin.Context) {
	var pf ProgramListForm
	var adp model.AssetsProgramModel

	if err := ctx.ShouldBind(&pf); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"code":    10001,
		})
		return
	}

	data, err := assets.NewProgramOperate().ProgramData(adp)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": fmt.Sprintf("获取程序操作列表失败, errMsg: %v", err.Error()),
			"code":    10004,
		})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"message": "获取程序操作列表成功",
		"code":    10000,
		"data":    data,
	})
}

// WebTerminalControllerOut 废弃
func WebTerminalControllerOut(ctx *gin.Context) {
	var wtq WebTerminalQuery
	var am model.AssetsModel
	if err := ctx.ShouldBind(&wtq); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"code":    10001,
		})
		return
	}

	ip, err := am.GetTerminalIp(wtq.ID)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"code":    10002,
		})
		return
	}

	// 代理请求
	terminalUrl := fmt.Sprintf("http://%s:17600", ip)
	target, err := url.Parse(terminalUrl)
	if err != nil {
		ctx.JSON(http.StatusOK, gin.H{
			"message": err.Error(),
			"code":    10003,
		})
		return
	}

	proxy := httputil.NewSingleHostReverseProxy(target)
	// 修改请求路径，去除/terminal前缀
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.URL.Path = strings.TrimPrefix(req.URL.Path, "/assets/terminal")
		if req.URL.Path == "" {
			req.URL.Path = "/"
		}
		// 可选：设置正确的Host和Header
		req.Host = target.Host
		req.Header.Set("X-Forwarded-Host", req.Host)
	}

	// 处理WebSocket升级请求
	proxy.ModifyResponse = func(resp *http.Response) error {
		if resp.StatusCode == http.StatusSwitchingProtocols {
			return nil
		}
		return nil
	}

	proxy.ServeHTTP(ctx.Writer, ctx.Request)
}

// WebTerminalController 终端连接
func WebTerminalController(ctx *gin.Context) {
	conn, err := upGrader.Upgrade(ctx.Writer, ctx.Request, nil)
	if err != nil {
		return
	}

	defer conn.Close()

	serverIp := ctx.Query("ip")
	remoteIp := ctx.RemoteIP()
	operator := ctx.Query("user")
	if err := NewWebTerminal(conn, operator, remoteIp, serverIp).Ssh(); err != nil {
		logger.Error(fmt.Sprintf("failed to ssh %s", serverIp))
	}

}

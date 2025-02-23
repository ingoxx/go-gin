package cluster

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
)

func (rc *GenericClusterJson) Delete(ctx *gin.Context) error {
	if err := ctx.ShouldBindJSON(rc); err != nil {
		return err
	}

	var errs = make([]error, 0)
	var si ServerNodeInput
	for _, id := range rc.ID {
		cluster, err := rc.gs.cm.GetCluster(id)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		rc.clusterName = append(rc.clusterName, cluster.Name)

		for _, ip := range cluster.Servers {
			si.Ip = ip.Ip
			si.NodeType = ip.NodeType
			rc.serversInfo = append(rc.serversInfo, si)
		}

		rc.gs.sw.ServersInfo = rc.serversInfo
		if err := rc.gs.sw.LeaveCluster(); err != nil {
			errs = append(errs, fmt.Errorf("%v", err))
			continue
		}

		if err := rc.gs.sw.UpdateServers(); err != nil {
			errs = append(errs, fmt.Errorf("%v", err))
			continue
		}

		if err := rc.gs.cm.Delete(id); err != nil {
			errs = append(errs, err)
			continue
		}
	}

	if len(errs) != 0 {
		return fmt.Errorf("%v", errs)
	}

	return nil
}

func (rc *GenericClusterJson) StartHealthCheck(ctx *gin.Context) error {
	if err := ctx.ShouldBindJSON(rc); err != nil {
		return err
	}

	var errs = make([]error, 0)
	for _, id := range rc.ID {
		cluster, err := rc.gs.cm.GetCluster(id)
		if err != nil {
			errs = append(errs, err)
			continue
		}

		rc.clusterName = append(rc.clusterName, cluster.Name)

		if err := mapstructure.Decode(cluster, &rc.gs.sw); err != nil {
			return fmt.Errorf("集群: [%s] 健康检测启动失败, errMsg: %v", cluster.Name, err.Error())
		}

		if err := rc.gs.sw.StartHealthCheck(); err != nil {
			errs = append(errs, err)
			continue
		}

	}

	if len(errs) != 0 {
		return fmt.Errorf("%v", errs)
	}

	return nil
}

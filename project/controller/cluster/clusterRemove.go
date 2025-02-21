package cluster

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

func (rc *DeleteSwarmJson) Delete(ctx *gin.Context) error {
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

		rc.deleteClusterName = append(rc.deleteClusterName, cluster.Name)

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

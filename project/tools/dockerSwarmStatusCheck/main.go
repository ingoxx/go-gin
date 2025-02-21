package dockerSwarmStatusCheck

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/Lxb921006/Gin-bms/project/config"
	"github.com/Lxb921006/Gin-bms/project/tools/ddwarning"
	"log"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	_ "github.com/go-sql-driver/mysql"
)

var clusterStatusInfo = map[string]uint{
	"Ready":   200,
	"Down":    100,
	"manager": 1,
	"worker":  2,
}

type ClusterHealthCheck struct {
	db  *sql.DB
	cli *client.Client
	cid string
	ctx context.Context
}

func NewClusterHealthCheck(cid string, db *sql.DB, cli *client.Client) *ClusterHealthCheck {
	return &ClusterHealthCheck{
		cid: cid,
		db:  db,
		cli: cli,
		ctx: context.Background(),
	}
}

func (chc *ClusterHealthCheck) updateWorkerStatus(ip string, status uint) {
	if status == 100 {
		ddwarning.SendWarning(fmt.Sprintf("集群健康监测告警, worker node failure,  worker ip: %s", ip))
	}

	query := "UPDATE assets_models SET node_status = ?, start = NOW() WHERE ip = ?"
	_, err := chc.db.Exec(query, status, ip)
	if err != nil {
		msg := fmt.Sprintf("集群健康监测告警, failed to connect to database, errMsg: %v\n", err)
		ddwarning.SendWarning(msg)
		log.Printf("Failed to update server status, errMsg:%v\n", err)
	} else {
		log.Printf("Updated status for server: %s %v\n", ip, status)
	}

	return
}

func (chc *ClusterHealthCheck) updateClusterStatus(ip string, status uint) {
	if status == 100 {
		ddwarning.SendWarning(fmt.Sprintf("集群健康监测告警, manager node failure,  manager ip: %s", ip))
	}
	query := "UPDATE cluster_models SET status = ?, date = NOW() WHERE cluster_id = ?"
	_, err := chc.db.Exec(query, status, chc.cid)
	if err != nil {
		msg := fmt.Sprintf("集群健康监测告警, failed to connect to database, errMsg: %v\n", err)
		ddwarning.SendWarning(msg)
		log.Printf("Failed to update cluster status, errMsg: %v\n", err)
	} else {
		fmt.Printf("Cluster status updated to: %v\n", status)
	}

	return
}

func (chc *ClusterHealthCheck) HealthCheck() {
	// **获取所有 Swarm 节点信息**
	nodes, err := chc.cli.NodeList(chc.ctx, types.NodeListOptions{})
	if err != nil {
		log.Printf("failed to list Swarm nodes, errMsg: %v\n", err)
		return
	}

	managerHealthyCount := 0
	managerTotalCount := 0
	var managerIp string
	// 遍历所有节点
	for _, node := range nodes {
		ip := node.Status.Addr
		status := string(node.Status.State) // Ready / Down
		role := string(node.Spec.Role)      // worker / manager

		// 统计 Manager 健康数量
		if role == "manager" {
			managerIp = ip
			managerTotalCount++
			if status == "ready" {
				managerHealthyCount++
			}
		}

		// **更新数据库**
		chc.updateWorkerStatus(ip, clusterStatusInfo["status"])

	}

	// **判断集群是否健康**
	if managerHealthyCount > managerTotalCount/2 {
		chc.updateClusterStatus(managerIp, clusterStatusInfo["status"])
	} else {
		chc.updateClusterStatus(managerIp, clusterStatusInfo["status"])
	}
}

func Check(cid string) {
	// **创建 Docker 客户端**
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		msg := fmt.Sprintf("集群健康监测告警, failed to initialize Docker client, errMsg: %v\n", err)
		ddwarning.SendWarning(msg)
		log.Fatalf(msg)
	}
	defer cli.Close()

	// **连接数据库**
	db, err := sql.Open("mysql", config.MyConAddre)
	if err != nil {
		msg := fmt.Sprintf("集群健康监测告警, failed to connect to database, errMsg: %v\n", err)
		ddwarning.SendWarning(msg)
		log.Fatalf(msg)
	}
	defer db.Close()

	// **定期检查集群健康状态**
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			NewClusterHealthCheck(cid, db, cli).HealthCheck()
		}
	}

	return
}

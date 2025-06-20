package dockerSwarmStatusCheck

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	_ "github.com/go-sql-driver/mysql"
	"github.com/ingoxx/go-gin/project/config"
	"github.com/ingoxx/go-gin/project/tools/ddwarning"
	"log"
	"time"
)

var clusterStatusInfo = map[string]uint{
	"ready":   200,
	"down":    100,
	"manager": 1,
	"worker":  2,
}

// ClusterHealthChecker 结构体
type ClusterHealthChecker struct {
	db    *sql.DB
	cli   *client.Client
	ctx   context.Context // cluster_id
	cid   string
	cache map[string]string
}

func (chc *ClusterHealthChecker) checkClusterExists(managerIp string) bool {
	var exists bool
	query := "SELECT EXISTS(SELECT 1 FROM cluster_models WHERE master_ip = ?)"
	err := chc.db.QueryRow(query, managerIp).Scan(&exists)
	if err != nil {
		return exists
	}

	return exists
}

func (chc *ClusterHealthChecker) getPrimaryManager() (string, error) {
	var primaryManagerIP string
	query := "SELECT master_ip FROM cluster_models WHERE cluster_cid = ?"
	err := chc.db.QueryRow(query, chc.cid).Scan(&primaryManagerIP)
	if err != nil {
		return "", err
	}

	return primaryManagerIP, nil
}

func (chc *ClusterHealthChecker) getPrimaryManagerStatus() (uint, error) {
	var status uint
	query := "SELECT status FROM cluster_models WHERE cluster_cid = ?"
	err := chc.db.QueryRow(query, chc.cid).Scan(&status)
	if err != nil {
		return 0, err
	}

	return status, nil
}

func (chc *ClusterHealthChecker) getWorkerStatus(ip string) (uint, error) {
	var isLeave uint
	query := "SELECT leave_type FROM assets_models WHERE ip = ?"
	err := chc.db.QueryRow(query, ip).Scan(&isLeave)
	if err != nil {
		log.Printf("failed to get worker %s leave type\n", ip)
		return isLeave, err
	}

	return isLeave, nil
}

func (chc *ClusterHealthChecker) getClusterId(managerIp string) (uint, error) {
	var clusterID uint
	query := "SELECT id FROM cluster_models WHERE master_ip = ?"
	err := chc.db.QueryRow(query, managerIp).Scan(&clusterID)
	if err != nil {
		return clusterID, err
	}

	return clusterID, nil
}

func (chc *ClusterHealthChecker) getClusterCid(managerIp string) (string, error) {
	var clusterID string
	query := "SELECT cluster_cid FROM cluster_models WHERE master_ip = ?"
	err := chc.db.QueryRow(query, managerIp).Scan(&clusterID)
	if err != nil {
		return clusterID, err
	}

	if clusterID != "" {
		chc.cid = clusterID
	}

	return clusterID, nil
}

func (chc *ClusterHealthChecker) getSwarmNodes() ([]swarm.Node, error) {
	ctx := context.Background()
	nodes, err := chc.cli.NodeList(ctx, types.NodeListOptions{})
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

func (chc *ClusterHealthChecker) updateServerStatus(ip string, role, status, cid uint) {
	var err error

	if cid == 0 {
		query := "UPDATE assets_models SET node_status = ?, node_type = ?, cluster_id = NULL WHERE ip = ?"
		_, err = chc.db.Exec(query, status, role, ip)
	} else {
		query := "UPDATE assets_models SET node_status = ?, node_type = ?, cluster_id = ? WHERE ip = ?"
		_, err = chc.db.Exec(query, status, role, cid, ip)
	}

	if err != nil {
		log.Printf("❌ Failed to update server status for %s: %v\n", ip, err)
		return
	}
}

func (chc *ClusterHealthChecker) updatePrimaryManager(newPrimaryIP string, status uint) error {
	query := "UPDATE cluster_models SET master_ip = ?, status = ?, date = ? WHERE cluster_cid = ?"
	_, err := chc.db.Exec(query, newPrimaryIP, status, time.Now(), chc.cid)
	if err != nil {
		return err
	}

	return nil
}

func (chc *ClusterHealthChecker) checkClusterHealth(managerIp string) {
	nodes, err := chc.getSwarmNodes()
	if err != nil {
		log.Printf("❌ Failed to get swarm nodes: %v\n", err)
		ddwarning.SendWarning(fmt.Sprintf("无法获取节点信息, 失败信息: '%v', 集群ID: '%s'", err, chc.cid))
		return
	}

	var primaryManagerIP string
	var foundLeader bool

	// 遍历所有 Swarm 节点
	for _, node := range nodes {
		ip := node.Status.Addr
		status := string(node.Status.State)
		role := string(node.Spec.Role)
		nodeID := node.ID

		if status == "ready" {
			// 如果是 Manager，记录健康的管理节点
			if role == "manager" {
				// 记录 Swarm 选出的 Leader
				if node.ManagerStatus != nil && node.ManagerStatus.Leader {
					primaryManagerIP = ip
					foundLeader = true
				}
			}
		}

		id, err := chc.getClusterId(managerIp)
		if err != nil || id == 0 {
			esg := fmt.Sprintf("获取管理节点id失败, 失败信息: '%v', 集群id: '%s'", err, chc.cid)
			chc.sendWarnMsg(esg)
			return
		}

		leaveType, err := chc.getWorkerStatus(ip)
		if err != nil {
			ddwarning.SendWarning(fmt.Sprintf("获取节点状态失败, 失败信息: '%v', 集群id: '%s'", err, chc.cid))
			return
		}

		if leaveType == 1 {
			if err := chc.NodeRemove(nodeID); err != nil {
				esg := fmt.Sprintf("移除'%s'失败，失败信息：'%v', 所在集群id：'%s'", ip, err, chc.cid)
				chc.sendWarnMsg(esg)
				return
			}
			chc.updateServerStatus(ip, 3, 300, 0)
		} else {
			if clusterStatusInfo[status] == 100 {
				chc.cache[ip] = ip
				esg := fmt.Sprintf("节点'%s'发生故障, 当前状态: '%s', 故障信息信息: '%v', 集群id: '%s'", ip, node.Status.State, node.Status.Message, chc.cid)
				chc.sendWarnMsg(esg)
			} else if clusterStatusInfo[status] == 200 {
				if chc.cache[ip] == ip {
					delete(chc.cache, ip)
					esg := fmt.Sprintf("节点'%s'已恢复, 当前状态: '%v', 集群id: '%s'", ip, node.Status.State, chc.cid)
					chc.sendWarnMsg(esg)
				}
			}
			chc.updateServerStatus(ip, clusterStatusInfo[role], clusterStatusInfo[status], id)
		}
	}

	primaryIP, err := chc.getPrimaryManager()
	if err != nil {
		log.Printf("Failed to get primary manager: %v\n", err)
		ddwarning.SendWarning(fmt.Sprintf("获取主节点ip失败, 失败信息: '%v', 集群id: '%s'", err, chc.cid))
		return
	}

	// 检查集群是否可用
	if !foundLeader {
		log.Printf("❌ No healthy manager found! Cluster %s may be unavailable.\n", chc.cid)
		ddwarning.SendWarning(fmt.Sprintf("集群异常, 异常节点id: '%s'", chc.cid))
		if err := chc.updatePrimaryManager(primaryManagerIP, 100); err != nil {
			log.Printf("❌ an error occurred while updating the manager node status, cluster [%s], errMsg: %s\n", chc.cid, err.Error())
			ddwarning.SendWarning(fmt.Sprintf("更新集群节点状态失败0, 失败信息: '%v',所在节点id: '%s'", err, chc.cid))
			return
		}
		return
	}

	// 检测leader是否更新
	var isLeaderChange bool
	if primaryIP != primaryManagerIP {
		isLeaderChange = true
	}

	if isLeaderChange {
		log.Printf("✅ Swarm elected new Leader: %s. Updating database...\n", primaryManagerIP)
		if err := chc.updatePrimaryManager(primaryManagerIP, 200); err != nil {
			log.Printf("❌ an error occurred while updating the manager node status, cluster [%s], errMsg: %s\n", chc.cid, err.Error())
			ddwarning.SendWarning(fmt.Sprintf("更新集群节点状态失败1, 失败信息: '%v',集群id: '%s'", err, chc.cid))
			return
		}
	}

	status, err := chc.getPrimaryManagerStatus()
	if err != nil {
		log.Printf("❌ an error occurred while get the manager node status, cluster [%s], errMsg: %s\n", chc.cid, err.Error())
		ddwarning.SendWarning(fmt.Sprintf("更新集群节点状态失败2, 失败信息: '%v',集群id: '%s'", err, chc.cid))
		return
	}

	if status == 300 {
		if err := chc.updatePrimaryManager(primaryManagerIP, 200); err != nil {
			log.Printf("❌ an error occurred while updating the manager node status, cluster [%s], errMsg: %s\n", chc.cid, err.Error())
			ddwarning.SendWarning(fmt.Sprintf("更新集群节点状态失败3, 失败信息: '%v',集群id: '%s'", err, chc.cid))
			return
		}
	}

	log.Printf("cluster %s health ok\n", chc.cid)
}

func (chc *ClusterHealthChecker) NodeRemove(nodeID string) error {
	if err := chc.cli.NodeRemove(chc.ctx, nodeID, types.NodeRemoveOptions{Force: true}); err != nil {
		return err
	}

	return nil
}

func (chc *ClusterHealthChecker) sendWarnMsg(esg string) {
	ddwarning.SendWarning(esg)
	log.Println(esg)
}

func Check(currentServerIp string) {
	cli, err := initCli()
	if err != nil {
		log.Printf("fail to init docker cli, errMsg: %s\n", err.Error())
		return
	}

	defer cli.Close()

	db, err := initDb()
	if err != nil {
		log.Printf("fail to init db, errMsg: %s\n", err.Error())
		return
	}

	defer db.Close()

	c := ClusterHealthChecker{
		ctx:   context.Background(),
		db:    db,
		cli:   cli,
		cache: make(map[string]string),
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	log.Println("start health check")
	for {
		select {
		case <-ticker.C:
			if !c.checkClusterExists(currentServerIp) {
				continue
			}

			cid, err := c.getClusterCid(currentServerIp)
			if err != nil {
				log.Printf("fail to get cluster ID, manager ip currentServerIp, errMsg: %s\n", err.Error())
				return
			}
			if cid != "" {
				c.checkClusterHealth(currentServerIp)
			}
		}
	}
}

func initDb() (*sql.DB, error) {
	db, err := sql.Open("mysql", config.MyConAddr)
	if err != nil {
		return db, err
	}

	return db, nil
}

func initCli() (*client.Client, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return cli, err
	}

	return cli, nil
}

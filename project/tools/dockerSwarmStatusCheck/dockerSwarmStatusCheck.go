package dockerSwarmStatusCheck

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/Lxb921006/Gin-bms/project/config"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
	_ "github.com/go-sql-driver/mysql"
	"log"
	"os/exec"
	"strings"
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
	db  *sql.DB
	cli *client.Client
	cid string
	ctx context.Context // cluster_id
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

func (chc *ClusterHealthChecker) getClusterId(managerIp string) (string, error) {
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

func (chc *ClusterHealthChecker) getCurrentServerIP() (string, error) {
	cmd := exec.Command("dig", "+short", "myip.opendns.com", "@resolver1.opendns.com")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// 获取 Swarm 节点信息（Manager 和 Worker）
func (chc *ClusterHealthChecker) getSwarmNodes() ([]swarm.Node, error) {
	ctx := context.Background()
	nodes, err := chc.cli.NodeList(ctx, types.NodeListOptions{})
	if err != nil {
		return nil, err
	}
	return nodes, nil
}

// 更新 `servers` 表中的节点状态
func (chc *ClusterHealthChecker) updateServerStatus(ip string, role, status uint) {
	query := "UPDATE assets_models SET node_status = ?, node_type = ?, start = NOW() WHERE ip = ?"
	_, err := chc.db.Exec(query, status, role, ip)
	if err != nil {
		log.Printf("❌ Failed to update server status for %s: %v\n", ip, err)
	}
	log.Printf("✅ Updated status for server %s (%v): %v\n", ip, role, status)
}

// 更新 `clusters` 表中的 Primary Manager
func (chc *ClusterHealthChecker) updatePrimaryManager(newPrimaryIP string, status uint) {
	query := "UPDATE assets_models SET master_ip = ?, date = NOW(), status = ? WHERE cluster_cid = ?"
	_, err := chc.db.Exec(query, newPrimaryIP, status, chc.cid)
	if err != nil {
		log.Printf("❌ Failed to update primary manager: %v\n", err)
	}
	log.Printf("✅ Updated primary manager to: %s\n", newPrimaryIP)
}

// **检测所有 Swarm 节点的健康状态**
func (chc *ClusterHealthChecker) checkClusterHealth() {
	log.Println("start health check")
	nodes, err := chc.getSwarmNodes()
	if err != nil {
		log.Fatalf("❌ Failed to get swarm nodes: %v", err)
	}

	var primaryManagerIP string
	var foundLeader bool

	// 遍历所有 Swarm 节点
	for _, node := range nodes {
		ip := node.Status.Addr
		status := string(node.Status.State)
		role := string(node.Spec.Role)

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

		// **更新 servers 表（Worker 和 Manager 状态）**
		chc.updateServerStatus(ip, clusterStatusInfo[role], clusterStatusInfo[status])
	}

	primaryIP, err := chc.getPrimaryManager()
	if err != nil {
		log.Fatalf("Failed to get primary manager: %v", err)
	}

	fmt.Println("AADDD >>> ", primaryIP, primaryManagerIP)

	// 检查集群是否可用
	if !foundLeader {
		log.Println("❌ No healthy manager found! Cluster may be unavailable.")
		chc.updatePrimaryManager(primaryManagerIP, 100)
		return
	}

	// 检测leader是否更新
	var isLeaderChange bool
	if primaryIP != primaryManagerIP {
		isLeaderChange = true
	}

	if isLeaderChange {
		log.Printf("✅ Swarm elected new Leader: %s. Updating database...\n", primaryManagerIP)
		chc.updatePrimaryManager(primaryManagerIP, 200)
	}
}

func Check(currentServerIp string) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		msg := fmt.Sprintf("集群健康检测失败, failed to initialize Docker client, errMsg: %v\n", err)
		//ddwarning.SendWarning(msg)
		log.Fatalln(msg)
	}
	defer cli.Close()

	db, err := sql.Open("mysql", config.MyConAddre)
	if err != nil {
		msg := fmt.Sprintf("集群健康检测失败, failed to connect to database, errMsg: %v\n", err)
		//ddwarning.SendWarning(msg)
		log.Fatalln(msg)
	}
	defer db.Close()

	c := ClusterHealthChecker{
		ctx: context.Background(),
		db:  db,
		cli: cli,
	}

	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			//currentServerIp, err := c.getCurrentServerIP()
			//if err != nil {
			//	log.Printf("fail to get current server ip, errMsg: %s\n", err.Error())
			//	return
			//}

			if !c.checkClusterExists(currentServerIp) {
				continue
			}

			cid, err := c.getClusterId(currentServerIp)
			if err != nil {
				log.Printf("fail to get cluster ID, manager ip currentServerIp, errMsg: %s\n", err.Error())
				return
			}
			if cid != "" {
				c.checkClusterHealth()
			}
		}
	}
}

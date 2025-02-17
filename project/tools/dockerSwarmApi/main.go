package dockerSwarmApi

import (
	"context"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
)

type DockerSwarmOp struct {
	masterIp string
	workIp   string
	token    string
	cli      *client.Client
	ctx      context.Context
}

func NewDockerSwarmOp(masterIp, workIp, token string, ctx context.Context) *DockerSwarmOp {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil
	}

	return &DockerSwarmOp{
		masterIp: masterIp,
		workIp:   workIp,
		token:    token,
		cli:      cli,
		ctx:      ctx,
	}
}

func (d *DockerSwarmOp) CreateSwarm() (string, string, error) {
	// Swarm 初始化请求
	req := swarm.InitRequest{
		ListenAddr:    "0.0.0.0:2377",
		AdvertiseAddr: d.masterIp,
		Spec: swarm.Spec{
			Annotations: swarm.Annotations{Name: "default"},
		},
	}

	_, err := d.cli.SwarmInit(d.ctx, req)
	if err != nil {
		return "", "", err
	}

	info, err := d.cli.Info(d.ctx)
	if err != nil {
		return "", "", err
	}

	token, err := d.getJoinTokens()
	if err != nil {
		return "", "", err
	}

	return info.Swarm.Cluster.ID, token, nil
}

func (d *DockerSwarmOp) getJoinTokens() (string, error) {
	info, err := d.cli.SwarmInspect(d.ctx)
	if err != nil {
		return "", err
	}
	return info.JoinTokens.Worker, nil
}

func (d *DockerSwarmOp) LeaveSwarm() error {
	// 执行节点退出 Swarm
	err := d.cli.SwarmLeave(d.ctx, true)
	if err != nil {
		return err
	}
	return nil
}

func (d *DockerSwarmOp) JoinSwarm() error {
	req := swarm.JoinRequest{
		ListenAddr:    "0.0.0.0:2377",
		AdvertiseAddr: d.workIp,
		RemoteAddrs:   []string{d.masterIp + ":2377"},
		JoinToken:     d.token,
	}

	if err := d.cli.SwarmJoin(d.ctx, req); err != nil {
		return err
	}

	return nil
}

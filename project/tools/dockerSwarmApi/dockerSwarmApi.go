package dockerSwarmApi

import (
	"context"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/client"
)

type DockerSwarmOp struct {
	masterIp    string
	workIp      string
	workToken   string
	masterToken string
	cli         *client.Client
	ctx         context.Context
}

func NewDockerSwarmOp(masterIp, workIp, workToken, masterToken string, ctx context.Context) *DockerSwarmOp {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil
	}

	return &DockerSwarmOp{
		masterIp:    masterIp,
		workIp:      workIp,
		workToken:   workToken,
		masterToken: masterToken,
		cli:         cli,
		ctx:         ctx,
	}
}

func (d *DockerSwarmOp) CreateSwarm() (string, string, string, error) {
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
		return "", "", "", err
	}

	info, err := d.cli.Info(d.ctx)
	if err != nil {
		return "", "", "", err
	}

	wToken, mToken, err := d.getJoinTokens()
	if err != nil {
		return "", "", "", err
	}

	return info.Swarm.Cluster.ID, wToken, mToken, nil
}

func (d *DockerSwarmOp) getJoinTokens() (string, string, error) {
	info, err := d.cli.SwarmInspect(d.ctx)
	if err != nil {
		return "", "", err
	}
	return info.JoinTokens.Worker, info.JoinTokens.Manager, nil
}

func (d *DockerSwarmOp) LeaveSwarm() error {
	// 执行节点退出 Swarm
	err := d.cli.SwarmLeave(d.ctx, true)
	if err != nil {
		return err
	}
	return nil
}

func (d *DockerSwarmOp) JoinWorkSwarm() error {
	req := swarm.JoinRequest{
		ListenAddr:    "0.0.0.0:2377",
		AdvertiseAddr: d.workIp,
		RemoteAddrs:   []string{d.masterIp + ":2377"},
		JoinToken:     d.workToken,
	}

	if err := d.cli.SwarmJoin(d.ctx, req); err != nil {
		return err
	}

	return nil
}

func (d *DockerSwarmOp) JoinMasterSwarm() error {
	req := swarm.JoinRequest{
		ListenAddr:    "0.0.0.0:2377",
		AdvertiseAddr: d.workIp,
		RemoteAddrs:   []string{d.masterIp + ":2377"},
		JoinToken:     d.masterToken,
	}

	if err := d.cli.SwarmJoin(d.ctx, req); err != nil {
		return err
	}

	return nil
}

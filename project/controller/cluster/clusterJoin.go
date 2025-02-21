package cluster

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/mitchellh/mapstructure"
)

func (jj *JoinJson) JoinMaster(ctx *gin.Context) error {
	if err := ctx.ShouldBindJSON(&jj); err != nil {
		return err
	}

	cluster, err := cm.GetCluster(jj.ID)
	if err != nil {
		return err
	}

	if err := mapstructure.Decode(jj, &jj.sw); err != nil {
		return err
	}

	if err := mapstructure.Decode(cluster, &jj.sw); err != nil {
		return err
	}

	if err := jj.sw.JoinMaster(); err != nil {
		return fmt.Errorf("%v", err)
	}

	if err := jj.sw.UpdateServers(); err != nil {
		return err
	}

	return nil
}

func (jj *JoinJson) JoinWork(ctx *gin.Context) error {
	if err := ctx.ShouldBindJSON(jj); err != nil {
		return err
	}

	cluster, err := cm.GetCluster(jj.ID)
	if err != nil {
		return err
	}

	if err := mapstructure.Decode(jj, &jj.sw); err != nil {
		return err
	}

	if err := mapstructure.Decode(cluster, &jj.sw); err != nil {
		return err
	}

	if err := jj.sw.JoinWork(); err != nil {
		return fmt.Errorf("%v", err)
	}

	if err := jj.sw.UpdateServers(); err != nil {
		return err
	}

	return nil
}

func (jj *JoinJson) LeaveSwarm(ctx *gin.Context) error {
	if err := ctx.ShouldBindJSON(jj); err != nil {
		return err
	}

	cluster, err := cm.GetCluster(jj.ID)
	if err != nil {
		return err
	}

	if err := mapstructure.Decode(jj, &jj.sw); err != nil {
		return err
	}

	if err := mapstructure.Decode(cluster, &jj.sw); err != nil {
		return err
	}

	if err := jj.sw.LeaveCluster(); err != nil {
		return fmt.Errorf("%v", err)
	}

	if err := jj.sw.UpdateServers(); err != nil {
		return err
	}

	return nil
}

func (jj *JoinJson) DeleteSwarm(ctx *gin.Context) error {
	if err := ctx.ShouldBindJSON(jj); err != nil {
		return err
	}

	cluster, err := cm.GetCluster(jj.ID)
	if err != nil {
		return err
	}

	if err := mapstructure.Decode(jj, &jj.sw); err != nil {
		return err
	}

	if err := mapstructure.Decode(cluster, &jj.sw); err != nil {
		return err
	}

	if err := jj.sw.LeaveCluster(); err != nil {
		return fmt.Errorf("%v", err)
	}

	if err := jj.sw.UpdateServers(); err != nil {
		return err
	}

	return nil
}

package main

import (
	"fmt"
	"github.com/ingoxx/go-gin/project/tools/ddwarning"
)

func main() {
	ddwarning.SendWarning(fmt.Sprintf("获取节点状态失败, 失败信息: '%v', 集群id: '%s'", "err", "11"))
}

package infrastructure

import (

	"context"
	"transform/pkg/executor/containerd"
	"transform/pkg/executor/docker"
	"transform/pkg/global"
	"transform/utils"
	"transform/utils/log"
)

func IsDocker() bool {
	if global.Docker == nil {
		global.Docker, _ = docker.NewDockerClient()

	}
	if global.Docker != nil {
		_, err := global.Docker.GetClient().Ping(context.Background())
		if err == nil {
			return true
		}
	}
	return false
}

func IsContainerd() bool {
	if global.Containerd == nil {
		global.Containerd, _ = containerd.NewContainedClient()
	}
	if global.Containerd != nil {
		flag, err := global.Containerd.GetClient().IsServing(context.Background())
		if flag && err == nil {
			if !utils.Exists(utils.NerdCtl) {
				log.BKEFormat(log.WARN, "The /usr/bin/nerdctl tool was not found.")
				return false
			}
			return true
		}
	}
	return false
}

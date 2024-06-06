package global

import (
	"os"
	"transform/pkg/executor/containerd"
	"transform/pkg/executor/docker"
	"transform/pkg/executor/exec"
	"transform/utils"
)

var (
	Docker      docker.DockerClient
	Containerd  containerd.ContainerdClient
	Command     exec.Executor
	Workspace   string
	CustomExtra map[string]string
)

func init() {
	Command = &exec.CommandExecutor{}
	Workspace = os.Getenv("BKE_WORKSPACE")
	if Workspace == "" {
		Workspace = "/bke"
		//c, _ := os.Getwd()
		//workspace = filepath.Dir(c)
	}
	if !utils.Exists(Workspace + "/tmpl") {
		_ = os.MkdirAll(Workspace+"/tmpl", 0644)
	}
	if !utils.Exists(Workspace + "/volumes") {
		_ = os.MkdirAll(Workspace+"/volumes", 0644)
	}
	if !utils.Exists(Workspace + "/mount") {
		_ = os.MkdirAll(Workspace+"/mount", 0644)
	}
	CustomExtra = make(map[string]string)
}



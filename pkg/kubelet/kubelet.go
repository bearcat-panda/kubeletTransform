package kubelet

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"time"
	"transform/pkg/executor/containerd"
	"transform/pkg/global"
	"transform/pkg/infrastructure"
	"transform/pkg/root"
	"transform/utils"
	"transform/utils/log"
)

type Options struct {
	root.Options
	Args []string `json:"args"`
	
	HttpRepo string `json:"httpRepo"`
	KubeVersion string `json:"kubeVersion"`
	Runtime string `json:"runtime"`
	Timeout int64 `json:"timeout"`
}

var kubeletService = `
[Unit]
Description=kubelet: The Kubernetes Node Agent
Documentation=https://kubernetes.io/docs/
After=containerd.service network.target local-fs.target


[Service]
ExecStart=/usr/bin/kubelet %s

Restart=always
StartLimitInterval=0
RestartSec=10

# Having non-zero Limit*s causes performance problems due to accounting overhead
# in the kernel. We recommend using cgroups to do container-local accounting.
LimitNPROC=infinity
LimitCORE=infinity
LimitNOFILE=infinity
# Comment TasksMax if your systemd version does not supports it.
# Only systemd 226 and above support this version.
#TasksMax=infinity
OOMScoreAdjust=-999

[Install]
WantedBy=multi-user.target
`
var fileName = "/etc/systemd/system/kubelet.service"
var kubeletName  = "kubelet-%s-%s"

func (op *Options) Resetbak() {
	if infrastructure.IsDocker() {
		log.BKEFormat(log.INFO, "current runtime is docker")
		//获取kubelet运行参数
		kubeletInfo, _ := global.Docker.ContainerExists(utils.KUBELET_NAME)

		//创建kubelet.service文件
		// 创建文件
		file, err := os.Create(fileName)
		defer file.Close()
		if err != nil {
			log.BKEFormat(log.ERROR, err.Error())
			return
		}

		args := strings.Join(kubeletInfo.Args, " ")
		content := fmt.Sprintf(fileName, args)
		content = strings.ReplaceAll(content, `"`, "")

		_, err = file.WriteString(content)
		if err != nil {
			log.BKEFormat(log.ERROR, err.Error())
			return
		}

		//停止容器kubelet
		err = global.Docker.ContainerStop(utils.KUBELET_NAME)

		//运行kubelet.service
		name := fmt.Sprintf(kubeletName, op.KubeVersion, runtime.GOARCH)
		utils.DownloadFile(op.HttpRepo+name, "/usr/bin/kubelet")
		err = global.Command.ExecuteCommand("chmod", "+x", "/usr/bin/kubelet")
		err = global.Command.ExecuteCommand("systemctl ", "enable", "kubelet", "--now")
		result, err := global.Command.ExecuteCommandWithCombinedOutput("systemctl", "restart", "kubelet")
		if err != nil {
			log.BKEFormat(log.ERROR, result)
			return
		}
		log.BKEFormat(log.INFO, result)


		log.BKEFormat(log.INFO, "completed")
		return
	}
	if infrastructure.IsContainerd() {
		log.BKEFormat(log.INFO, "current runtime is containerd")
		//获取kubelet运行参数
		kubeletInfo, ok := containerd.ContainerExists(utils.KUBELET_NAME)
		if !ok {
			log.Error("no found kubelet container")
			return
		}

		//创建kubelet.service文件
		// 创建文件
		file, err := os.Create(fileName)
		defer file.Close()
		if err != nil {
			log.BKEFormat(log.ERROR, err.Error())
			return
		}

		args := strings.Join(kubeletInfo.Args, " ")
		content := fmt.Sprintf(fileName, args)
		content = strings.ReplaceAll(content, `"`, "")

		_, err = file.WriteString(content)
		if err != nil {
			log.BKEFormat(log.ERROR, err.Error())
			return
		}

		//停止容器kubelet
		err = global.Docker.ContainerStop(utils.KUBELET_NAME)

		//运行kubelet.service
		name := fmt.Sprintf(kubeletName, op.KubeVersion, runtime.GOARCH)
		utils.DownloadFile(op.HttpRepo+name, "/usr/bin/kubelet")
		err = global.Command.ExecuteCommand("chmod", "+x", "/usr/bin/kubelet")
		err = global.Command.ExecuteCommand("systemctl ", "enable", "kubelet", "--now")
		result, err := global.Command.ExecuteCommandWithCombinedOutput("systemctl", "restart", "kubelet")
		if err != nil {
			log.BKEFormat(log.ERROR, result)
			return
		}
		log.BKEFormat(log.INFO, result)


		log.BKEFormat(log.INFO, "completed")
		return
	}
}
func (op *Options) Reset() {
	if op.Runtime == "docker" && infrastructure.IsDocker(){
		log.BKEFormat(log.INFO, "current runtime is docker")
		//获取kubelet运行参数
		kubeletInfo, ok := global.Docker.ContainerExists(utils.KUBELET_NAME)
		if !ok {
			//重新启动kubelet
			err := global.Command.ExecuteCommand("bash", "/etc/kubernetes/kubelet.sh", "-a", "start", "-r", op.Runtime)
			if err != nil {
				log.Error(err)
			}
			log.Info("restart kubelet")
			time.Sleep(5*time.Second)
			kubeletInfo, _ = global.Docker.ContainerExists(utils.KUBELET_NAME)
		}
		log.Info(kubeletInfo.Args)

		//创建kubelet.service文件
		// 创建文件
		// 打开文件，如果文件存在则清空它的内容
		file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
		if err != nil {
			log.Fatalf("Failed to open file: %s", err)
			return
		}
		defer file.Close()
		log.Infof("create %s success", fileName)


		args := strings.Join(kubeletInfo.Args, " ")
		content := fmt.Sprintf(kubeletService, args)
		content = strings.ReplaceAll(content, `"`, "")

		_, err = file.WriteString(content)
		if err != nil {
			log.Error(err)
			return
		}

		//停止容器kubelet
		err = global.Docker.ContainerRemove(utils.KUBELET_NAME)
		if err != nil {
			log.Error(err)
			//重新启动kubelet
			err = global.Command.ExecuteCommand("bash", "/etc/kubernetes/kubelet.sh", "-a", "start", "-r", op.Runtime)
			if err != nil {
				log.Error(err)
			}
			log.Info("restart kubelet")
		}
		log.Info("remove kubelet success")

		//运行kubelet.service
		name := fmt.Sprintf(kubeletName, op.KubeVersion, runtime.GOARCH)
		err = utils.DownloadFile(op.HttpRepo+name, "/usr/bin/kubelet")
		if err != nil {
			log.Error(err)
		}
		log.Info("wget kubelet success")


		err = global.Command.ExecuteCommand("chmod", "+x", "/usr/bin/kubelet")
		if err != nil {
			log.Error(err)
		}
retry_docker:
		err = global.Command.ExecuteCommand("systemctl", "daemon-reload")
		if err != nil {
			log.Error(err)
		}
		log.Info("systemctl daemon-reload success")

		err = global.Command.ExecuteCommand("systemctl", "enable", "kubelet", "--now")
		if err != nil {
			log.Error(err)
		}
		log.Info("systemctl enable kubelet --now success")

		result, err := global.Command.ExecuteCommandWithCombinedOutput("systemctl", "restart", "kubelet")
		if err != nil {
			log.Error(err)
		}
		log.Info("systemctl restart kubelet", result)

		result, err = global.Command.ExecuteCommandWithCombinedOutput("systemctl", "status", "kubelet")
		if err != nil {
			log.Error(err)
		}
		log.Info("systemctl status kubelet", result)

		ticker := time.NewTicker(time.Second * 60)
		defer ticker.Stop()
		done := make(chan bool)
		go func() {
			time.Sleep(time.Minute * time.Duration(op.Timeout))
			done <- true
		}()

		for {
			select {
			case <-done:
				log.Error("timeout")
				return
			case <-ticker.C:
				state, err := global.Command.ExecuteCommandWithCombinedOutput("systemctl", "status", "kubelet")
				log.Infof(state, err)
				if !strings.Contains(state, "running"){
					log.Info("Retrying docker...")
					goto retry_docker
				}else {
					log.Info("docker is running.")
					return
				}
			}
		}

		log.BKEFormat(log.INFO, "completed")
		return
	}
	if op.Runtime == "containerd" && infrastructure.IsContainerd(){
		log.BKEFormat(log.INFO, "current runtime is containerd")
		//获取kubelet运行参数
		kubeletInfo, ok := containerd.ContainerExists(utils.KUBELET_NAME)
		if !ok {
			//重新启动kubelet
			err := global.Command.ExecuteCommand("bash", "/etc/kubernetes/kubelet.sh", "-a", "start", "-r", op.Runtime)
			if err != nil {
				log.Error(err)
			}
			log.Info("restart kubelet")
			time.Sleep(5*time.Second)
			kubeletInfo, _ = containerd.ContainerExists(utils.KUBELET_NAME)
		}

		log.Info(kubeletInfo.Args)

		//创建kubelet.service文件
		// 创建文件
		// 打开文件，如果文件存在则清空它的内容
		file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0644)
		if err != nil {
			log.Fatalf("Failed to open file: %s", err)
			return
		}
		defer file.Close()
		log.Infof("create %s success", fileName)


		args := strings.Join(kubeletInfo.Args, " ")
		content := fmt.Sprintf(kubeletService, args)
		content = strings.ReplaceAll(content, `"`, "")

		_, err = file.WriteString(content)
		if err != nil {
			log.Error(err)
			return
		}

		//停止容器kubelet
		err = containerd.ContainerRemove(utils.KUBELET_NAME)
		if err != nil {
			log.Error(err)
		}
		log.Info("remove kubelet success")

		//运行kubelet.service
		name := fmt.Sprintf(kubeletName, op.KubeVersion, runtime.GOARCH)
		err = utils.DownloadFile(op.HttpRepo+name, "/usr/bin/kubelet")
		if err != nil {
			log.Error(err)
		}
		log.Info("wget kubelet success")


		err = global.Command.ExecuteCommand("chmod", "+x", "/usr/bin/kubelet")
		if err != nil {
			log.Error(err)
		}
retry_containerd:
		err = global.Command.ExecuteCommand("systemctl", "daemon-reload")
		if err != nil {
			log.Error(err)
		}
		log.Info("systemctl daemon-reload success")

		err = global.Command.ExecuteCommand("systemctl", "enable", "kubelet", "--now")
		if err != nil {
			log.Error(err)
		}
		log.Info("systemctl enable kubelet --now success")

		result, err := global.Command.ExecuteCommandWithCombinedOutput("systemctl", "restart", "kubelet")
		if err != nil {
			log.Error(err)
		}
		log.Info("systemctl restart kubelet", result)

		result, err = global.Command.ExecuteCommandWithCombinedOutput("systemctl", "status", "kubelet")
		if err != nil {
			log.Error(err)
		}
		log.Info("systemctl status kubelet", result)

		ticker := time.NewTicker(time.Second * 60)
		defer ticker.Stop()
		done := make(chan bool)
		go func() {
			time.Sleep(time.Minute * time.Duration(op.Timeout))
			done <- true
		}()

		for {
			select {
			case <-done:
				log.Error("timeout")
				return
			case <-ticker.C:
				state, err := global.Command.ExecuteCommandWithCombinedOutput("systemctl", "status", "kubelet")
				log.Infof(state, err)
				if !strings.Contains(state, "running"){
					log.Info("Retrying containerd...")
					goto retry_containerd
				}else {
					log.Info("Container is running.")
					return
				}
			}
		}


		log.BKEFormat(log.INFO, "completed")
		return
	}
}
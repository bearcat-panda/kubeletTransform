package containerd

import (
	"encoding/json"
	"errors"
	"transform/pkg/executor/exec"
	"transform/utils"
	"transform/utils/log"
)

type NerdContainerInfo struct {
	Id    string `json:"Id"`
	State struct {
		Status     string `json:"Status"`
		Running    bool   `json:"Running"`
		Paused     bool   `json:"Paused"`
		Restarting bool   `json:"Restarting"`
		Pid        uint   `json:"Pid"`
		ExitCode   uint   `json:"ExitCode"`
		FinishedAt string `json:"FinishedAt"`
	} `json:"State"`
	Image           string `json:"Image"`
	Args            []string `json:"Args"`
	Name            string `json:"Name"`
	RestartCount    uint   `json:"RestartCount"`
	Platform        string `json:"Platform"`
	NetworkSettings struct {
		IPAddress  string `json:"IPAddress"`
		MacAddress string `json:"MacAddress"`
	} `json:"NetworkSettings"`
}



var cmd = exec.CommandExecutor{}



func ContainerExists(containerId string) (NerdContainerInfo, bool) {
	info := []NerdContainerInfo{}
	result, err := cmd.ExecuteCommandWithOutput(utils.NerdCtl, "-n", "k8s.io", "inspect", containerId)
	if err != nil {
		log.Error(err)
		return NerdContainerInfo{}, false
	}
	log.Debug(result)
	err = json.Unmarshal([]byte(result), &info)
	if err != nil {
		log.BKEFormat(log.ERROR, err.Error())
		return NerdContainerInfo{}, false
	}
	if len(info) == 1 {
		return info[0], true
	}
	return NerdContainerInfo{}, false
}

func ContainerInspect(containerId string) (NerdContainerInfo, error) {
	info := []NerdContainerInfo{}
	result, err := cmd.ExecuteCommandWithOutput(utils.NerdCtl, "-n", "k8s.io", "inspect", containerId)
	if err != nil {
		return NerdContainerInfo{}, err
	}
	log.Debug(result)
	err = json.Unmarshal([]byte(result), &info)
	if err != nil {
		log.BKEFormat(log.ERROR, err.Error())
		return NerdContainerInfo{}, err
	}
	if len(info) == 1 {
		return info[0], nil
	}
	return NerdContainerInfo{}, errors.New("not found")
}

func ContainerRemove(containerId string) error {
	err := cmd.ExecuteCommand(utils.NerdCtl, "-n", "k8s.io", "rm", "-f", containerId)
	if err != nil {
		return err
	}
	return nil
}


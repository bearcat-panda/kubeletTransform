package configuration

import (
	"fmt"
	"net"
)



// HostConfig 初始配置
type HostConfig struct {
	// 主机列表
	Hosts      []Host `json:"hosts" yaml:"hosts"`
}

type Host struct {
	IP       string   `json:"ip" yaml:"ip"`
	UserName string   `json:"username" yaml:"username"`
	Password string   `json:"password" yaml:"password"`
	Port     string   `json:"port" yaml:"port"`
}

var Instance HostConfig

func (h *Host) Validate() (*Host, error) {
	if h.UserName == "" {
		return nil, fmt.Errorf("Host's user field is required ")
	}
	if h.Password == "" {
		return nil, fmt.Errorf("At least one of the host's password and ssh key is provided ")
	}
	if h.IP == "" {
		return nil, fmt.Errorf("Host address is required ")
	}
	if a := net.ParseIP(h.IP); a == nil {
		return nil, fmt.Errorf("Host's address not a valid IP address ")
	}
	if h.Port == "" {
		return nil, fmt.Errorf("Host's port must be greater than zero ")
	}
	return h, nil
}

func (h Host) Fields() (string, string, string, string) {
	return h.UserName, h.Password, h.IP, h.Port
}




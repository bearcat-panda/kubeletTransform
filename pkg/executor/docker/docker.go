package docker

import (
	"context"
	"errors"
	"transform/utils"
	"transform/utils/log"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	dockerapi "github.com/docker/docker/client"
)

type DockerClient interface {
	GetClient() *dockerapi.Client
	ContainerStop(containerId string) error
	ContainerRemove(containerId string) error
	ContainerExists(containerName string) (types.ContainerJSON, bool)
	Exec(containerID string, command []string) (ExecResult, error)
}

type Client struct {
	Client *dockerapi.Client
	ctx    context.Context
}


type ContainerRef struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

const dockerSock = "/var/run/docker.sock"

func NewDockerClient() (DockerClient, error) {
	if !utils.Exists(dockerSock) {
		log.BKEFormat(log.ERROR, "docker service does not exist. ")
		return nil, errors.New("docker service does not exist. ")
	}

	ctx := context.Background()
	cli, err := dockerapi.NewClientWithOpts(dockerapi.FromEnv, dockerapi.WithAPIVersionNegotiation())
	if err != nil {
		log.Debugf("get container runtime client err:", err)
		return nil, err
	}
	return &Client{
		Client: cli,
		ctx:    ctx,
	}, nil
}

func (c *Client) Close() {
	_ = c.Client.Close()
}

func (c *Client) GetClient() *dockerapi.Client {
	return c.Client
}


func (c *Client) ContainerExists(containerName string) (types.ContainerJSON, bool) {
	containerInfo, _ := c.Client.ContainerInspect(c.ctx, containerName)
	// Check whether the mirror warehouse already exists
	if containerInfo.ContainerJSONBase != nil {
		return containerInfo, true
	}
	return types.ContainerJSON{}, false

}
func (c *Client) ContainerRemove(containerId string) error {
	// docker rm
	containerRemoveOptions := types.ContainerRemoveOptions{Force: true}
	if err := c.Client.ContainerRemove(c.ctx, containerId, containerRemoveOptions); err != nil {
		log.Debugf("remove container %s error: %v", containerId, err)
		return err
	}
	return nil
}

func (c *Client) ContainerStop(containerId string) error {
	// docker stop
	if err := c.Client.ContainerStop(c.ctx, containerId, container.StopOptions{}); err != nil {
		log.Debugf("stop container %s error: %v", containerId, err)
		return err
	}
	return nil
}



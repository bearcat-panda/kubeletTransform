package containerd

import (
	"context"
	"errors"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/containers"
	"github.com/containerd/containerd/namespaces"
	"transform/utils"
	"transform/utils/log"
)

type ContainerdClient interface {
	GetClient() *containerd.Client
	ContainerExists(containerName string) (containers.Container, bool)

}


type Client struct {
	condClient  *containerd.Client
	ctx         context.Context
	cancel      context.CancelFunc
}

var (
	containerdSock      = "unix:///var/run/containerd/containerd.sock"
	containerdNamespace = "k8s.io"
	containerdSockLinux = "/var/run/containerd/containerd.sock"
)

func NewContainedClient() (ContainerdClient, error) {
	if !utils.Exists(containerdSockLinux) {
		log.Error("docker service does not exist. ")
		return nil, errors.New("containerd service does not exist. ")
	}

	ctx := context.Background()
	ctx = namespaces.WithNamespace(ctx, containerdNamespace)
	condClient, err := containerd.New(containerdSockLinux)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	var cancel context.CancelFunc
	ctx, cancel = context.WithCancel(ctx)
	return &Client{
		condClient:  condClient,
		ctx:         ctx,
		cancel:      cancel,
	}, nil
}

func (c *Client) Close() {
	if c.condClient != nil {
		_ = c.condClient.Close()
	}
}

func (c *Client) GetClient() *containerd.Client {
	return c.condClient
}


func (c *Client) ContainerExists(containerName string) (containers.Container, bool) {
	container, err := c.condClient.ContainerService().Get(c.ctx, containerName)
	if err != nil {
		log.Error(err)
		return container, false
	}
	// Check whether the mirror warehouse already exists
	if len(container.Image) > 0 {
		return container, true
	}
	return container, false

}


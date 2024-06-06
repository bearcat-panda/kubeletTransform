package docker

import (
	"bytes"
	"context"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/stdcopy"
	"io"
)

type ExecResult struct {
	StdOut   string
	StdErr   string
	ExitCode int
}

func (c *Client) Exec(containerID string, command []string) (ExecResult, error) {

	var execResult ExecResult
	ctx, cancel := context.WithCancel(c.ctx)
	defer cancel()
	config := types.ExecConfig{
		AttachStderr: true,
		AttachStdout: true,
		Cmd:          command,
	}
	idResponse, err := c.Client.ContainerExecCreate(ctx, containerID, config)
	if err != nil {
		return execResult, err
	}

	resp, err := c.Client.ContainerExecAttach(ctx, idResponse.ID, types.ExecStartCheck{})
	if err != nil {
		return execResult, err
	}
	defer resp.Close()

	// read the output
	var outBuf, errBuf bytes.Buffer
	outputDone := make(chan error)

	go func() {
		// StdCopy demultiplexes the stream into two buffers
		_, err = stdcopy.StdCopy(&outBuf, &errBuf, resp.Reader)
		outputDone <- err
	}()

	select {
	case err = <-outputDone:
		if err != nil {
			return execResult, err
		}
		break

	case <-ctx.Done():
		return execResult, ctx.Err()
	}

	stdout, err := io.ReadAll(&outBuf)
	if err != nil {
		return execResult, err
	}
	stderr, err := io.ReadAll(&errBuf)
	if err != nil {
		return execResult, err
	}

	res, err := c.Client.ContainerExecInspect(ctx, idResponse.ID)
	if err != nil {
		return execResult, err
	}

	execResult.ExitCode = res.ExitCode
	execResult.StdOut = string(stdout)
	execResult.StdErr = string(stderr)
	return execResult, nil
}

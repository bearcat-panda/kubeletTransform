package remote

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	gossh "golang.org/x/crypto/ssh"
)

type ssh struct {
	sshClient *gossh.Client
}

// NewSSHClient new ssh client
func NewSSHClient(user string, password string, host string, port string, sshKey interface{}) (*ssh, error) {
	var (
		sshClient *gossh.Client
		err       error
	)

	switch {
	case user != "" && password != "" && host != "":
		if sshClient, err = NewNormalSSHClient(user, password, host, port); err != nil {
			return nil, err
		}
		break
	case sshKey != "":
		if sshClient, err = NewWithOutPassSSHClient(sshKey); err != nil {
			return nil, err
		}
	default:
		return nil, fmt.Errorf("some fields are blank")
	}

	return &ssh{
		sshClient: sshClient,
	}, nil
}

// NewNormalSSHClient new ssh client with username and password
func NewNormalSSHClient(user string, password string, host string, port string) (*gossh.Client, error) {
	config := &gossh.ClientConfig{
		User:            user,
		Auth:            []gossh.AuthMethod{gossh.Password(password)},
		Timeout:         30 * time.Second,
		HostKeyCallback: func(hostname string, remote net.Addr, key gossh.PublicKey) error { return nil },
	}
	config.SetDefaults()

	address := fmt.Sprintf("%s:%s", host, port)

	client, err := gossh.Dial("tcp", address, config)
	if err != nil {
		return nil, err
	}

	return client, nil
}

// NewWithOutPassSSHClient new ssh client with ssh key
func NewWithOutPassSSHClient(sshKey interface{}) (*gossh.Client, error) {
	return nil, nil
}

// Exec	command on remote host
// just return stderr and error
func (s *ssh) Exec(cmd string) ([]string, []string, error) {
	var stderrs []string
	var result []string

	if s.sshClient == nil {
		return result, stderrs, errors.New("before run, have to new a ssh client")
	}

	// 不执行命令直接返回
	if cmd == "" {
		return result, stderrs, nil
	}

	session, err := s.sshClient.NewSession()
	if err != nil {
		return result, stderrs, errors.New(fmt.Sprintf(err.Error(), "Create session failed"))
	}

	defer session.Close()

	r, _ := session.StdoutPipe()
	e, _ := session.StderrPipe()

	go func() {
		if err := session.Run(cmd); err != nil {
			return
		}
	}()

	outReader := bufio.NewReader(r)
	errReader := bufio.NewReader(e)

	for {
		if res, err := s.readPipe(outReader); err != nil {
			if err == io.EOF {
				return result, stderrs, nil
			}
			return result, stderrs, err
		} else {
			result = append(result, res)
		}
		stderr, err := s.readPipe(errReader)
		if stderr != "" {
			stderrs = append(stderrs, stderr)
		}
		if err != nil {
			if err == io.EOF {
				continue
			}
			return result, stderrs, err
		}
	}
}

// readPipe read pipe buffer
func (s *ssh) readPipe(reader *bufio.Reader) (string, error) {
	line, _, err := reader.ReadLine()
	if err == io.EOF {
		return "", err
	}
	if err != nil && err != io.EOF {
		return "", errors.New(fmt.Sprintf(err.Error(), "Read pipe buffer failed"))
	}
	return string(line), nil
}

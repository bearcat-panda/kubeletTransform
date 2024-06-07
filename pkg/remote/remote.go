package remote

import (
	"errors"
	"fmt"
	"io"
	"log"
	"transform/pkg/configuration"
)

type Cli struct {
	User     string
	Password string
	SSHKey   interface{}
	Address  string
	Port     string
	SSH      *ssh
	SFTP     *sftp
}

func (c Cli) Fields() (string, string, string, string, interface{}) {
	return c.User, c.Password, c.Address, c.Port, c.SSHKey
}

// Run supports executing commands and uploading files on the remote hosts
// return a map contains the standard stderr,and the map key is remote host ip
func Run(hosts []configuration.Host, cmd Command) map[string][]string {
	if len(hosts) == 0 {
		return map[string][]string{}
	}
	stopChan := make(chan bool)
	stderrsChan := make(chan map[string][]string)
	stderrs := make(map[string][]string)

	q := 0

	for _, h := range hosts {
		go run(h, cmd, stopChan, stderrsChan)
	}

	for {
		select {
		case _ = <-stopChan:
			if q += 1; q == len(hosts) {
				return stderrs
			}
		case stderrsMap := <-stderrsChan:
			for host := range stderrsMap {
				stderrs[host] = append(stderrs[host], stderrsMap[host]...)
			}
		}
	}
}

// run Command on the remote hosts
func run(h configuration.Host, cmd Command, stopChan chan bool, stderrsChan chan map[string][]string) {
	stderrs := make(map[string][]string)
	RemoteClient, err := NewRemoteClient(&h)

	if err != nil || RemoteClient == nil {
		log.Println(fmt.Sprintf("Failed to create remote ssh client: %v", err))
		stderrs[h.IP] = append(stderrs[h.IP], err.Error())
		stderrsChan <- stderrs
		stopChan <- true
		return
	}

	defer func() {
		_ = RemoteClient.CloseRemoteCli(stopChan)
		if err := recover(); err != nil {
			stderrs[h.IP] = append(stderrs[h.IP], fmt.Sprintf("%v", err))
		}
		stderrsChan <- stderrs
	}()

	for _, file := range cmd.FileUp {
		if err := RemoteClient.SFTP.UploadFile(file.Src, file.Dst); err != nil {
			log.Println(fmt.Sprintf("Failed to upload file %q to %q %q, err %s", file.Src, h.IP, file.Dst, err.Error()))
			for i := 0; i < 3; i++ {
				log.Println(fmt.Sprintf("Wiat 10 seconds and try again %d", i+1))
				err = RemoteClient.SFTP.UploadFile(file.Src, file.Dst)
				if err == nil {
					break
				}
			}
			if err != nil {
				stderrs[h.IP] = append(stderrs[h.IP], err.Error())
				stderrsChan <- stderrs
			}
			continue
		}
		log.Println(fmt.Sprintf("Upload file %q to %q %q, Success", file.Src, h.IP, file.Dst))
	}

	for _, c := range cmd.List() {
		_, stderr, err := RemoteClient.SSH.Exec(c)
		if len(stderr) != 0 {
			log.Println(fmt.Sprintf("Failed to execute command %q on %q, stderr: %s", c, h.IP, stderr))
			stderrs[h.IP] = append(stderrs[h.IP], stderr...)
			stderrsChan <- stderrs
		}
		if err != nil {
			log.Println(fmt.Sprintf("Failed to execute command %q on %q, err: %s", c, h.IP, err.Error()))
			stderrs[h.IP] = append(stderrs[h.IP], err.Error())
			stderrsChan <- stderrs
		}
		if len(stderr) == 0 && err == nil {
			log.Println(fmt.Sprintf("Execute command %q on %q, Success", c, h.IP))
		}
	}
}

// NewRemoteClient returns a new remote client with ssh and sftp client
func NewRemoteClient(h *configuration.Host) (*Cli, error) {
	var err error

	h, err = h.Validate()
	if err != nil {
		return nil, errors.New(fmt.Sprintf(err.Error(), "host validation failed"))
	}

	c := &Cli{
		User:     h.UserName,
		Password: h.Password,
		SSHKey:   "",
		Address:  h.IP,
		Port:     h.Port,
	}

	c.SSH, err = NewSSHClient(c.Fields())
	if err != nil {
		return nil, errors.New(fmt.Sprintf(err.Error(), "Failed to create ssh client"))
	}

	c.SFTP, err = NewSFTPClient(c.SSH.sshClient)
	if err != nil {
		return nil, err
	}

	return c, nil
}

// CloseRemoteCli 关闭远程客户端
func (c *Cli) CloseRemoteCli(stopChan chan bool) error {
	if err := c.SFTP.sftpClient.Close(); err != nil && err != io.EOF {
		return errors.New("failed to close sftp client")
	}
	stopChan <- true
	return nil
}

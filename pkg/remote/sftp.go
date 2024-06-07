package remote

import (
	"io"
	"os"
	"path"

	"github.com/pkg/errors"
	gosftp "github.com/pkg/sftp"
	gossh "golang.org/x/crypto/ssh"
)

type sftp struct {
	sftpClient *gosftp.Client
}

// NewSFTPClient 以ssh客户端为基础建立sftp连接客户端
func NewSFTPClient(sshClient *gossh.Client) (*sftp, error) {
	sftpClient, err := gosftp.NewClient(sshClient)
	if err != nil {
		return nil, errors.Wrap(err, "Failed to new sftp client")
	}
	return &sftp{
		sftpClient: sftpClient,
	}, nil
}

func (s *sftp) UploadFile(localFilePath string, remoteDirPath string) error {
	if s.sftpClient == nil {
		return errors.New("Before run, have to new a sftp client")
	}

	// 打开本地文件
	localFile, err := os.Open(localFilePath)
	defer localFile.Close()
	if err != nil {
		return errors.Wrap(err, "Failed to open local file")
	}

	// 创建远程文件
	remoteFileName := path.Base(localFilePath)
	remoteFilePath := path.Join(remoteDirPath, remoteFileName)

	// 有时因为远程文件已经存在，s.sftpClient.Create会报错，所以默认先删除在重新创建。
	_ = s.sftpClient.Remove(remoteFilePath)

	remoteFile, err := s.sftpClient.Create(remoteFilePath)
	if err != nil {
		return errors.Wrap(err, "Failed to create remote file")
	}
	defer remoteFile.Close()

	// 用io.Copy的方式上传文件，速度更快
	_, err = io.Copy(remoteFile, localFile)
	if err != nil {
		return errors.Wrap(err, "Failed to copy local file to remote file")
	}

	remoteFileStat, err := remoteFile.Stat()
	if err != nil {
		return err
	}
	localFileStat, err := localFile.Stat()
	if err != nil {
		return err
	}

	if remoteFileStat.Size() != localFileStat.Size() {
		if err := s.sftpClient.Remove(path.Join(remoteDirPath, remoteFileName)); err != nil {
			return errors.Wrap(err, "Failed to remove damaged file")
		}
		return errors.New("Failed to upload file, file size not match")
	}
	return nil
}

func (s *sftp) DownloadFile(localFilePath string, remoteDirPath string) error {
	if s.sftpClient == nil {
		return errors.New("Before run, have to new a sftp client")
	}

	// 从SFTP服务器下载文件
	srcFile, err := s.sftpClient.Open(remoteDirPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	_ = os.Remove(localFilePath)
	dstFile, err := os.Create(localFilePath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	_, err = srcFile.WriteTo(dstFile)
	if err != nil {
		return err
	}
	return nil
}

package utils

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
)

func Md5string(file string) (string, error) {
	f, err := os.Open(file) //打开文件
	if nil != err {
		fmt.Println(err)
		return "", err
	}
	defer f.Close()

	md5Handle := md5.New()         //创建 md5 句柄
	_, err = io.Copy(md5Handle, f) //将文件内容拷贝到 md5 句柄中
	if nil != err {
		fmt.Println(err)
		return "", err
	}
	md := md5Handle.Sum(nil)        //计算 MD5 值，返回 []byte
	md5str := fmt.Sprintf("%x", md) //将 []byte 转为 string
	return md5str, nil
}

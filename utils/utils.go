package utils

import (
	"math/rand"
	"net"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return os.IsExist(err)
	}
	return true
}

func ContainsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func ContainsStringPrefix(slice []string, s string) bool {
	for _, item := range slice {
		if strings.HasPrefix(item, s) {
			return true
		}
	}
	return false
}

// IsDir 判断所给路径是否为文件夹
func IsDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

// IsFile 判断所给路径是否为文件
func IsFile(path string) bool {
	return !IsDir(path)
}

func AbsPath(path string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	} else {
		return filepath.Abs(path)
	}
}

func DirectoryIsEmpty(path string) bool {
	dir, _ := os.ReadDir(path)
	if len(dir) == 0 {
		return true
	}
	return false
}

func GetOutBoundIP() (ip string, err error) {
	conn, err := net.Dial("udp", "8.8.8.8:53")
	if err != nil {
		return
	}
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ip = strings.Split(localAddr.String(), ":")[0]
	return
}

func GetIntranetIp() (ip string, err error) {
	addresses, err := net.InterfaceAddrs()
	if err != nil {
		return
	}
	for _, address := range addresses {
		if ipNet, ok := address.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String(), nil
			}
		}
	}
	return
}

func IsNum(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

var letterRunes = []rune("abcdefghijklmnopqrstuvwxyz123456789")

func RandStringRunes(n int) string {
	b := make([]rune, n)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	return string(b)
}

func CopyFile(sourceFile, destinationFile string) error {
	input, err := os.ReadFile(sourceFile)
	if err != nil {
		return err
	}
	err = os.WriteFile(destinationFile, input, 0644)
	if err != nil {
		return err
	}
	return nil
}

func RemoveStringObject(array []string, obj string) []string {
	for i := 0; i < len(array); i++ {
		if array[i] == obj {
			array = append(array[:i], array[i+1:]...)
		}
	}
	return array
}

func IsChanClosed(ch interface{}) bool {
	if reflect.TypeOf(ch).Kind() != reflect.Chan {
		panic("only channels!")
	}
	cptr := *(*uintptr)(unsafe.Pointer(
		unsafe.Pointer(uintptr(unsafe.Pointer(&ch)) + unsafe.Sizeof(uint(0))),
	))

	cptr += unsafe.Sizeof(uint(0)) * 2
	cptr += unsafe.Sizeof(unsafe.Pointer(uintptr(0)))
	cptr += unsafe.Sizeof(uint16(0))
	return *(*uint32)(unsafe.Pointer(cptr)) > 0
}

// ReverseArray reverse array
func ReverseArray(arr []string) []string {
	for i, j := 0, len(arr)-1; i < j; i, j = i+1, j-1 {
		arr[i], arr[j] = arr[j], arr[i]
	}
	return arr
}

package utils

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"
	"strings"
)

func DownloadAllFiles(url, targetDirectory string) error {
	htmlData, err := httpGet(url)
	if err != nil {
		return err
	}
	if len(htmlData) == 0 {
		return errors.New(fmt.Sprintf("url: %s, Failed to get download list", url))
	}
	re := regexp.MustCompile("<a href=\"(.*?)\">(.*?)</a>")
	result := re.FindAllStringSubmatch(htmlData, -1)

	for _, res := range result {
		if len(res) < 2 {
			continue
		}
		if !strings.HasSuffix(res[1], ".rpm") {
			continue
		}
		fmt.Println(res[1])
		err = DownloadFile(url+res[1], path.Join(targetDirectory, res[1]))
		if err != nil {
			return err
		}
	}
	return nil
}


func httpGet(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", errors.New(fmt.Sprintf(" get url %s, status code %d", url, resp.StatusCode))
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func DownloadFile(url, destinationFile string) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return errors.New(fmt.Sprintf("File cannot be found %d", resp.StatusCode))
	}
	reader := bufio.NewReaderSize(resp.Body, 32*1024)
	file, err := os.Create(destinationFile)
	if err != nil {
		return err
	}
	writer := bufio.NewWriter(file)
	_, err = io.Copy(writer, reader)
	if err != nil {
		return err
	}
	return nil
}

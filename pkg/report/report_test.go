package report

import (
	"fmt"
	"gopkg.in/yaml.v3"
	"log"
	"os"
	"strings"
	"testing"
	"time"
)

// TestReport 生成一个测试Report的方法
func TestReport(t *testing.T) {
	data := ReportData{
		StartTime:    "2023-04-01 12:10:10",
		DurationTime: "5600",
		RandomSeed:   "123456789",
		Total:        10,
		Success:      10,
		Failure:      0,
		Warning:      0,
		Result:       "pass",
		Case: []CaseInfo{
			{
				Name:   "case1",
				Status: "Success",
				Detail: "case1",
			},
			{
				Name:   "case2",
				Status: "failure",
				Detail: "case2",
			},
			{
				Name:   "case3",
				Status: "warning",
				Detail: "case3",
			},
		},
		Server: []ServerInfo{
			{OS: "Ubuntu 20.04", CPU: "Intel Core i7", Memory: "16 GB", Network: "eth0", Disk: "512 GB SSD"},
			{OS: "CentOS 7", CPU: "AMD Ryzen 5", Memory: "8 GB", Network: "eth0", Disk: "256 GB SSD"},
		},
	}
	err := Report(data)
	if err != nil {
		panic(err)
	}
}

func TestGenerateReport(t *testing.T) {
	startTime := time.Now()
	// 获取/tmp/report目录下所有文件
	files, err := os.ReadDir("/tmp/report")
	if err != nil {
		t.Fatal(err)
	}
	reportList := []ReportData{}
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if strings.HasSuffix(file.Name(), ".yaml") {
			// 读取文件并序列化
			data, err := os.ReadFile(fmt.Sprintf("/tmp/report/%s", file.Name()))
			if err != nil {
				log.Println(err.Error())
				continue
			}
			var rep ReportData
			err = yaml.Unmarshal(data, &rep)
			if err != nil {
				log.Println(err.Error())
				continue
			}
			reportList = append(reportList, rep)
			continue
		}
	}

	err = GenerateReport(startTime, reportList)
	if err != nil {
		log.Println(err.Error())
		return
	}

}

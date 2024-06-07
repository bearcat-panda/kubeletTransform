package report

import (
	"bytes"
	"crypto/md5"
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"text/template"
	"time"
)

var (
	//go:embed report.tpl
	reportTpl string
)

const (
	Success = "Success"
	Failure = "Failure"
	Warning = "Warning"
	PASS    = "pass"
	NOTPASS = "not pass"
)

// ReportData 结构体
type ReportData struct {
	StartTime    string       `yaml:"startTime" json:"startTime"`
	DurationTime string       `yaml:"durationTime" json:"durationTime"`
	RandomSeed   string       `yaml:"randomSeed" json:"randomSeed"`
	Total        int          `yaml:"total" json:"total"`
	Success      int          `yaml:"success" json:"success"`
	Failure      int          `yaml:"failure" json:"failure"`
	Warning      int          `yaml:"warning" json:"warning"`
	Result       string       `yaml:"result" json:"result"`
	Case         []CaseInfo   `yaml:"case" json:"case"`
	Server       []ServerInfo `yaml:"server" json:"server"`
}

type CaseInfo struct {
	Identify     string `yaml:"identify" json:"identify"`
	IP           string `yaml:"ip" json:"ip"`
	Role         string `yaml:"role" json:"role"`
	Name         string `yaml:"name" json:"name"`
	Status       string `yaml:"status" json:"status"`
	Detail       string `yaml:"detail" json:"detail"`
	DurationTime string `yaml:"durationTime" json:"durationTime"`
}

type ServerInfo struct {
	IP        string   `yaml:"ip" json:"ip"`
	RoleDef   []string `yaml:"roleDef" json:"roleDef"`
	Role      string   `yaml:"role" json:"role"`
	OS        string   `yaml:"os" json:"os"`
	Kernel    string   `yaml:"kernel" json:"kernel"`
	CPU       string   `yaml:"cpu" json:"cpu"`
	Memory    string   `yaml:"memory" json:"memory"`
	Network   string   `yaml:"network" json:"network"`
	Disk      string   `yaml:"disk" json:"disk"`
	IsPhysics int      `yaml:"isPhysics" json:"isPhysics"`
}

// Report is a report generator
func Report(data ReportData) error {
	tmpl0, err := template.New("report").Parse(reportTpl)
	if err != nil {
		return err
	}
	buf0 := new(bytes.Buffer)
	err = tmpl0.Execute(buf0, data)
	if err != nil {
		return err
	}
	_ = os.Remove("report.html")
	err = os.WriteFile("report.html", buf0.Bytes(), 0644)
	if err != nil {
		return err
	}
	return nil
}

// GenerateReport 提供多个ReportData结构体，合并结构体结果，生成报告
func GenerateReport(startTime time.Time, dataList []ReportData) error {
	rd := ReportData{
		StartTime: startTime.Format("2006-01-02 15:04:05"),
		Result:    PASS,
	}
	for _, data := range dataList {
		rd.Total += data.Total
		rd.Success += data.Success
		rd.Failure += data.Failure
		rd.Warning += data.Warning
		rd.Case = append(rd.Case, data.Case...)
		rd.Server = append(rd.Server, data.Server...)
	}


	if rd.Failure > 0 {
		rd.Result = NOTPASS
	}
	factor := fmt.Sprintf("%s%d%d%d%d", rd.StartTime, rd.Total, rd.Success, rd.Failure, rd.Warning)
	// 计算factor的md5值
	rd.RandomSeed = fmt.Sprintf("%x", md5.Sum([]byte(factor)))
	rd.DurationTime = time.Now().Sub(startTime).String()
	err := Report(rd)
	if err != nil {
		return err
	}
	err = outputJsonReport(rd)
	if err != nil {
		return err
	}
	return nil
}

// outputJsonReport 输出json报告
func outputJsonReport(data ReportData) error {
	buf, err := json.Marshal(data)
	if err != nil {
		return err
	}
	err = os.WriteFile("report.json", buf, 0644)
	if err != nil {
		return err
	}
	return nil
}

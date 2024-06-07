package batch

import (
	"errors"
	"fmt"
	"os"
	"path"
	"strings"
	"time"
	"transform/pkg/configuration"
	"transform/pkg/remote"
	"transform/pkg/report"
	"transform/pkg/root"
	"transform/utils"
	"transform/utils/log"

	"gopkg.in/yaml.v3"
)

var transformName = "transfrom"

type Options struct {
	root.Options
	File string `yaml:"file" json:"file"`
	HttpRepo string `json:"httpRepo"`
	KubeVersion string `json:"kubeVersion"`
	Runtime string `json:"runtime"`
}

type nodeTask struct {
	ip   string
	cli  *remote.Cli
}

var (
	configFile    = "nodes.yaml"
	errorFile     = "error.log"
	resultFile    = "result.yaml"
	httppid       = "httppid"
	checkpid      = "checkpid"
	nodeTaskMap   = make(map[string]nodeTask)
	nodeResultMap = make(map[string]uint)
	AMD64Host     = []configuration.Host{}
	ARM64Host     = []configuration.Host{}
)

func (op *Options) Run() {
	startTime := time.Now()
	_ = os.RemoveAll("/tmp/report")
	err := os.Mkdir("/tmp/report", 0644)
	if err != nil {
		log.Info(err.Error())
		return
	}
	// 第一步：初始化配置文件
	log.Info("建立与各个节点的连接...")
	res, errNum, err := op.ConfigValidation()
	if err != nil {
		log.Info(fmt.Sprintf("校验失败: %s,详细信息见report.html", err.Error()))
		generateErrorReport(startTime, res, errNum, nil)
		return
	}

	// 第三步：分发二进制文件到各个节点，并启动
	envInit1 := remote.Command{
		Cmds: []string{"sudo mkdir -p /tmp/precheck",
			"sudo chmod 777 /tmp/precheck",
			"sudo rm -rf /tmp/precheck/error.log",
			fmt.Sprintf("sudo kill -9 `cat /tmp/precheck/%s`", httppid),
			fmt.Sprintf("sudo kill -9 `cat /tmp/precheck/%s`", checkpid),
		},
	}

	envInit3 := remote.Command{
		Cmds: []string{fmt.Sprintf("transform kubelet -v %s -r %s"), op.KubeVersion, op.Runtime},
	}

	cleanCmd := remote.Command{
		Cmds: []string{fmt.Sprintf("sudo kill -9 `cat /tmp/precheck/%s`", httppid), "sudo rm -rf /tmp/precheck"},
	}

	result := remote.Run(configuration.Instance.Hosts, envInit1)
	if len(result) > 0 {
		errs := "环境清理失败: "
		for key, value := range result {
			errs += fmt.Sprintf("%s,%s ", key, strings.Join(value, ""))
		}
		log.Info(errs)
		generateErrorReport(startTime, []report.CaseInfo{}, 0, errors.New(errs))
		return
	}

	log.Info("开始分发检查文件...")
	configName := path.Base(op.File)
	if len(AMD64Host) > 0 || len(ARM64Host) > 0 {
		log.Info("分发文件到各个节点...")
		pwd, _ := os.Getwd()
		result = remote.Run(AMD64Host, disPatchScript(pwd+"/transform_amd64", op.File, "transform_amd64", configName))
		if len(result) > 0 {
			for key, value := range result {
				log.Info(fmt.Sprintf("%s: %s", key, strings.Join(value, "")))
			}
			generateErrorReport(startTime, []report.CaseInfo{}, 0, errors.New("分发文件失败"))
			return
		}
		result = remote.Run(ARM64Host, disPatchScript(pwd+"/transform_arm64", op.File, "transform_arm64", configName))
		if len(result) > 0 {
			for key, value := range result {
				log.Info(fmt.Sprintf("%s: %s", key, strings.Join(value, "")))
			}
			generateErrorReport(startTime, []report.CaseInfo{}, 0, errors.New("分发文件失败"))
			return
		}
	} else {
		// 获取本服务二进制文件
		exePath, err := os.Executable()
		if err != nil {
			log.Info(err.Error())
			generateErrorReport(startTime, []report.CaseInfo{}, 0, err)
			return
		}
		fileName := path.Base(exePath)
		result = remote.Run(configuration.Instance.Hosts, disPatchScript(exePath, op.File, fileName, configName))
		if len(result) > 0 {
			for key, value := range result {
				log.Info(fmt.Sprintf("%s: %s", key, strings.Join(value, "")))
			}
			generateErrorReport(startTime, []report.CaseInfo{}, 0, errors.New("分发文件失败"))
			return
		}
	}

	result = remote.Run(configuration.Instance.Hosts, envInit3)
	if len(result) > 0 {
		for key, value := range result {
			log.Info(fmt.Sprintf("%s: %s", key, strings.Join(value, "")))
		}
		generateErrorReport(startTime, []report.CaseInfo{}, 0, errors.New("启动检查失败"))
		return
	}

	// 第四步：定时搜索节点，检查执行结果
	cycleCmd := []string{fmt.Sprintf("sudo ls /proc/`cat /tmp/precheck/%s`/exe", checkpid), "ls /tmp/precheck"}
	n := 0
	noResultReport := []report.ReportData{}
	// 循环等待直到完成
	for {
		if len(nodeTaskMap) == n {
			break
		}
		time.Sleep(15 * time.Second)
		for k, v := range nodeTaskMap {
			if _, ok := nodeResultMap[k]; ok {
				continue
			}
			log.Info(fmt.Sprintf("周期巡检节点：%s", k))
			stdOut, stdErr, err := v.cli.SSH.Exec(cycleCmd[0])
			log.Info(fmt.Sprintf("巡检输出%s", stdOut))
			if err != nil {
				log.Info(err.Error())
				generateErrorReport(startTime, []report.CaseInfo{}, 0, err)
				return
			}
			if len(stdOut) > 0 {
				continue
			}
			log.Info(fmt.Sprintf("已经完成检查节点: %s", k))
			stdOut, stdErr, err = v.cli.SSH.Exec(cycleCmd[1])
			if err != nil {
				log.Info(err.Error())
				generateErrorReport(startTime, []report.CaseInfo{}, 0, err)
				return
			}
			if len(stdErr) > 0 {
				log.Info(strings.Join(stdErr, ""))
				generateErrorReport(startTime, []report.CaseInfo{}, 0, errors.New(strings.Join(stdErr, "")))
				return
			}
			log.Info(fmt.Sprintf("目标节点%s, 目录文件列表：%s", k, stdOut))
			for _, s := range stdOut {
				if s == resultFile {
					err = v.cli.SFTP.DownloadFile(fmt.Sprintf("/tmp/report/%s.yaml", k), fmt.Sprintf("/tmp/precheck/%s", resultFile))
					if err != nil {
						log.Info(err.Error())
						generateErrorReport(startTime, []report.CaseInfo{}, 0, err)
						return
					}
					nodeResultMap[k] = 0
					n += 1
					log.Info(fmt.Sprintf("测试成功，测试结果收集完成: %s", k))
				}
				if s == errorFile {
					err = v.cli.SFTP.DownloadFile(fmt.Sprintf("/tmp/report%s.errorlog", k), fmt.Sprintf("/tmp/precheck/%s", errorFile))
					if err != nil {
						log.Info(err.Error())
						generateErrorReport(startTime, []report.CaseInfo{}, 0, err)
						return
					}
					nodeResultMap[k] = 0
					n += 1
					log.Info(fmt.Sprintf("测试失败，请查看/tmp/report/%s.errorlog", k))
				}
			}
			if _, ok := nodeResultMap[k]; !ok {
				log.Info(fmt.Sprintf("节点%s收集检查结果失败", k))
				nodeResultMap[k] = 0
				n += 1
				noResultReport = append(noResultReport, report.ReportData{
					StartTime:    "",
					DurationTime: "",
					RandomSeed:   "",
					Total:        1,
					Success:      0,
					Failure:      1,
					Warning:      0,
					Result:       report.NOTPASS,
					Case: []report.CaseInfo{
						{
							IP:           k,
							Name:         "收集测试结果",
							Status:       report.Failure,
							Detail:       "任务执行异常，未能收集到检查结果，请确认用户是否有免密root权限或者其他异常导致结果文件丢失",
							DurationTime: "0",
						},
					},
					Server: nil,
				})
			}
		}
	}
	// 第五步：执行完成，收集结果
	reportList := []report.ReportData{}
	reportList = append(reportList, noResultReport...)
	// 获取/tmp/report目录下所有文件
	files, err := os.ReadDir("/tmp/report")
	if err != nil {
		log.Info(err.Error())
		generateErrorReport(startTime, []report.CaseInfo{}, 0, err)
		return
	}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		if strings.HasSuffix(file.Name(), ".errorlog") {
			// 读取文件并序列化
			data, err := os.ReadFile(fmt.Sprintf("/tmp/report/%s", file.Name()))
			if err != nil {
				log.Info(err.Error())
				continue
			}
			// 文件名称，切去.errorlog后缀
			rep := report.ReportData{
				StartTime:    "",
				DurationTime: "",
				RandomSeed:   "",
				Total:        1,
				Success:      0,
				Failure:      1,
				Warning:      0,
				Result:       report.NOTPASS,
				Case: []report.CaseInfo{
					{
						IP:           strings.TrimSuffix(file.Name(), ".errorlog"),
						Name:         "任务执行报错",
						Status:       report.Failure,
						Detail:       string(data),
						DurationTime: "0",
					},
				},
				Server: nil,
			}
			reportList = append(reportList, rep)
			continue
		}
		if strings.HasSuffix(file.Name(), ".yaml") {
			// 读取文件并序列化
			data, err := os.ReadFile(fmt.Sprintf("/tmp/report/%s", file.Name()))
			if err != nil {
				log.Info(err.Error())
				continue
			}
			var rep report.ReportData
			err = yaml.Unmarshal(data, &rep)
			if err != nil {
				log.Info(err.Error())
				continue
			}
			reportList = append(reportList, rep)
			continue
		}
	}
	err = report.GenerateReport(startTime, reportList)
	if err != nil {
		log.Info(err.Error())
		generateErrorReport(startTime, []report.CaseInfo{}, 0, err)
		return
	}
	// 第六步：生成报告，清理各个节点
	_ = os.RemoveAll("/tmp/report")
	result = remote.Run(configuration.Instance.Hosts, cleanCmd)
	if len(result) > 0 {
		for key, value := range result {
			log.Info(fmt.Sprintf("%s: %s", key, strings.Join(value, "")))
		}
		return
	}
	log.Info(fmt.Sprintf("预检报告生成完成: report.html"))
}

// ConfigValidation 配置校验
func (op *Options) ConfigValidation() ([]report.CaseInfo, int, error) {
	reportCase := []report.CaseInfo{}
	b, err := os.ReadFile(op.File)
	if err != nil {
		reportCase = append(reportCase, report.CaseInfo{
			Identify:     "public",
			IP:           "",
			Role:         "配置检查",
			Name:         "读取配置文件出错",
			Status:       report.Failure,
			Detail:       err.Error(),
			DurationTime: "0",
		})
		return reportCase, 1, err
	}
	err = yaml.Unmarshal(b, &configuration.Instance)
	if err != nil {
		reportCase = append(reportCase, report.CaseInfo{
			Identify:     "public",
			IP:           "",
			Role:         "配置检查",
			Name:         "解析配置文件出错",
			Status:       report.Failure,
			Detail:       err.Error(),
			DurationTime: "0",
		})
		return reportCase, 1, err
	}


	hostMap := map[string]uint8{}
	for _, node := range configuration.Instance.Hosts {
		hostMap[node.IP] = 0
	}
	if len(hostMap) < len(configuration.Instance.Hosts) {
		reportCase = append(reportCase, report.CaseInfo{
			Identify:     "public",
			IP:           "",
			Role:         "配置检查",
			Name:         "主机IP地址重复",
			Status:       report.Failure,
			Detail:       "主机IP地址重复",
			DurationTime: "0",
		})
		return reportCase, 1, errors.New("主机IP地址重复")
	}

	mutilArch := true

	amdName := fmt.Sprintf("%s-amd64", transformName)
	err = utils.DownloadFile(op.HttpRepo+amdName, "./")
	if err != nil {
		log.Error(err)
	}
	log.Info("wget name success")

	armName := fmt.Sprintf("%s-arm64", transformName)
	err = utils.DownloadFile(op.HttpRepo+armName, "./")
	if err != nil {
		log.Error(err)
	}
	log.Info("wget name success")

	if utils.Exists(amdName) && utils.Exists(armName) {
		mutilArch = true
	}
	errNum := 0
	for _, node := range configuration.Instance.Hosts {
		startTime := time.Now()
		if len(node.UserName) == 0 && len(node.Password) == 0 {
			log.Info(fmt.Sprintf("用户名密码均为空，不检查节点：%s", node.IP))
			continue
		}
		if node.IP == "" {
			errNum++
			reportCase = append(reportCase, report.CaseInfo{
				Identify:     "ip",
				IP:           node.IP,
				Role:         "配置检查",
				Name:         "主机IP地址为空",
				Status:       report.Failure,
				Detail:       "主机IP地址为空",
				DurationTime: "0",
			})
		}
		if node.UserName == "" {
			errNum++
			reportCase = append(reportCase, report.CaseInfo{
				Identify:     "ip",
				IP:           node.IP,
				Role:         "配置检查",
				Name:         "userName为空",
				Status:       report.Failure,
				Detail:       "userName为空",
				DurationTime: "0",
			})
		}
		if node.Password == "" {
			errNum++
			reportCase = append(reportCase, report.CaseInfo{
				Identify:     "ip",
				IP:           node.IP,
				Role:         "配置检查",
				Name:         "password为空",
				Status:       report.Failure,
				Detail:       "password为空",
				DurationTime: "0",
			})
		}
		if node.Port == "" {
			errNum++
			reportCase = append(reportCase, report.CaseInfo{
				Identify:     "ip",
				IP:           node.IP,
				Role:         "配置检查",
				Name:         "port为空",
				Status:       report.Failure,
				Detail:       "port为空",
				DurationTime: "0",
			})
		}

		cli, err := remote.NewRemoteClient(&configuration.Host{
			IP:       node.IP,
			UserName: node.UserName,
			Password: node.Password,
			Port:     node.Port,
		})

		if err != nil || cli == nil {
			errNum++
			detail := "建立ssh连接失败"
			if err != nil {
				detail = err.Error()
			}
			reportCase = append(reportCase, report.CaseInfo{
				Identify:     "ip",
				IP:           node.IP,
				Role:         "配置检查",
				Name:         "ssh连接",
				Status:       report.Failure,
				Detail:       detail,
				DurationTime: time.Now().Sub(startTime).String(),
			})
		} else {
			reportCase = append(reportCase, report.CaseInfo{
				Identify:     "ip",
				IP:           node.IP,
				Role:         "配置检查",
				Name:         "ssh连接",
				Status:       report.Success,
				Detail:       "建立ssh连接成功",
				DurationTime: "0",
			})
			if mutilArch {
				stdOut, stdErr, err := cli.SSH.Exec("echo $(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/;s/^unknown$/amd64/')")
				if err != nil || len(stdErr) > 0 || len(stdOut) == 0 {
					errMessage := ""
					if err != nil {
						errMessage += err.Error()
					}
					if len(stdErr) > 0 {
						errMessage += strings.Join(stdErr, ",")
					}
					reportCase = append(reportCase, report.CaseInfo{
						Identify:     "arch",
						IP:           node.IP,
						Role:         "配置检查",
						Name:         "获取目标系统架构",
						Status:       report.Failure,
						Detail:       errMessage,
						DurationTime: "0",
					})
				}
				if stdOut[0] == "arm64" {
					ARM64Host = append(ARM64Host, node)
				} else {
					AMD64Host = append(AMD64Host, node)
				}
			}
		}
		nodeTaskMap[node.IP] = nodeTask{
			ip:   node.IP,
			cli:  cli,
		}
	}
	if errNum > 0 {
		return reportCase, errNum, errors.New("配置文件校验失败")
	}

	return reportCase, 0, nil
}



func writeErrorFile(err error) {
	_ = os.WriteFile(errorFile, []byte(err.Error()), 0644)
}



func generateErrorReport(startTime time.Time, res []report.CaseInfo, errNum int, er error) {
	total := len(res)
	if er != nil {
		total += 1
	}
	failure := errNum
	if er != nil {
		failure += 1
	}
	if er != nil {
		res = append(res, report.CaseInfo{
			Identify:     "error",
			IP:           "主控节点",
			Role:         "主控节点",
			Name:         "异常错误",
			Status:       report.Failure,
			Detail:       er.Error(),
			DurationTime: "0",
		})
	}

	data := report.ReportData{
		StartTime:    startTime.Format("2006-01-02 15:04:05"),
		DurationTime: "",
		RandomSeed:   "",
		Total:        total,
		Success:      total - failure,
		Failure:      failure,
		Warning:      0,
		Result:       report.NOTPASS,
		Case:         res,
		Server:       make([]report.ServerInfo, 0),
	}
	err := report.GenerateReport(startTime, []report.ReportData{data})
	if err != nil {
		log.Info(err.Error())
		return
	}
}

// dispatch script 分发文件
func disPatchScript(binary, conf, binaryName, confName string) remote.Command {
	return remote.Command{
		FileUp: []remote.File{
			{
				Src: binary,
				Dst: "/tmp/precheck",
			},
			{
				Src: conf,
				Dst: "/tmp/precheck/",
			},
		},
		Cmds: []string{
			fmt.Sprintf("sudo mv /tmp/precheck/%s /tmp/precheck/transform", binaryName),
			fmt.Sprintf("sudo mv /tmp/precheck/%s /tmp/precheck/nodes.yaml", confName),
			"sudo chmod +x /tmp/precheck/transform",
			//"cd /tmp/precheck; sudo nohup ./bc run check >/dev/null 2>&1 &",
		},
	}
}

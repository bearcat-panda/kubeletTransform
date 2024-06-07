package cmd

import (
	"errors"
	"github.com/spf13/cobra"
	"transform/pkg/batch"
	"transform/utils/log"
)

var batchOption batch.Options

var batchCmd = &cobra.Command{
	Use:   "batch",
	Short: "批量执行kubelet任务转换",
	Long:  `批量执行kubelet任务转换.`,
	Example: `
# 启动服务器资源检查
transform batch --file nodes.yaml
`,
	Args: func(cmd *cobra.Command, args []string) error {
		if batchOption.File == "" {
			log.Error("The `file` parameter is required. ")
			return errors.New("The `file` parameter is required. ")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		batchOption.Args = args
		batchOption.Options = options
		batchOption.Run()
	},
}

func init() {
	rootCmd.AddCommand(batchCmd)

	batchCmd.PersistentFlags().StringVarP(&batchOption.File, "file", "f", "", "服务器配置列表")
	batchCmd.Flags().StringVarP(&kubeletOption.HttpRepo, "http-repo", "p", "http://deploy.bocloud.k8s:40080/files/", "Kubelet file storage address. example http://deploy.bocloud.k8s:40080/files/ ")
	batchCmd.Flags().StringVarP(&kubeletOption.KubeVersion, "kubernetes-version", "v", "", "The version of kubernetes. For example, 1.21.13/1.26.15")
	batchCmd.Flags().StringVarP(&kubeletOption.Runtime, "runtime", "r", "", "The type of runtime. For example, docker/containerd")
}
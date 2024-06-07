package cmd

import (
	"github.com/spf13/cobra"
	"transform/pkg/kubelet"
)

var kubeletOption = kubelet.Options{}

// kubeletCmd represents the kubelet converted
var kubeletCmd = &cobra.Command{
	Use:   "kubelet",
	Short: "Convert container-type kubelet to binary-type kubelet",
	Long:  `The container type kubelet is automatically converted to the binary type kubelet.
Contains containers of docker and containerd types`,
	Example: `
# Resetting the boot node
transform kubelet -v 1.21.13 -r docker
transform kubelet -v 1.21.13 -r containerd
transform kubelet -v 1.26.15 -r containerd
`,
	Run: func(cmd *cobra.Command, args []string) {
		kubeletOption.Options = options
		kubeletOption.Args = args
		kubeletOption.Reset()
	},
}

func init() {
	rootCmd.AddCommand(kubeletCmd)

	// Here you will define your flags and configuration settings.
	kubeletCmd.Flags().StringVarP(&kubeletOption.HttpRepo, "http-repo", "p", "http://deploy.bocloud.k8s:40080/files/", "Kubelet file storage address. example http://deploy.bocloud.k8s:40080/files/ ")
	kubeletCmd.Flags().StringVarP(&kubeletOption.KubeVersion, "kubernetes-version", "v", "", "The version of kubernetes. For example, 1.21.13/1.26.15")
	kubeletCmd.Flags().StringVarP(&kubeletOption.Runtime, "runtime", "r", "", "The type of runtime. For example, docker/containerd")
	kubeletCmd.Flags().Int64VarP(&kubeletOption.Timeout, "timeout", "t", 2, "timout. default is 2 minute")
}
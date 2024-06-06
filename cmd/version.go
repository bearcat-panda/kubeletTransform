/*
Copyright © 2022 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"transform/utils/log"
	"transform/utils/version"
)

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "version",
	Long:  `bke version.`,
	Example: `
# View the BKE version
bke version
`,
	Run: func(cmd *cobra.Command, args []string) {
		log.BKEFormat("", fmt.Sprintf("version: %s", version.Version))
		log.BKEFormat("", fmt.Sprintf("gitCommitID: %s", version.GitCommitID))
		log.BKEFormat("", fmt.Sprintf("os/arch: %s", version.Architecture))
		log.BKEFormat("", fmt.Sprintf("date: %s", version.Timestamp))
	},
}

// onlyCmd represents the version command
var onlyCmd = &cobra.Command{
	Use:   "only",
	Short: "only",
	Long:  `bke version only.`,
	Example: `
# View the BKE version
bke version only
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(version.Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	versionCmd.AddCommand(onlyCmd)
}

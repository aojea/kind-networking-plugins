/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

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
	"strings"

	"github.com/spf13/cobra"

	"sigs.k8s.io/kind/pkg/cluster"
	kindcmd "sigs.k8s.io/kind/pkg/cmd"
)

// getCmd represents the get command
var getCmd = &cobra.Command{
	Use:   "get",
	Short: "Get the clusters that belong to the multi cluster",
	Long:  `Get the clusters that belong to the multi cluster`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return getMultiCluster(cmd)
	},
}

func init() {
	rootCmd.AddCommand(getCmd)

	getCmd.Flags().String(
		"name",
		cluster.DefaultName,
		"the multicluster context name",
	)
}

func getMultiCluster(cmd *cobra.Command) error {
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}
	logger := kindcmd.NewLogger()

	provider := cluster.NewProvider(
		cluster.ProviderWithLogger(logger),
	)
	clusters, err := provider.List()
	if err != nil {
		return err
	}
	if len(clusters) == 0 {
		logger.V(0).Info("No kind clusters found.")
		return nil
	}

	var inClusters []string
	clusterNamePrefix := fmt.Sprintf("multi-%s-", name)
	for _, cluster := range clusters {
		if strings.Contains(cluster, clusterNamePrefix) {
			inClusters = append(inClusters, cluster)
		}
	}
	fmt.Printf("Multicluster %s contain following clusters: %v\n", name, inClusters)
	// TODO: accumulate errors
	return nil
}

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
	"strings"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"sigs.k8s.io/kind/pkg/cluster"
	kindcmd "sigs.k8s.io/kind/pkg/cmd"
	"sigs.k8s.io/kind/pkg/exec"

	"github.com/aojea/kind-networking-plugins/pkg/docker"
)

// deleteCmd represents the delete command
var deleteCmd = &cobra.Command{
	Use:   "delete",
	Short: "Delete the specified multicluster",
	Long:  `Delete the specified multicluster`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return deleteMultiCluster(cmd)
	},
}

func init() {
	rootCmd.AddCommand(deleteCmd)

	deleteCmd.Flags().String(
		"name",
		cluster.DefaultName,
		"the multicluster context name",
	)

	deleteCmd.Flags().String(
		"config",
		"./config.yml",
		"the config file with the cluster configuration",
	)
	deleteCmd.MarkFlagRequired("config")
}

func deleteMultiCluster(cmd *cobra.Command) error {
	name, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}
	configPath, err := cmd.Flags().GetString("config")
	if err != nil {
		return err
	}
	cfg, err := NewConfig(configPath)
	if err != nil {
		return err
	}

	deleteWanem(name)
	logger := kindcmd.NewLogger()

	provider := cluster.NewProvider(
		cluster.ProviderWithLogger(logger),
	)
	clusters, err := provider.List()
	if err != nil {
		return err
	}

	for clusterName, _ := range cfg.Clusters {

		for _, cluster := range clusters {
			if strings.Contains(cluster, clusterName) {
				if err = provider.Delete(cluster, ""); err != nil {
					logger.V(0).Infof("%s\n", errors.Wrapf(err, "failed to delete cluster %q", cluster))
					continue
				}
				logger.V(0).Infof("Deleted clusters: %q", cluster)
			}
		}
		networks, err := docker.ListNetwork()
		if err != nil {
			return err
		}
		for _, network := range networks {
			if strings.Contains(network, clusterName) {
				if err = docker.DeleteNetwork(network); err != nil {
					logger.V(0).Infof("%s\n", errors.Wrapf(err, "failed to delete network %q", network))
					continue
				}
			}
		}
	}
	// TODO accumulate errors
	return nil
}

func deleteWanem(name string) error {
	containerName := "wan-" + name
	return exec.Command("docker", "rm", "-f", containerName).Run()
}

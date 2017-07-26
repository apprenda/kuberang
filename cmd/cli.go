package main

import (
	"io"

	"github.com/apprenda/kuberang/pkg/config"
	"github.com/apprenda/kuberang/pkg/kuberang"
	"github.com/spf13/cobra"
)

// NewKuberangCommand creates the kuberang command
func NewKuberangCommand(version string, in io.Reader, out io.Writer) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kuberang",
		Short: "kuberang tests your kubernetes cluster using kubectl",
		RunE: func(cmd *cobra.Command, args []string) error {
			return doCheckKubernetes()
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	cmd.PersistentFlags().StringVarP(&config.Namespace, "namespace", "n", "",
		"Kubernetes namespace in which kuberang will operate. Defaults to 'default' if not specified.")
	cmd.PersistentFlags().StringVar(&config.RegistryURL, "registry-url", "",
		"Override the default Docker Hub URL to use a local offline registry for required Docker images.")
	cmd.Flags().BoolVar(&config.SkipCleanup, "skip-cleanup", false, "Don't clean up. Leave all deployed artifacts running on the cluster.")
	cmd.Flags().BoolVar(&config.SkipDNSTests, "skip-dns-tests", false, "Don't test kubernetes DNS if none is deployed.")
	cmd.Flags().BoolVar(&config.IgnorePodIPAccessibilityCheck, "ignore-pod-ip-accessibility-check", false, "Don't fail the smoke test if the pod IP accessibility check fails.")
	cmd.AddCommand(NewCmdVersion(out))

	return cmd
}

func doCheckKubernetes() error {
	return kuberang.CheckKubernetes()
}

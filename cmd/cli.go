package main

import (
	"io"

	"github.com/apprenda/kuberang/pkg/config"
	"github.com/apprenda/kuberang/pkg/kuberang"
	"github.com/spf13/cobra"
)

// NewKuberangCommand creates the kuberang command
func NewKuberangCommand(version string, in io.Reader, out io.Writer) *cobra.Command {
	var outputFormat string
	cmd := &cobra.Command{
		Use:   "kuberang",
		Short: "kuberang tests your kubernetes cluster using kubectl",
		RunE: func(cmd *cobra.Command, args []string) error {
			return doCheckKubernetes(outputFormat)
		},
		SilenceUsage:  true,
		SilenceErrors: true,
	}
	cmd.PersistentFlags().StringVarP(&config.Namespace, "namespace", "n", "",
		"Kubernetes namespace in which kuberang will operate. Defaults to 'default' if not specified.")
	cmd.Flags().StringVarP(&outputFormat, "output-format", "o", "simple", "set the output format. One of: simple|json")
	return cmd
}

func doCheckKubernetes(outputFormat string) error {
	return kuberang.CheckKubernetes(outputFormat)
}

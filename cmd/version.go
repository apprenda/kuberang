package main

import (
	"encoding/json"
	"fmt"
	"io"
	"runtime"

	"github.com/spf13/cobra"
)

type versionOut struct {
	Version   string
	BuildDate string
	GoVersion string
}

// NewCmdVersion returns the version command
func NewCmdVersion(out io.Writer) *cobra.Command {
	var outFormat string
	cmd := &cobra.Command{
		Use:   "version",
		Short: "display the Kismatic CLI version",
		RunE: func(cmd *cobra.Command, args []string) error {
			v := versionOut{
				Version:   version,
				BuildDate: buildDate,
				GoVersion: runtime.Version(),
			}
			if outFormat == "json" {
				b, err := json.MarshalIndent(v, "", "    ")
				if err != nil {
					return fmt.Errorf("error marshaling data: %v", err)
				}
				fmt.Fprintf(out, string(b))
				return nil
			}
			fmt.Fprintln(out, "Kuberang:")
			fmt.Fprintf(out, "  Version: %s\n", version)
			fmt.Fprintf(out, "  Built: %s\n", buildDate)
			fmt.Fprintf(out, "  Go Version: %s\n", runtime.Version())
			return nil
		},
	}
	cmd.Flags().StringVarP(&outFormat, "output", "o", "simple", `output format (options "simple"|"json")`)
	return cmd
}

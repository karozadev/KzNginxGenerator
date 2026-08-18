package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Affiche la version de kznginx",
	RunE: func(cmd *cobra.Command, args []string) error {
		out := cmd.OutOrStdout()
		if _, err := fmt.Fprintf(out, "kznginx %s", Version); err != nil {
			return err
		}
		if GitCommit != "" && GitCommit != "none" {
			if _, err := fmt.Fprintf(out, " (%s)", GitCommit); err != nil {
				return err
			}
		}
		_, err := fmt.Fprintln(out)
		return err
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

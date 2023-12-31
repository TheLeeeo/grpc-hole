package servicecmd

import (
	"github.com/spf13/cobra"
)

func init() {
	ServiceCmd.AddCommand(ScanCmd)
	ServiceCmd.AddCommand(LoadCmd)
	ServiceCmd.AddCommand(ListCmd)
	ServiceCmd.AddCommand(InfoCmd)
	ServiceCmd.AddCommand(CreateDefaultsCmd)
}

var ServiceCmd = &cobra.Command{
	Use:     "service",
	Aliases: []string{"services"},
	Short:   "Manage services",
	Run: func(cmd *cobra.Command, args []string) {
		_ = cmd.Help()
	},
}

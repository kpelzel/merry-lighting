package cmd

import (
	"github.com/kpelzel/merry-lighting/internal"
	"github.com/spf13/cobra"
)

var (
	scanCmd = &cobra.Command{
		Use:   "scan",
		Short: "Scan for available BLE Devices",
		Long:  "Scan for available BLE Devices",
		Run: func(cmd *cobra.Command, args []string) {
			internal.Scan()
		},
	}
)

func init() {
	RootCmd.AddCommand(scanCmd)
}

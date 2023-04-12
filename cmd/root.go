package cmd

import (
	"os"

	"github.com/kpelzel/merry-lighting/internal"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	debug bool

	RootCmd = &cobra.Command{
		Use:   "merry-lighting",
		Short: "run merry-lighting",
		Long:  "run merry-lighting",
		Run: func(cmd *cobra.Command, args []string) {
			internal.StartMerryLighting(debug)
		},
	}
)

func Execute() {
	if err := RootCmd.Execute(); err != nil {
		logrus.Errorf("failed to execute command: %v", err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().BoolVarP(&debug, "debug", "d", false, "enable debugging")
}

package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "xon",
	Short: "A W.I.P cli to check your breached data with XON",
	Long: `Checks your email and password with XON

XON cli support checking creds mannually or
from a password manager export in CSV format`,
}

func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

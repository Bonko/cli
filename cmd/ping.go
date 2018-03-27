package cmd

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// pingCmd represents the ping command
var pingCmd = &cobra.Command{
	Use:   "ping",
	Short: "Ping the StormForger API",
	Long:  `Ping the StormForger API and try to authenticate.`,
	Run:   runPing,
}

func runPing(cmd *cobra.Command, args []string) {
	client := NewClient()

	status, err := client.Ping()

	if !status {
		log.Info(err)
		os.Exit(-1)
	} else {
		fmt.Println("pong!")
	}
}

func init() {
	RootCmd.AddCommand(pingCmd)
}

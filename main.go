package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/stormforger/cli/cmd"
)

func main() {
	log.SetLevel(log.DebugLevel)
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})

	cmd.Execute()
}

package main

import (
	"os"

	"github.com/khulnasoft/scaffold/toolkit/cmd"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.InfoLevel)
}

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

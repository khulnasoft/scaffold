package main

import (
	"github.com/khulnasoft/scaffold/toolkit/cmd"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetLevel(log.InfoLevel)
}

func main() {
	cmd.Execute()
}

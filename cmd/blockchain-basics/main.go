package main

import (
	"blockchain-basics/pkg/blockchain"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

func main() {
	cli := blockchain.NewCLI(getLogger(), nil)
	if err := cli.Run(); err != nil {
		panic(err)
	}
}

func getLogger() *logrus.Logger {
	log := logrus.New()
	if strings.ToLower(os.Getenv("VERBOSE")) == "true" {
		log.SetLevel(logrus.DebugLevel)
		log.Debug("log level set to debug")
	}
	log.Formatter = &logrus.JSONFormatter{}
	return log
}

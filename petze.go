package main

import (
	"flag"
	"os"

	"github.com/foomo/petze/service"
	"github.com/foomo/petze/config"
	log "github.com/sirupsen/logrus"
)

func main() {
	flag.Usage = usage
	flag.Parse()
	if flag.NArg() != 1 {
		flag.Usage()
		os.Exit(1)
	}

	configurationDirectory := os.Args[1]
	if stat, err := os.Stat(configurationDirectory); err == nil && stat.IsDir() {
		runServer(configurationDirectory)
	} else {
		log.Fatal("specified configuration directory does not exist or is not a directory")
	}
}

func runServer(configurationDirectory string) {
	serverConfig, err := config.LoadServer(configurationDirectory)
	if err != nil {
		log.Fatal(err)
	}
	log.Info(service.Run(serverConfig, configurationDirectory))
}

func usage() {
	log.Printf("Usage: %s configuration-directory \n", os.Args[0])
	flag.PrintDefaults()
}

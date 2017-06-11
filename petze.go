package main

import (
	"flag"
	"log"
	"os"

	"github.com/foomo/petze/config"
	"github.com/foomo/petze/service"
)

func main() {
	flagConfigDir := flag.String("config-dir", "", "config-dir")
	flag.Parse()

	if *flagConfigDir == "" {
		flag.Usage()
		os.Exit(1)
	}

	serverConfig, err := config.LoadServer(*flagConfigDir)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(service.Run(serverConfig, *flagConfigDir))
}

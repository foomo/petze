package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/foomo/petze/config"
	"github.com/foomo/petze/service"
)

func main() {
	flagServer := flag.String("server", "", "server config file")
	flagPeople := flag.String("people", "", "server config file")
	flagServices := flag.String("services", "", "server config file")
	flag.Parse()

	if len(*flagPeople) == 0 || len(*flagServer) == 0 || len(*flagServices) == 0 {
		fmt.Println("usage", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	serverConfig, err := config.LoadServer(*flagServer)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(service.Run(serverConfig, *flagServices, *flagPeople))
}

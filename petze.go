package main

import (
	"flag"
	"fmt"
	"github.com/foomo/petze/mail"
	"github.com/foomo/petze/sms"
	"github.com/foomo/petze/watch"
	"os"

	"github.com/foomo/petze/config"
	"github.com/foomo/petze/service"
	"github.com/foomo/petze/slack"
	log "github.com/sirupsen/logrus"
)

var flagJsonOutput bool

// Version is set during build via ldflags
var Version string

func main() {
	flag.Usage = usage
	flag.BoolVar(&flagJsonOutput, "json-output", false, "specifies if the logging format is json or not")

	flag.Parse()
	initializeLogger()

	if len(flag.Args()) == 0 {
		log.Fatal("please pass the configuration directory as a first argument")
	}

	// add version to user agent
	watch.SetUserAgentVersion(Version)
	fmt.Println("petze", Version, "starting")

	configurationDirectory := flag.Args()[0]
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
	if serverConfig.SMTP != nil {
		// init mailer
		mail.InitMailer(
			serverConfig.SMTP.Server,
			serverConfig.SMTP.User,
			serverConfig.SMTP.Pass,
			serverConfig.SMTP.From,
			serverConfig.SMTP.Port,
			serverConfig.SMTP.To,
		)
	}
	// init slackbot
	if serverConfig.Slack != "" {
		slack.InitSlackBot(serverConfig.Slack)
	}
	// init SMS
	if serverConfig.Sms != nil {
		sms.InitSMS(serverConfig.Sms)
	}
	log.Info(service.Run(serverConfig, configurationDirectory))
}

func usage() {
	log.Printf("Usage: %s configuration-directory \n", os.Args[0])
	flag.PrintDefaults()
}

func initializeLogger() {
	if flagJsonOutput {
		log.SetFormatter(&log.JSONFormatter{})
	}
}

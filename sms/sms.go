package sms

import (
	"github.com/foomo/petze/config"
)

const timestampFormat = "Mon 2 Jan 2006 15:04:05"

var conf *config.SMS

func IsInitialized() bool {
	return conf != nil
}

func InitSMS(c *config.SMS) {
	conf = c
}

func SendErrors(errs []error, service string) {
	if conf.TwilioSID != "" && conf.TwilioToken != "" {
		SendTwilioSMS(GenerateTwilioErrorSMS(errs, service))
	}
	if conf.SendInBlueAPIKey != "" {
		SendSIB(GenerateSIBErrorSMS(errs, service))
	}
}

func SendResolvedNotification(service string) {
	if conf.TwilioSID != "" && conf.TwilioToken != "" {
		SendTwilioSMS(GenerateTwilioResolvedSMS(service))
	}
	if conf.SendInBlueAPIKey != "" {
		SendSIB(GenerateSIBResolvedSMS(service))
	}
}

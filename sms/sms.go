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

func SendServiceErrors(errs []error, service string) {
	if conf.TwilioSID != "" && conf.TwilioToken != "" {
		SendTwilioSMS(GenerateTwilioServiceErrorSMS(errs, service))
	}
	if conf.SendInBlueAPIKey != "" {
		SendSIB(GenerateSIBErrorSMS(errs, service))
	}
}

func SendHostErrors(errs []error, service string) {
	if conf.TwilioSID != "" && conf.TwilioToken != "" {
		SendTwilioSMS(GenerateTwilioHostErrorSMS(errs, service))
	}
	if conf.SendInBlueAPIKey != "" {
		SendSIB(GenerateSIBErrorSMS(errs, service))
	}
}

func SendServiceErrorResolvedNotification(service string) {
	if conf.TwilioSID != "" && conf.TwilioToken != "" {
		SendTwilioSMS(GenerateTwilioServiceErrorResolvedSMS(service))
	}
	if conf.SendInBlueAPIKey != "" {
		SendSIB(GenerateSIBResolvedSMS(service))
	}
}

func SendHostErrorResolvedNotification(service string) {
	if conf.TwilioSID != "" && conf.TwilioToken != "" {
		SendTwilioSMS(GenerateTwilioHostErrorResolvedSMS(service))
	}
	if conf.SendInBlueAPIKey != "" {
		SendSIB(GenerateSIBResolvedSMS(service))
	}
}

package sms

import (
	"log"
	"strings"
	"time"

	"github.com/kevinburke/twilio-go"
)

type TwilioSMS struct {
	To   string
	Body string
}

func GenerateTwilioServiceErrorSMS(errs []error, service string) []*TwilioSMS {

	var smsArr []*TwilioSMS
	for _, recipient := range conf.To {

		var lines = []string{
			"Dear Admin,",
			"An error with the service " + strings.ToUpper(service) + " occurred:",
			"Timestamp: " + time.Now().Format(timestampFormat),
		}
		if len(errs) > 0 {
			lines = append(lines, "Errors: ")
			for _, e := range errs {
				lines = append(lines, e.Error())
			}
		}
		smsArr = append(smsArr, &TwilioSMS{
			To:   recipient,
			Body: strings.Join(lines, "\n"),
		})
	}

	return smsArr
}

func GenerateTwilioHostErrorSMS(errs []error, service string) []*TwilioSMS {

	var smsArr []*TwilioSMS
	for _, recipient := range conf.To {

		var lines = []string{
			"Dear Admin,",
			"An error with the host " + strings.ToUpper(service) + " occurred:",
			"Timestamp: " + time.Now().Format(timestampFormat),
		}
		if len(errs) > 0 {
			lines = append(lines, "Errors: ")
			for _, e := range errs {
				lines = append(lines, e.Error())
			}
		}
		smsArr = append(smsArr, &TwilioSMS{
			To:   recipient,
			Body: strings.Join(lines, "\n"),
		})
	}

	return smsArr
}

func GenerateTwilioServiceErrorResolvedSMS(service string) []*TwilioSMS {

	var smsArr []*TwilioSMS
	for _, recipient := range conf.To {

		var lines = []string{
			"Dear Admin,",
			"Service " + strings.ToUpper(service) + " is back to normal operation",
			"Timestamp: " + time.Now().Format(timestampFormat),
		}
		smsArr = append(smsArr, &TwilioSMS{
			To:   recipient,
			Body: strings.Join(lines, "\n"),
		})
	}

	return smsArr
}

func GenerateTwilioHostErrorResolvedSMS(service string) []*TwilioSMS {

	var smsArr []*TwilioSMS
	for _, recipient := range conf.To {

		var lines = []string{
			"Dear Admin,",
			"Host " + strings.ToUpper(service) + " is back to normal operation",
			"Timestamp: " + time.Now().Format(timestampFormat),
		}
		smsArr = append(smsArr, &TwilioSMS{
			To:   recipient,
			Body: strings.Join(lines, "\n"),
		})
	}

	return smsArr
}

func SendTwilioSMS(sms []*TwilioSMS) {

	client := twilio.NewClient(conf.TwilioSID, conf.TwilioToken, nil)
	for _, s := range sms {

		// Send a message
		_, err := client.Messages.SendMessage(conf.From, s.To, s.Body, nil)
		if err != nil {
			log.Println("sending twilio sms failed:", err)
		}
		//fmt.Println(msg.Status)
	}
}

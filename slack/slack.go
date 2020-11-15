package slack

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
)

var (
	Log     = logrus.New()
	webhook string
)

func IsInitialized() bool {
	return webhook != ""
}

type Message struct {
	Text string `json:"text"`
}

func init() {
	Log.Formatter = &prefixed.TextFormatter{
		ForceColors:     true,
		ForceFormatting: true,
	}
}

// ConfigureLogger toggles debugging and adds full timestamps in production mode
func ConfigureLogger(level logrus.Level, prod bool) {

	if prod {
		Log.Formatter = &prefixed.TextFormatter{
			ForceColors:     true,
			ForceFormatting: true,
			FullTimestamp:   true,
			TimestampFormat: "Mon 2 Jan 2006 15:04:05",
		}
	}

	Log.Level = level
}

const timestampFormat = "Mon 2 Jan 2006 15:04:05"

func InitSlackBot(w string) {
	webhook = w
}

func Send(message []byte) {
	client := &http.Client{}
	requestBody := bytes.NewReader(message)
	request, err := http.NewRequest("POST", webhook, requestBody)
	if err != nil {
		Log.Error(err)
	}

	request.Header.Add("Content-type", "application/json")

	// execute the HTTP request
	response, err := client.Do(request)
	if err != nil {
		Log.Error(err)
	}
	defer response.Body.Close()

	// read response body
	responseBody, err := ioutil.ReadAll(response.Body)
	if err != nil {
		Log.Error(err.Error())
	}
	Log.Info("slack bot response body: ", string(responseBody))
}

func GenerateServiceErrorMessage(errs []error, service string) []byte {

	var errMessage = []string{
		time.Now().Format(timestampFormat),
		"an error occured for service " + strings.ToUpper(service) + "\n",
	}

	if len(errs) > 0 {
		for _, e := range errs {
			errMessage = append(errMessage, e.Error())
		}
	}

	unmarshalledMessage := &Message{Text: strings.Join(errMessage, " ")}
	marshalledMessage, err := json.Marshal(unmarshalledMessage)
	if err != nil {
		Log.Error(err)
	}

	return marshalledMessage
}

func GenerateHostErrorMessage(errs []error, service string) []byte {

	var errMessage = []string{
		time.Now().Format(timestampFormat),
		"an error occured for host " + strings.ToUpper(service) + "\n",
	}

	if len(errs) > 0 {
		for _, e := range errs {
			errMessage = append(errMessage, e.Error())
		}
	}

	unmarshalledMessage := &Message{Text: strings.Join(errMessage, " ")}
	marshalledMessage, err := json.Marshal(unmarshalledMessage)
	if err != nil {
		Log.Error(err)
	}

	return marshalledMessage
}

func GenerateServiceErrorResolvedNotification(service string) []byte {

	var errMessage = []string{
		time.Now().Format(timestampFormat),
		"Service " + strings.ToUpper(service) + " is back to normal operation!",
	}

	unmarshalledMessage := &Message{Text: strings.Join(errMessage, " ")}
	marshalledMessage, err := json.Marshal(unmarshalledMessage)
	if err != nil {
		Log.Error(err)
	}

	return marshalledMessage
}

func GenerateHostErrorResolvedNotification(service string) []byte {

	var errMessage = []string{
		time.Now().Format(timestampFormat),
		"Host " + strings.ToUpper(service) + " is back to normal operation!",
	}

	unmarshalledMessage := &Message{Text: strings.Join(errMessage, " ")}
	marshalledMessage, err := json.Marshal(unmarshalledMessage)
	if err != nil {
		Log.Error(err)
	}

	return marshalledMessage
}

package sms

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

const (
	endpoint = "https://api.sendinblue.com/v3/transactionalSMS/sms"
)

var (
	ErrInvalidResponseData = errors.New("invalid SIBResponseData")
	ErrInvalidMessageID = errors.New("message-id field is not a string")
	ErrInvalidData = errors.New("unable to convert data to map[string]interface{}")
)

// V3 API
// https://developers.sendinblue.com/docs
type SendInBlueSMS struct {

	// required fields
	From    string `json:"sender"`
	To      string `json:"recipient"`
	Content string `json:"content"`

	// optional fields
	Tag  string `json:"tag"`
	Type string `json:"type"`
	URL  string `json:"URL"`
}

func GenerateSIBErrorSMS(errs []error, service string) []*SendInBlueSMS {

	var smsArr []*SendInBlueSMS
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
		smsArr = append(smsArr, &SendInBlueSMS{
			From:    conf.From,
			To:      recipient,
			Content: strings.Join(lines, "\n"),
			Tag:     service,
			Type:    "transactional",
		})
	}

	return smsArr
}

func GenerateSIBResolvedSMS(service string) []*SendInBlueSMS {

	var smsArr []*SendInBlueSMS
	for _, recipient := range conf.To {

		var lines = []string{
			"Dear Admin,",
			"service " + strings.ToUpper(service) + " is back to normal operation",
			"Timestamp: " + time.Now().Format(timestampFormat),
		}

		smsArr = append(smsArr, &SendInBlueSMS{
			From:    conf.From,
			To:      recipient,
			Content: strings.Join(lines, "\n"),
			Tag:     service,
			Type:    "transactional",
		})
	}

	return smsArr
}

func SendSIB(smsArr []*SendInBlueSMS) {

	for _, sms := range smsArr {
		resp, err := sendSIBSMS(sms)
		if err != nil {
			log.Println(err)

			return
		}

		rd, err := resp.GetSIBResponseData()
		if err != nil {
			log.Println(err)

			return
		}

		spew.Dump(rd)
	}
}

func sendSIBSMS(sms *SendInBlueSMS) (*SendInBlueResponse, error) {

	body := &bytes.Buffer{}
	defer body.Reset()

	encoder := json.NewEncoder(body)

	err := encoder.Encode(sms)
	if err != nil {
		return nil, err
	}

	res, err := sendSIBRequest(endpoint, nil, ioutil.NopCloser(body), body.Len())
	if err != nil {
		if res != nil {
			c, _ := ioutil.ReadAll(res.Body)
			fmt.Println(string(c))
		}

		return nil, err
	}

	defer func() {
		// Drain and close the body to let the Transport reuse the connection
		io.Copy(ioutil.Discard, res.Body)
		res.Body.Close()
	}()

	rawResBody, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode >= 400 {
		return nil, fmt.Errorf("failed to send SMS: %s", res.Status)
	}

	resp := &SendInBlueResponse{}

	err = json.Unmarshal(rawResBody, resp)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func sendSIBRequest(url string, headers map[string]string, body io.ReadCloser, length int) (*http.Response, error) {

	req, err := http.NewRequest("POST", url, body)
	if err != nil {
		return nil, err
	}

	req.ContentLength = int64(length)

	for key, val := range headers {
		req.Header.Add(key, val)
	}

	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("api-key", conf.SendInBlueAPIKey)

	client := &http.Client{}

	return client.Do(req)
}

// SendInblue JSON response from the server.
type SendInBlueResponse struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type SIBResponseData struct {
	Status          string            `json:"status"`
	Message         string            `json:"message"`
	NumberSent      int               `json:"number_sent"`
	SMSCount        int               `json:"sms_count"`
	ErrorCode       int               `json:"error_code"`
	CreditsUsed     float64           `json:"credits_used"`
	To              string            `json:"to"`
	Reply           string            `json:"reply"`
	Description     string            `json:"description"`
	BounceType      string            `json:"bounce_type"`
	Reference       map[string]string `json:"reference"`
	RemainingCredit float64           `json:"remaining_credit"`
}

// GetMessageId retrieves the sendinblue message-id.
func (s *SendInBlueResponse) GetMessageId() (string, error) {

	dataInterface, ok := s.Data.(map[string]interface{})
	if !ok {
		return "", fmt.Errorf("invalid data type: %w", ErrInvalidData)
	}

	emailID, ok := dataInterface["message-id"].(string)
	if !ok {
		return "", fmt.Errorf("invalid data type: %w", ErrInvalidMessageID)
	}

	return emailID, nil
}

// GetSIBResponseData retrieves the sendinblue API response.
func (s *SendInBlueResponse) GetSIBResponseData() (*SIBResponseData, error) {
	smsResponse, ok := s.Data.(*SIBResponseData)
	if !ok {
		return nil, fmt.Errorf("invalid data type: %w", ErrInvalidResponseData)
	}

	return smsResponse, nil
}

package sms

// commented out because we dont want to run those tests in the Travis CI on every build

//func TestTwilioSMS(t *testing.T) {
//
//	client := twilio.NewClient(os.Getenv("TWILIO_SID"), os.Getenv("TWILIO_TOKEN"), nil)
//
//	// Send a message
//	msg, err := client.Messages.SendMessage(os.Getenv("TEST_PHONE_FROM"), os.Getenv("TEST_PHONE_TO"), "this is a test", nil)
//	if err != nil {
//		t.Fatal(err)
//	}
//
//	fmt.Println(msg.Status)
//}
//
//func TestSendInBlueSMS(t *testing.T) {
//
//	InitSMS(&config.SMS{
//		SendInBlueAPIKey: os.Getenv("SENDINBLUE_API_KEY"),
//		To:               []string{os.Getenv("TEST_PHONE")},
//		From:             "Petze Test",
//	})
//
//	resp, err := sendSIBSMS(&SendInBlueSMS{
//		From:    os.Getenv("TEST_PHONE_FROM"),
//		To:      os.Getenv("TEST_PHONE_TO"),
//		Content: "this is a test",
//		Type:    "transactional",
//	})
//	if err != nil {
//		t.Fatal(err)
//	}
//	rd, err := resp.GetSIBResponseData()
//	if err != nil {
//		t.Fatal(err)
//	}
//	spew.Dump(rd)
//}

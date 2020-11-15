package mail

import (
	"strings"
	"time"

	"github.com/matcornic/hermes/v2"
	"github.com/sirupsen/logrus"
	prefixed "github.com/x-cray/logrus-prefixed-formatter"
	"gopkg.in/gomail.v2"
)

var (
	m   *Mailer
	Log = logrus.New()
)

func IsInitialized() bool {
	return m != nil
}

const timestampFormat = "Mon 2 Jan 2006 15:04:05"

// Mailer handles sending email to the configured SMTP server
type Mailer struct {
	port     int
	server   string
	user     string
	password string
	from     string
	dialer   *gomail.Dialer
	to       []string // default recipients
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

// InitMailer returns a new mailer instance
func InitMailer(smtpServer, smtpUser, smtpPassword, from string, smtpPort int, to []string) {
	m = &Mailer{
		server:   smtpServer,
		port:     smtpPort,
		user:     smtpUser,
		password: smtpPassword,
		from:     from,
		to:       to,
	}
	m.dialer = gomail.NewDialer(m.server, m.port, m.user, m.password)
}

func SendMails(subject string, mail hermes.Email) {
	for _, recipient := range m.to {
		Send(recipient, subject, mail)
	}
}

// Send handles dispatching an email to the specified receiver
func Send(to string, subject string, mail hermes.Email) {

	cLog := Log.WithFields(logrus.Fields{
		"prefix":  "mailer",
		"from":    m.from,
		"to":      to,
		"subject": subject,
	})

	if to == "" {
		cLog.Error("empty receiver email address")
		return
	}

	html, _, errRender := renderMail(mail)
	if errRender != nil {
		cLog.WithError(errRender).Error("failed to render mail")
		return
	}

	cLog.Info("sending email")

	if m.server == "fake" {
		cLog.Info("running locally, not sending email")
		return
	}

	msg := gomail.NewMessage()
	msg.SetAddressHeader("From", m.from, "Petze Mailservice")
	msg.SetHeader("To", to)
	msg.SetHeader("Subject", subject)
	msg.SetBody("text/html", html)

	// msg.Attach(filename string, settings ...gomail.FileSetting)
	// msg.SetBody("text/plain", plainText)

	err := m.dialer.DialAndSend(msg)
	if err != nil {
		cLog.WithError(err).Error("failed to send mail")
		// prevent loop
		if to != m.from {
			// notify grand master
			Send(m.from, "[Mail Error] "+subject+" to "+to, GenerateServiceErrorMail([]error{err}, "failed to send mail", "internal"))
		}
	}
}

func getHermes() hermes.Hermes {
	return hermes.Hermes{
		// Optional Theme
		Theme: new(hermes.Default),
		Product: hermes.Product{
			// Appears in header & footer of e-mails
			Name:      "Petze Mailservice",
			Link:      "",
			Logo:      "",
			Copyright: "Copyright © 2020",
		},
	}
}

func renderMail(mail hermes.Email) (html, plainText string, err error) {

	var (
		h            = getHermes()
		errPlainText error
	)

	html, errHTML := h.GenerateHTML(mail)
	if errHTML != nil {
		err = errHTML
		return
	}

	plainText, errPlainText = h.GeneratePlainText(mail)
	if errPlainText != nil {
		err = errPlainText
		return
	}
	return
}

func GenerateServiceErrorMail(errs []error, msg string, service string) hermes.Email {

	var intros = []string{
		"An error with the service " + strings.ToUpper(service) + " occurred:",
		"Timestamp: " + time.Now().Format(timestampFormat),
	}
	if msg != "" {
		intros = append(intros, "Message: "+msg)
	}
	if len(errs) > 0 {
		intros = append(intros, "Errors: ")
		for _, e := range errs {
			intros = append(intros, e.Error())
		}
	}

	return hermes.Email{
		Body: hermes.Body{
			Greeting:  "Dear",
			Name:      "Admin",
			Signature: "kind regards",
			Intros:    intros,
		},
	}
}

func GenerateHostErrorMail(errs []error, msg string, service string) hermes.Email {

	var intros = []string{
		"An error with the host " + strings.ToUpper(service) + " occurred:",
		"Timestamp: " + time.Now().Format(timestampFormat),
	}
	if msg != "" {
		intros = append(intros, "Message: "+msg)
	}
	if len(errs) > 0 {
		intros = append(intros, "Errors: ")
		for _, e := range errs {
			intros = append(intros, e.Error())
		}
	}

	return hermes.Email{
		Body: hermes.Body{
			Greeting:  "Dear",
			Name:      "Admin",
			Signature: "kind regards",
			Intros:    intros,
		},
	}
}

func GenerateServiceResolvedNotificationMail(service string) hermes.Email {
	return hermes.Email{
		Body: hermes.Body{
			Greeting:  "Dear",
			Name:      "Admin",
			Signature: "kind regards",
			Intros: []string{
				"Service " + strings.ToUpper(service) + " is back to normal operation",
				"Timestamp: " + time.Now().Format(timestampFormat),
			},
		},
	}
}

func GenerateHostResolvedNotificationMail(service string) hermes.Email {
	return hermes.Email{
		Body: hermes.Body{
			Greeting:  "Dear",
			Name:      "Admin",
			Signature: "kind regards",
			Intros: []string{
				"Host " + strings.ToUpper(service) + " is back to normal operation",
				"Timestamp: " + time.Now().Format(timestampFormat),
			},
		},
	}
}

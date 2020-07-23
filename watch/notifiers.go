package watch

import (
	"errors"
	"fmt"
	"github.com/foomo/petze/mail"
	"github.com/foomo/petze/slack"
	"github.com/foomo/petze/sms"
)

func (w *Watcher) smsNotify(r *Result) {

	// if SMS notifications are not enabled, return immediately
	if !sms.IsInitialized() {
		return
	}

	// if there are errors
	// send emails to all people to be notified
	if len(r.Errors) > 0 {
		var errs []error
		for _, e := range r.Errors {
			if len(e.Comment) > 0 {
				errs = append(errs, errors.New(fmt.Sprintln("-", e.Error, "type:", e.Type, "comment:", e.Comment)))
			} else {
				errs = append(errs, errors.New(fmt.Sprintln("-", e.Error, "type:", e.Type)))
			}
		}

		if !w.didReceiveSMSNotification || w.didErrorsChange(r) {
			go func() {
				sms.SendErrors(errs, w.service.ID)
			}()
			w.didReceiveSMSNotification = true
			w.lastErrors = r.Errors
		}
	} else {

		if len(w.lastErrors) > 0 {
			// reset boolean when there are no service errors anymore
			w.didReceiveSMSNotification = false
			w.lastErrors = []Error{}

			if w.service.NotifyIfResolved {
				go func() {
					sms.SendResolvedNotification(w.service.ID)
				}()
			}
		}
	}
}

func (w *Watcher) mailNotify(r *Result) {

	// if SMTP notifications are not enabled, return immediately
	if !mail.IsInitialized() {
		return
	}

	// if there are errors
	// send emails to all people to be notified
	if len(r.Errors) > 0 {
		var errs []error
		for _, e := range r.Errors {
			if len(e.Comment) > 0 {
				errs = append(errs, errors.New(fmt.Sprintln("-", e.Error, "type:", e.Type, "comment:", e.Comment)))
			} else {
				errs = append(errs, errors.New(fmt.Sprintln("-", e.Error, "type:", e.Type)))
			}
		}

		if !w.didReceiveMailNotification || w.didErrorsChange(r) {
			go func() {
				mail.SendMails("Error for Service: "+w.service.ID, mail.GenerateErrorMail(errs, "", w.service.ID))
			}()
			w.didReceiveMailNotification = true
			w.lastErrors = r.Errors
		}
	} else {

		if len(w.lastErrors) > 0 {
			// reset boolean when there are no service errors anymore
			w.didReceiveMailNotification = false
			w.lastErrors = []Error{}

			if w.service.NotifyIfResolved {
				go func() {
					mail.SendMails("Issues resolved for service: "+w.service.ID, mail.GenerateResolvedNotificationMail(w.service.ID))
				}()
			}
		}
	}
}

func (w *Watcher) slackNotify(r *Result) {

	// if Slack notifications are not enabled, return immediately
	if !slack.IsInitialized() {
		return
	}

	// if there are errors
	// trigger slack webhook and generate an error summary
	if len(r.Errors) > 0 {
		var errs []error
		for _, e := range r.Errors {
			if len(e.Comment) > 0 {
				errs = append(errs, errors.New(fmt.Sprintln("-", e.Error, "type:", e.Type, "comment:", e.Comment)))
			} else {
				errs = append(errs, errors.New(fmt.Sprintln("-", e.Error, "type:", e.Type)))
			}
		}
		if !w.didReceiveSlackNotification || w.didErrorsChange(r) {
			go func() {
				slack.Send(slack.GenerateErrorMessage(errs, w.service.ID))
			}()
			w.didReceiveSlackNotification = true
			w.lastErrors = r.Errors
		}
	} else {

		if len(w.lastErrors) > 0 {
			// reset boolean when there are no service errors anymore
			w.didReceiveSlackNotification = false
			w.lastErrors = []Error{}

			if w.service.NotifyIfResolved {
				go func() {
					slack.Send(slack.GenerateResolvedNotification(w.service.ID))
				}()
			}
		}
	}
}

func (w *Watcher) didErrorsChange(r *Result) bool {

	// if the number of errors changed, return true
	if len(w.lastErrors) != len(r.Errors) {
		return true
	}

	// compare the location of each error to see if anything changed
	for i, e := range r.Errors {
		// array access via index is safe here because we know the length is identical
		if e.Location != w.lastErrors[i].Location {
			return true
		}
	}

	// all the same
	return false
}

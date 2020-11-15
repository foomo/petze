package watch

import (
	"errors"
	"fmt"

	"github.com/foomo/petze/mail"
	"github.com/foomo/petze/slack"
	"github.com/foomo/petze/sms"
)

func (w *Watcher) smsNotify(r *Result, isService bool, serviceOrHostid string, notifyIfResolved bool) {

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
				if isService {
					sms.SendServiceErrors(errs, serviceOrHostid)
				} else {
					sms.SendHostErrors(errs, serviceOrHostid)
				}
			}()
			w.didReceiveSMSNotification = true
			w.lastErrors = r.Errors
		}
	} else {

		if len(w.lastErrors) > 0 {
			// reset boolean when there are no service errors anymore
			w.didReceiveSMSNotification = false
			w.lastErrors = []Error{}

			if notifyIfResolved {
				go func() {
					if isService {
						sms.SendServiceErrorResolvedNotification(serviceOrHostid)
					} else {
						sms.SendHostErrorResolvedNotification(serviceOrHostid)
					}
				}()
			}
		}
	}
}

func (w *Watcher) mailNotify(r *Result, isService bool, serviceOrHostid string, notifyIfResolved bool) {

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
				if isService {
					mail.SendMails("Error for Service: "+serviceOrHostid, mail.GenerateServiceErrorMail(errs, "", serviceOrHostid))
				} else {
					mail.SendMails("Error for Host: "+serviceOrHostid, mail.GenerateHostErrorMail(errs, "", serviceOrHostid))
				}
			}()
			w.didReceiveMailNotification = true
			w.lastErrors = r.Errors
		}
	} else {

		if len(w.lastErrors) > 0 {
			// reset boolean when there are no service errors anymore
			w.didReceiveMailNotification = false
			w.lastErrors = []Error{}

			if notifyIfResolved {
				go func() {
					if isService {
						mail.SendMails("Issues resolved for service: "+serviceOrHostid, mail.GenerateServiceResolvedNotificationMail(serviceOrHostid))
					} else {
						mail.SendMails("Issues resolved for host: "+serviceOrHostid, mail.GenerateHostResolvedNotificationMail(serviceOrHostid))
					}
				}()
			}
		}
	}
}

func (w *Watcher) slackNotify(r *Result, isService bool, serviceOrHostid string, notifyIfResolved bool) {

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
				if isService {
					slack.Send(slack.GenerateServiceErrorMessage(errs, serviceOrHostid))
				} else {
					slack.Send(slack.GenerateHostErrorMessage(errs, serviceOrHostid))
				}
			}()
			w.didReceiveSlackNotification = true
			w.lastErrors = r.Errors
		}
	} else {

		if len(w.lastErrors) > 0 {
			// reset boolean when there are no service errors anymore
			w.didReceiveSlackNotification = false
			w.lastErrors = []Error{}

			if notifyIfResolved {
				go func() {
					if isService {
						slack.Send(slack.GenerateServiceErrorResolvedNotification(serviceOrHostid))
					} else {
						slack.Send(slack.GenerateHostErrorResolvedNotification(serviceOrHostid))
					}
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

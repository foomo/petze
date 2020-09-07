package watch

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"strings"
	"time"

	"reflect"

	"github.com/foomo/petze/config"

	log "github.com/sirupsen/logrus"
)

var (
	typeDNSConfigErr                = reflect.TypeOf(&net.DNSConfigError{})
	typeDNSErr                      = reflect.TypeOf(&net.DNSError{})
	typeOpErr                       = reflect.TypeOf(&net.OpError{})
	typeX509CertificateInvalidError = reflect.TypeOf(x509.CertificateInvalidError{})
	typeX509HostnameError           = reflect.TypeOf(x509.HostnameError{})
	typeX509SystemRootsError        = reflect.TypeOf(x509.SystemRootsError{})
	typeX509UnknownAuthorityError   = reflect.TypeOf(x509.UnknownAuthorityError{})
)

type ErrorType string

const (
	ErrorInvalidEndpoint           ErrorType = "endpointInvalid"
	ErrorTypeServerTooSlow                   = "serverTooSlow"
	ErrorTypeNotImplemented                  = "notImplemented"
	ErrorTypeUnknownError                    = "unknownError"
	ErrorTypeClientError                     = "clientError"
	ErrorTypeDNS                             = "dns"
	ErrorTypeDNSConfig                       = "dnsConfig"
	ErrorTypeTLSCertificateInvalid           = "tlsCertificateInvalid"
	ErrorTypeTLSHostNameError                = "tlsHostNameError"
	ErrorTypeTLSSystemRootsError             = "tlsSystemRootsError"
	ErrorTypeTLSUnknownAuthority             = "tlsUnknownAutority"
	ErrorTypeWrongHTTPStatusCode             = "wrongHTTPStatus"
	ErrorTypeCertificateIsExpiring           = "certificateIsExpiring"
	ErrorTypeUnexpectedContentType           = "unexpectedContentType"
	ErrorTypeSessionFail                     = "sessionFail"
	ErrorTypeGoQueryMismatch                 = "goqueryMismatch"
	ErrorTypeGoQuery                         = "goQueryGeneralError"
	ErrorTypeDataMismatch                    = "dataMismatch"
	ErrorJsonPath                            = "jsonPathError"
	ErrorRegex                               = "regexError"
	ErrorBadResponseBody                     = "badResponseBody"
	ErrorTypeHeaderMismatch                  = "headerMismatch"
	ErrorTypeRedirectMismatch                = "redirectMismatch"
	ErrorTypeReplyMismatch                   = "replyMismatch"
)

type Error struct {
	Error    string    `json:"error"`
	Type     ErrorType `json:"type"`
	Comment  string    `json:"comment,omitempty"`
	Location string    `json:"location,omitempty"`
}

type Result struct {
	ID        string        `json:"id"`
	Errors    []Error       `json:"errors"`
	Timeout   bool          `json:"timeout"`
	Timestamp time.Time     `json:"timestamp"`
	RunTime   time.Duration `json:"runtime"`
}

func NewResult(id string) *Result {
	return &Result{
		ID:        id,
		Errors:    []Error{},
		Timestamp: time.Now(),
	}
}

func (r *Result) addError(e error, t ErrorType, comment string) {
	r.Errors = addError(r.Errors, e, t, comment)
}

func addError(errors []Error, err error, t ErrorType, comment string) []Error {
	return append(errors, Error{
		Error:   err.Error(),
		Type:    t,
		Comment: comment,
	})
}

type dialerErrRecorder struct {
	errors                     []Error
	unknownErr                 error
	err                        net.Error
	dnsError                   net.Error
	dnsConfigError             net.Error
	tlsCertificateInvalidError *x509.CertificateInvalidError
	tlsHostnameError           *x509.HostnameError
	tlsSystemRootsError        *x509.SystemRootsError
	tlsUnknownAuthorityError   *x509.UnknownAuthorityError
}

type Watcher struct {
	active                      bool
	service                     *config.Service

	// notifications
	didReceiveMailNotification  bool
	didReceiveSlackNotification bool
	didReceiveSMSNotification   bool
	lastErrors                  []Error
}

// Watch create a watcher and start watching
func Watch(service *config.Service, chanResult chan Result) *Watcher {

	w := &Watcher{
		active:  true,
		service: service,
	}
	go w.watchLoop(chanResult)
	return w
}

// Stop watching - beware this is async
func (w *Watcher) Stop() {
	w.active = false
}

func (w *Watcher) LastErrors() []Error {
	if w.lastErrors != nil {
		return w.lastErrors
	}
	return nil
}

func (w *Watcher) SetLastErrors(errs []Error) {
	w.lastErrors = errs
}

func (w *Watcher) watchLoop(chanResult chan Result) {
	httpClient, errRecorder := w.getClientAndDialErrRecorder()

	for w.active {
		r := w.watch(httpClient, errRecorder)
		if w.active {

			// send notifications
			w.smsNotify(r)
			w.mailNotify(r)
			w.slackNotify(r)

			chanResult <- *r
			time.Sleep(w.service.Interval)
		}
	}
}

func (w *Watcher) getClientAndDialErrRecorder() (client *http.Client, errRecorder *dialerErrRecorder) {
	errRecorder = &dialerErrRecorder{
		errors: []Error{},
	}
	tlsConfig := &tls.Config{}
	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 0 * time.Second,
	}
	dialTLS := func(network, address string) (conn net.Conn, err error) {
		tlsConn, tlsErr := tls.DialWithDialer(dialer, network, address, tlsConfig)
		if tlsErr == nil {
			//conn = tlsConn.(net.Conn)
			connectionState := tlsConn.ConnectionState()
			for _, cert := range connectionState.PeerCertificates {
				durationUntilExpiry := cert.NotAfter.Sub(time.Now())
				if durationUntilExpiry < w.service.TLSWarning {
					var (
						prefix = "cert CN=\"" + cert.Subject.CommonName
						certErr = Error{
							Error:   errors.New(
								fmt.Sprint(
									"cert CN=\"",
									cert.Subject.CommonName,
									"\" is expiring in less than "+strconv.FormatFloat(w.service.TLSWarning.Hours(), 'f', 0, 64)+"h: ",
									cert.NotAfter,
									", left: ",
									strconv.FormatFloat(durationUntilExpiry.Hours(), 'f', 0, 64),
									" hours",
								),
							).Error(),
							Type:    ErrorTypeCertificateIsExpiring,
							Comment: "",
						}
						updatedCertError bool
					)

					// iterate over previously recorded errors
					// and update the error for the currently affected service if a certificate expiry warning is already present
					for i, e := range errRecorder.errors {
						if e.Type == ErrorTypeCertificateIsExpiring {
							if strings.HasPrefix(e.Error, prefix) {
								updatedCertError = true
								errRecorder.errors[i] = certErr
								break
							}
						}
					}

					// if the error has not been updated it needs to be added initially
					if !updatedCertError {
						errRecorder.errors = append(
							errRecorder.errors,
							certErr,
						)
					}
				}
			}
			conn = tlsConn
		} else {

			switch reflect.TypeOf(tlsErr) {
			case typeX509UnknownAuthorityError:
				unknownAuthorityError := tlsErr.(x509.UnknownAuthorityError)
				errRecorder.tlsUnknownAuthorityError = &unknownAuthorityError
			case typeX509HostnameError:
				hostnameErr := tlsErr.(x509.HostnameError)
				errRecorder.tlsHostnameError = &hostnameErr
			case typeX509CertificateInvalidError:
				tlsCertificateInvalidError := tlsErr.(x509.CertificateInvalidError)
				errRecorder.tlsCertificateInvalidError = &tlsCertificateInvalidError
			case typeX509SystemRootsError:
				systemRootsError := tlsErr.(x509.SystemRootsError)
				errRecorder.tlsSystemRootsError = &systemRootsError
			default:
				log.Debug("unknown tls error", reflect.TypeOf(tlsErr), tlsErr)
			}
		}
		return conn, tlsErr
	}
	dial := func(network, address string) (conn net.Conn, err error) {
		conn, err = dialer.Dial(network, address)
		if err != nil {
			switch reflect.TypeOf(err) {
			case typeOpErr:
				opError := reflect.ValueOf(err).Elem().Interface().(net.OpError)
				switch reflect.TypeOf(opError.Err) {
				case typeDNSConfigErr:
					log.Error("dns config error")
					errRecorder.dnsConfigError = opError.Err.(net.Error)
				case typeDNSErr:
					log.Error("dns error")
					errRecorder.dnsError = opError.Err.(net.Error)
				default:
					errRecorder.unknownErr = opError.Err
				}
			default:
				log.Error("again some general bullshit", err)
				errRecorder.err = err.(net.Error)
			}
		}
		return
	}
	client = &http.Client{
		Transport: &http.Transport{
			Proxy:               http.ProxyFromEnvironment,
			Dial:                dial,
			DialTLS:             dialTLS,
			TLSHandshakeTimeout: 10 * time.Second,
			TLSClientConfig:     tlsConfig,
		},
		// do not follow redirects to allow checking the status code
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	return
}

// actual watch
func (w *Watcher) watch(client *http.Client, errRecorder *dialerErrRecorder) (r *Result) {

	r = NewResult(w.service.ID)

	// parsing, the endpoint
	request, err := http.NewRequest("GET", w.service.Endpoint, nil)
	if err != nil {
		r.addError(err, ErrorInvalidEndpoint, "")
		return r
	}
	// my personal dns error check
	if len(request.Host) > 0 {
		host := request.Host
		parts := strings.Split(host, ":")
		if len(parts) > 1 {
			host, _, err = net.SplitHostPort(request.Host)
			if err != nil {
				r.addError(err, ErrorInvalidEndpoint, "")
				return
			}
		}
		_, lookupErr := net.LookupIP(host)
		if lookupErr != nil {
			r.addError(lookupErr, ErrorTypeDNS, "")
			return
		}
	}

	// i am explicitly not calling http.Get, because it does 30x handling and i do not want that
	response, err := client.Do(request)
	r.Errors = append(r.Errors, errRecorder.errors...)

	if response != nil && response.Body != nil {
		// always close the body
		response.Body.Close()
	}

	if err != nil {
		// sth. went wrong
		r.addError(err, ErrorTypeClientError, "")
		var netErr net.Error
		switch true {
		case errRecorder.tlsHostnameError != nil:
			r.addError(errRecorder.tlsHostnameError, ErrorTypeTLSHostNameError, "")
		case errRecorder.tlsSystemRootsError != nil:
			r.addError(errRecorder.tlsSystemRootsError, ErrorTypeTLSSystemRootsError, "")
		case errRecorder.tlsUnknownAuthorityError != nil:
			r.addError(errRecorder.tlsUnknownAuthorityError, ErrorTypeTLSUnknownAuthority, "")
		case errRecorder.tlsCertificateInvalidError != nil:
			r.addError(errRecorder.tlsCertificateInvalidError, ErrorTypeTLSCertificateInvalid, "")
		case errRecorder.unknownErr != nil:
			r.addError(errRecorder.unknownErr, ErrorTypeUnknownError, "")
		case errRecorder.dnsConfigError != nil:
			netErr = errRecorder.dnsConfigError
			r.addError(errRecorder.dnsConfigError, ErrorTypeDNSConfig, "")
		case errRecorder.dnsError != nil:
			netErr = errRecorder.dnsError
			r.addError(errRecorder.dnsError, ErrorTypeDNS, "")
		case errRecorder.err != nil:
			netErr = errRecorder.err
			r.addError(errRecorder.err, ErrorTypeUnknownError, "")
		}
		if netErr != nil {
			r.Timeout = netErr.Timeout()
		}
		return
	}

	// prepare to run the session with cookies
	cookieJar, _ := cookiejar.New(nil)
	client.Jar = cookieJar
	errSession := w.runSession(r, client)
	if errSession != nil {
		log.Error("session error", errSession)
		r.addError(errSession, ErrorTypeSessionFail, "")
	}
	r.RunTime = time.Since(r.Timestamp)

	// r.addError(errors.New(fmt.Sprint("response time too slow:", runTimeMilliseconds, ", should not be more than:", maxRuntime)), ErrorTypeServerTooSlow)
	//r.StatusCode = response.StatusCode
	return
}

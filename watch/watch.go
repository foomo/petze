package watch

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"log"

	"reflect"

	"github.com/foomo/petze/config"
)

var typeDNSConfigErr = reflect.TypeOf(&net.DNSConfigError{})
var typeDNSErr = reflect.TypeOf(&net.DNSError{})
var typeOpErr = reflect.TypeOf(&net.OpError{})
var typeX509CertificateInvalidError = reflect.TypeOf(x509.CertificateInvalidError{})
var typeX509HostnameError = reflect.TypeOf(x509.HostnameError{})
var typeX509SystemRootsError = reflect.TypeOf(x509.SystemRootsError{})
var typeX509UnknownAuthorityError = reflect.TypeOf(x509.UnknownAuthorityError{})

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
	ErrorTypeGoQuerySyntax                	 = "goqueryMismatch"
	ErrorTypeDataMismatch                    = "dataMismatch"
)

type WarningType string

const (
	WarningTypeCertificateWillExpire WarningType = "certificateWillExpire"
)

type Error struct {
	Error   string    `json:"error"`
	Type    ErrorType `json:"type"`
	Comment string    `json:"comment,omitempty"`
}

type Warning struct {
	Warning string      `json:"warning"`
	Type    WarningType `json:"type"`
}

type Result struct {
	ID        string    `json:"id"`
	Errors    []Error   `json:"errors"`
	Warnings  []Warning `json:"warnings"`
	Timeout   bool      `json:"timeout"`
	Timestamp time.Time `json:"timestamp"`
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

func addWarning(warnings []Warning, warning string, t WarningType) []Warning {
	return append(warnings, Warning{
		Warning: warning,
		Type:    t,
	})
}

type dialerErrRecorder struct {
	warnings                   []Warning
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
	active  bool
	service *config.Service
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

func (w *Watcher) watchLoop(chanResult chan Result) {
	for w.active {
		r := watch(w.service)
		if w.active {
			chanResult <- *r
			time.Sleep(w.service.Interval)
		}
	}
}

func getClientAndDialErrRecorder() (client *http.Client, errRecorder *dialerErrRecorder) {
	errRecorder = &dialerErrRecorder{
		warnings: []Warning{},
		errors:   []Error{},
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
				if durationUntilExpiry < time.Hour*24 {
					// well in less than 24 h is worth an error
					errRecorder.errors = addError(errRecorder.errors, errors.New(fmt.Sprint("cert CN=\"", cert.Subject.CommonName, "\" is expiring in less than 24h: ", cert.NotAfter, ", left: ", durationUntilExpiry)), ErrorTypeCertificateIsExpiring, "")
				} else if durationUntilExpiry < time.Hour*24*14 {
					// we will start issuing warnings two weeks before a cert expires
					errRecorder.warnings = addWarning(errRecorder.warnings, fmt.Sprint("cert CN=\"", cert.Subject.CommonName, "\" expires on: ", cert.NotAfter, ", left: ", durationUntilExpiry), WarningTypeCertificateWillExpire)
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
				log.Println("unknown tls error", reflect.TypeOf(tlsErr), tlsErr)
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
					log.Println("dns config error")
					errRecorder.dnsConfigError = opError.Err.(net.Error)
				case typeDNSErr:
					log.Println("dns error")
					errRecorder.dnsError = opError.Err.(net.Error)
				default:
					errRecorder.unknownErr = opError.Err
				}
			default:
				log.Println("again some general bullshit", err)
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
	}
	return
}

// actual watch
func watch(service *config.Service) (r *Result) {
	r = NewResult(service.ID)
	// parsing, the endpoint
	request, err := http.NewRequest("GET", service.Endpoint, nil)
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
	client, errRecorder := getClientAndDialErrRecorder()
	response, err := client.Do(request)
	r.Errors = append(r.Errors, errRecorder.errors...)
	r.Warnings = append(r.Warnings, errRecorder.warnings...)
	//r.RunTime = time.Since(r.Timestamp)

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
		//log.Println("service", service.ID, err, "errRecorder dns:", errRecorder.dnsError, ", dnsConfig", errRecorder.dnsConfigError, ", err:", errRecorder.err, ", tls cert invalid:", errRecorder.tlsCertificateInvalidError)
		if netErr != nil {
			r.Timeout = netErr.Timeout()
			//r.ErrorIsTemporary = netErr.Temporary()
		}
		return
	}

	cookieJar, _ := cookiejar.New(nil)
	client.Jar = cookieJar
	errSession := runSession(service, r, client)
	if errSession != nil {
		log.Println("session error", errSession)
		r.addError(errSession, ErrorTypeSessionFail, "")
	}

	// r.addError(errors.New(fmt.Sprint("response time too slow:", runTimeMilliseconds, ", should not be more than:", maxRuntime)), ErrorTypeServerTooSlow)
	//r.StatusCode = response.StatusCode
	if response.StatusCode != http.StatusOK {
		r.addError(errors.New(fmt.Sprint("unexpected status code: ", response.StatusCode)), ErrorTypeWrongHTTPStatusCode, "")
	}
	return
}

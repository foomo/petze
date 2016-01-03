package watch

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"net/http"
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

type ErrorList struct {
	DNS                   bool `json:"dns"`
	DNSConfig             bool `json:"dnsConfig"`
	TLSCertificateInvalid bool `json:"tlsCertificateInvalid"`
	TLSHostNameError      bool `json:"tlsHostNameError"`
	TLSSystemRootsError   bool `json:"tlsSystemRootsError"`
	TLSUnknownAuthority   bool `json:"tlsUnknownAutority"`
}

type Result struct {
	ID      string    `json:"id"`
	Error   string    `json:"error"`
	Errors  ErrorList `json:"errInfo"`
	Timeout bool      `json:"timeout"`
	//ErrorIsTemporary bool          `json:"errorIsTemporary"`
	Timestamp  time.Time     `json:"timestamp"`
	RunTime    time.Duration `json:"runTime"`
	StatusCode int           `json:"statusCode"`
}

type dialerErrRecorder struct {
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
func Watch(service *config.Service, chanResult chan *Result) *Watcher {
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

func (w *Watcher) watchLoop(chanResult chan *Result) {
	for w.active {
		r := watch(w.service)
		if w.active {
			chanResult <- r
			time.Sleep(time.Second * time.Duration(w.service.Interval))
		}
	}
}

func getClientAndDialErrRecorder() (client *http.Client, errRecorder *dialerErrRecorder) {
	errRecorder = &dialerErrRecorder{}
	tlsConfig := &tls.Config{}
	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 0 * time.Second,
	}
	dialTLS := func(network, address string) (conn net.Conn, err error) {
		tlsConn, tlsErr := tls.DialWithDialer(dialer, network, address, tlsConfig)
		if tlsErr == nil {
			//conn = tlsConn.(net.Conn)
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
					log.Println("misc error", reflect.TypeOf(opError.Err), opError.Err)
					errRecorder.err = opError.Err.(net.Error)
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
	r = &Result{
		ID:        service.ID,
		Timestamp: time.Now(),
		Timeout:   false,
		Errors: ErrorList{
			DNS: false,
			TLSCertificateInvalid: false,
		},
	}
	// parsing, the endpoint
	request, err := http.NewRequest("GET", service.Endpoint, nil)
	if err != nil {
		r.Error = err.Error()
		return r
	}
	// my opersonal dns error check
	if len(request.Host) > 0 {
		host := request.Host
		parts := strings.Split(host, ":")
		if len(parts) > 1 {
			host, _, err = net.SplitHostPort(request.Host)
			if err != nil {
				r.Error = err.Error()
				return
			}
		}
		_, lookupErr := net.LookupIP(host)
		if lookupErr != nil {
			r.Error = lookupErr.Error()
			r.Errors.DNS = true
			return
		}
	}

	// i am explicitly not calling http.Get, because it does 30x handling, that
	// I do not want
	client, errRecorder := getClientAndDialErrRecorder()
	response, err := client.Do(request)
	r.RunTime = time.Since(r.Timestamp)
	if response != nil && response.Body != nil {
		// always close the body
		response.Body.Close()
	}
	if err != nil {
		r.Errors.TLSUnknownAuthority = errRecorder.tlsUnknownAuthorityError != nil
		r.Errors.TLSCertificateInvalid = errRecorder.tlsCertificateInvalidError != nil
		r.Errors.TLSHostNameError = errRecorder.tlsHostnameError != nil
		r.Errors.TLSSystemRootsError = errRecorder.tlsSystemRootsError != nil
		r.Errors.DNS = errRecorder.dnsError != nil
		r.Errors.DNSConfig = errRecorder.dnsConfigError != nil

		var netErr net.Error
		switch true {
		case r.Errors.DNSConfig:
			netErr = errRecorder.dnsConfigError
		case r.Errors.DNS:
			netErr = errRecorder.dnsError
		case errRecorder.err != nil:
			netErr = errRecorder.err

		}
		//log.Println("service", service.ID, err, "errRecorder dns:", errRecorder.dnsError, ", dnsConfig", errRecorder.dnsConfigError, ", err:", errRecorder.err, ", tls cert invalid:", errRecorder.tlsCertificateInvalidError)
		if netErr != nil {
			r.Timeout = netErr.Timeout()
			//r.ErrorIsTemporary = netErr.Temporary()
		}
		r.Error = err.Error()
		return
	}
	r.StatusCode = response.StatusCode
	if response.StatusCode != http.StatusOK {
		r.Error = fmt.Sprint("unexpected status code: ", response.StatusCode)
	}
	return
}

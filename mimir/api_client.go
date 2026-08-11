package mimir

// Largely copied from https://github.com/Mastercard/terraform-provider-restapi/blob/master/restapi/api_client.go

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"regexp"
	"strings"
	"time"
)

type apiClientOpt struct {
	uri             string
	rulerURI        string
	alertmanagerURI string
	distributorURI  string
	cert            string
	key             string
	ca              string
	token           string
	insecure        bool
	username        string
	password        string
	proxyURL        string
	headers         map[string]string
	timeout         int
	debug           bool
}

type apiClient struct {
	httpClient      *http.Client
	uri             string
	rulerURI        string
	alertmanagerURI string
	distributorURI  string
	insecure        bool
	token           string
	username        string
	password        string
	headers         map[string]string
	debug           bool
}

// Make a new api client for RESTful calls
func NewAPIClient(opt *apiClientOpt) (*apiClient, error) {
	/* Remove any trailing slashes since we will append
	   to this URL with our own root-prefixed location. Every per-component
	   override is concatenated with a root-prefixed path the same way, so they
	   all need the same treatment or the request URL grows a double slash. */
	opt.uri = strings.TrimSuffix(opt.uri, "/")
	opt.rulerURI = strings.TrimSuffix(opt.rulerURI, "/")
	opt.alertmanagerURI = strings.TrimSuffix(opt.alertmanagerURI, "/")
	opt.distributorURI = strings.TrimSuffix(opt.distributorURI, "/")

	// Setup HTTPS client
	tlsConfig := &tls.Config{}

	// Set insecure verify
	if opt.insecure {
		tlsConfig.InsecureSkipVerify = true
	}

	if opt.cert != "" && opt.key != "" {
		var cert tls.Certificate
		var err error
		if strings.HasPrefix(opt.cert, "-----BEGIN") && strings.HasPrefix(opt.key, "-----BEGIN") {
			cert, err = tls.X509KeyPair([]byte(opt.cert), []byte(opt.key))
		} else {
			cert, err = tls.LoadX509KeyPair(opt.cert, opt.key)
		}
		if err != nil {
			return nil, err
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	if opt.ca != "" {
		var caCert []byte
		var err error
		if strings.HasPrefix(opt.ca, "-----BEGIN") {
			caCert = []byte(opt.ca)
		} else {
			caCert, err = os.ReadFile(opt.ca)

			if err != nil {
				return nil, err
			}
		}
		caCertPool := x509.NewCertPool()
		if !caCertPool.AppendCertsFromPEM(caCert) {
			return nil, fmt.Errorf("no valid PEM certificate found in the provided ca")
		}
		tlsConfig.RootCAs = caCertPool
	}

	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
		Proxy:           http.ProxyFromEnvironment,
	}

	if opt.proxyURL != "" {
		// Gated on debug like every other log in this file: this ran on every
		// provider init with a proxy, wrote to stderr (which CI surfaces as an
		// error annotation), and the URL can carry credentials.
		if opt.debug {
			log.Printf("api_client.go: Using proxy: %s\n", opt.proxyURL)
		}
		proxy, err := url.Parse(opt.proxyURL)
		if err != nil {
			return nil, fmt.Errorf("error parsing proxy url: %s", err)
		}
		tr.Proxy = http.ProxyURL(proxy)
	}

	client := apiClient{
		httpClient: &http.Client{
			Timeout:   time.Second * time.Duration(opt.timeout),
			Transport: tr,
		},
		uri:             opt.uri,
		rulerURI:        opt.rulerURI,
		alertmanagerURI: opt.alertmanagerURI,
		distributorURI:  opt.distributorURI,
		insecure:        opt.insecure,
		token:           opt.token,
		username:        opt.username,
		password:        opt.password,
		headers:         opt.headers,
		debug:           opt.debug,
	}

	return &client, nil
}

/*
Helper function that handles sending/receiving and handling

	of HTTP data in and out.
*/
func (client *apiClient) sendRequest(component, method string, path, data string, headers map[string]string) (string, error) {
	var fullURI string

	switch {
	case component == "ruler" && client.rulerURI != "":
		fullURI = client.rulerURI + path
	case component == "alertmanager" && client.alertmanagerURI != "":
		fullURI = client.alertmanagerURI + path
	case component == "distributor" && client.distributorURI != "":
		fullURI = client.distributorURI + path
	default:
		fullURI = client.uri + path
	}

	var req *http.Request
	var err error

	buffer := bytes.NewBuffer([]byte(data))

	if data == "" {
		req, err = http.NewRequest(method, fullURI, nil)
	} else {
		req, err = http.NewRequest(method, fullURI, buffer)
	}

	if err != nil {
		return "", err
	}

	if client.token != "" {
		req.Header.Set("Authorization", "Bearer "+client.token)
	}

	// Set client headers from provider
	if len(client.headers) > 0 {
		for n, v := range client.headers {
			req.Header.Set(n, v)
		}
	}

	// Set client headers from resource
	if len(headers) > 0 {
		for n, v := range headers {
			req.Header.Set(n, v)
		}
	}

	if client.username != "" && client.password != "" {
		/* ... and fall back to basic auth if configured */
		req.SetBasicAuth(client.username, client.password)
	}

	if client.debug {
		reqDump, err := httputil.DumpRequestOut(req, false)
		if err != nil {
			return "", err
		}

		log.Printf("REQUEST:\n%s", redactCredentials(string(reqDump)))
	}

	resp, err := client.httpClient.Do(req)

	if err != nil {
		if client.debug {
			log.Printf("api_client.go: Error detected: %s\n", err)
		}
		return "", err
	}

	if client.debug {
		respDump, err := httputil.DumpResponse(resp, false)
		if err != nil {
			return "", err
		}

		log.Printf("RESPONSE:\n%s", redactCredentials(string(respDump)))
	}

	bodyBytes, err2 := io.ReadAll(resp.Body)
	resp.Body.Close()

	if err2 != nil {
		return "", err2
	}
	body := string(bodyBytes)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return body, fmt.Errorf("unexpected response code '%d': %s", resp.StatusCode, body)
	}

	return body, nil
}

// credentialHeaderRegexp matches the header lines that carry a credential in an
// httputil dump. The dumps deliberately exclude bodies: an alertmanager
// configuration body carries slack api_url, pushover token, smtp_auth_password
// and the like, so dumping it would leak exactly what this redaction exists to
// protect. Headers still need masking for the bearer token and basic-auth pair.
var credentialHeaderRegexp = regexp.MustCompile(`(?im)^(Authorization|Proxy-Authorization|X-Amz-Security-Token):.*$`)

func redactCredentials(dump string) string {
	return credentialHeaderRegexp.ReplaceAllString(dump, "$1: <redacted>")
}

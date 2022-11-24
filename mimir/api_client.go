package mimir

// Largely copied from https://github.com/Mastercard/terraform-provider-restapi/blob/master/restapi/api_client.go

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"strings"
	"time"
)

type apiClientOpt struct {
	uri              string
	ruler_uri        string
	alertmanager_uri string
	cert             string
	key              string
	ca               string
	token            string
	insecure         bool
	username         string
	password         string
	headers          map[string]string
	timeout          int
	debug            bool
}

type api_client struct {
	http_client      *http.Client
	uri              string
	ruler_uri        string
	alertmanager_uri string
	cert             string
	key              string
	ca               string
	insecure         bool
	token            string
	username         string
	password         string
	headers          map[string]string
	timeout          int
	debug            bool
}

// Make a new api client for RESTful calls
func NewAPIClient(opt *apiClientOpt) (*api_client, error) {

	if opt.uri == "" && opt.ruler_uri == "" && opt.alertmanager_uri == "" {
		return nil, errors.New("No provider URIs defined. Please set uri, or ruler_rui/alertmanager_uri.")
	}

	/* Remove any trailing slashes since we will append
	   to this URL with our own root-prefixed location */
	if strings.HasSuffix(opt.uri, "/") {
		opt.uri = opt.uri[:len(opt.uri)-1]
	}

	// Setup HTTPS client
	tlsConfig := &tls.Config{
		InsecureSkipVerify: opt.insecure,
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
		tlsConfig.BuildNameToCertificate()
	}

	if opt.ca != "" {
		var caCert []byte
		var err error
		if strings.HasPrefix(opt.ca, "-----BEGIN") {
			caCert = []byte(opt.ca)
		} else {
			caCert, err = ioutil.ReadFile(opt.ca)

			if err != nil {
				return nil, err
			}
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
		tlsConfig.BuildNameToCertificate()
	}

	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	client := api_client{
		http_client: &http.Client{
			Timeout:   time.Second * time.Duration(opt.timeout),
			Transport: tr,
		},
		uri:              opt.uri,
		ruler_uri:        opt.ruler_uri,
		alertmanager_uri: opt.alertmanager_uri,
		insecure:         opt.insecure,
		token:            opt.token,
		username:         opt.username,
		password:         opt.password,
		headers:          opt.headers,
		debug:            opt.debug,
	}

	return &client, nil
}

/* Helper function that handles sending/receiving and handling
   of HTTP data in and out. */
func (client *api_client) send_request(component, method string, path, data string, headers map[string]string) (string, error) {
	var full_uri string

	if component == "ruler" && client.ruler_uri != "" {
		full_uri = client.ruler_uri + path
	} else if component == "alertmanager" && client.ruler_uri != "" {
		full_uri = client.alertmanager_uri + path
	} else {
		full_uri = client.uri + path
	}

	var req *http.Request
	var err error

	buffer := bytes.NewBuffer([]byte(data))

	if data == "" {
		req, err = http.NewRequest(method, full_uri, nil)
	} else {
		req, err = http.NewRequest(method, full_uri, buffer)
	}

	if err != nil {
		log.Fatal(err)
	}

	if client.token != "" {
		req.Header.Set("Authorization", "Bearer " + client.token)
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

		reqDump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("REQUEST:\n%s", string(reqDump))
	}

	resp, err := client.http_client.Do(req)

	if err != nil {
		log.Printf("api_client.go: Error detected: %s\n", err)
		return "", err
	}

	if client.debug {

		respDump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("RESPONSE:\n%s", string(respDump))
	}

	bodyBytes, err2 := ioutil.ReadAll(resp.Body)
	resp.Body.Close()

	if err2 != nil {
		return "", err2
	}
	body := string(bodyBytes)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return body, errors.New(fmt.Sprintf("Unexpected response code '%d': %s", resp.StatusCode, body))
	}

	return body, nil

}

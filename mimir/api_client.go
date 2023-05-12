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
	"os"
	"strings"
	"time"
)

type apiClientOpt struct {
	uri             string
	rulerURI        string
	alertmanagerURI string
	cert            string
	key             string
	ca              string
	token           string
	insecure        bool
	username        string
	password        string
	headers         map[string]string
	timeout         int
	debug           bool
}

type apiClient struct {
	httpClient      *http.Client
	uri             string
	rulerURI        string
	alertmanagerURI string
	insecure        bool
	token           string
	username        string
	password        string
	headers         map[string]string
	debug           bool
}

// Make a new api client for RESTful calls
func NewAPIClient(opt *apiClientOpt) (*apiClient, error) {
	if opt.uri == "" && opt.rulerURI == "" && opt.alertmanagerURI == "" {
		return nil, fmt.Errorf("no provider URIs defined. Please set uri, or ruler_uui/alertmanager_uri")
	}

	/* Remove any trailing slashes since we will append
	   to this URL with our own root-prefixed location */
	opt.uri = strings.TrimSuffix(opt.uri, "/")

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
		caCertPool.AppendCertsFromPEM(caCert)
		tlsConfig.RootCAs = caCertPool
	}

	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	client := apiClient{
		httpClient: &http.Client{
			Timeout:   time.Second * time.Duration(opt.timeout),
			Transport: tr,
		},
		uri:             opt.uri,
		rulerURI:        opt.rulerURI,
		alertmanagerURI: opt.alertmanagerURI,
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
		log.Fatal(err)
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
		reqDump, err := httputil.DumpRequestOut(req, true)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("REQUEST:\n%s", string(reqDump))
	}

	resp, err := client.httpClient.Do(req)

	if err != nil {
		if client.debug {
			log.Printf("api_client.go: Error detected: %s\n", err)
		}
		return "", err
	}

	if client.debug {
		respDump, err := httputil.DumpResponse(resp, true)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("RESPONSE:\n%s", string(respDump))
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

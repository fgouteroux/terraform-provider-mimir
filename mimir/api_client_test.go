package mimir

// Largely copied from https://github.com/Mastercard/terraform-provider-restapi/blob/master/restapi/api_client_test.go

import (
	"fmt"
	"log"
	"net/http"
	"testing"
	"time"
)

var apiClientServer *http.Server

func TestAPIClient(t *testing.T) {
	debug := false
	address := "127.0.0.1:8082"

	if debug {
		log.Println("api_client_test.go: Starting HTTP server")
	}
	setupAPIClientServer(debug, address)
	defer shutdownAPIClientServer()

	/* Notice the intentional trailing / */
	opt := &apiClientOpt{
		uri:     "http://127.0.0.1:8082/",
		headers: make(map[string]string, 0),
		timeout: 2,
		debug:   debug,
	}
	client, err := NewAPIClient(opt)
	if err != nil {
		t.Fatalf("api_client_test.go: Failed to init api client, err: %v", err)
	}

	var res string

	if debug {
		log.Printf("api_client_test.go: Testing standard OK request\n")
	}
	var headers map[string]string
	res, err = client.sendRequest("", "GET", "/ok", "", headers)
	if err != nil {
		t.Fatalf("api_client_test.go: %s", err)
	}
	if res != "It works!" {
		t.Fatalf("api_client_test.go: Got back '%s' but expected 'It works!'\n", res)
	}

	if debug {
		log.Printf("api_client_test.go: Testing redirect request\n")
	}
	res, err = client.sendRequest("", "GET", "/redirect", "", headers)
	if err != nil {
		t.Fatalf("api_client_test.go: %s", err)
	}
	if res != "It works!" {
		t.Fatalf("api_client_test.go: Got back '%s' but expected 'It works!'\n", res)
	}

	/* Verify timeout works */
	if debug {
		log.Printf("api_client_test.go: Testing timeout aborts requests\n")
	}
	_, err = client.sendRequest("", "GET", "/slow", "", headers)
	if err != nil {
		if debug {
			log.Println("api_client_test.go: slow request expected")
		}
	} else {
		t.Fatalf("api_client_test.go: Timeout did not trigger on slow request")
	}

	if debug {
		log.Println("api_client_test.go: Stopping HTTP server")
	}
}

func TestAPIClientTLSUnsecure(t *testing.T) {
	debug := false
	address := "127.0.0.1:8083"
	setupAPIClientServer(debug, address)
	defer shutdownAPIClientServer()

	/* Notice the intentional trailing / */
	opt := &apiClientOpt{
		uri:      fmt.Sprintf("https://%s/", address),
		insecure: true,
		cert:     "../tests/certs/server.crt",
		key:      "../tests/certs/server.key",
		ca:       "../tests/certs/ca.crt",
		headers:  make(map[string]string, 0),
		timeout:  2,
		debug:    debug,
	}
	_, err := NewAPIClient(opt)
	if err != nil {
		t.Fatalf("api_client_test.go: Failed to init api client, err: %v", err)
	}
}

func TestAPIClientTLS(t *testing.T) {
	debug := false
	address := "127.0.0.1:8084"
	setupAPIClientServer(debug, address)
	defer shutdownAPIClientServer()

	/* Notice the intentional trailing / */
	opt := &apiClientOpt{
		uri:     fmt.Sprintf("https://%s/", address),
		cert:    "../tests/certs/server.crt",
		key:     "../tests/certs/server.key",
		ca:      "../tests/certs/ca.crt",
		headers: make(map[string]string, 0),
		timeout: 2,
		debug:   debug,
	}
	_, err := NewAPIClient(opt)
	if err != nil {
		t.Fatalf("api_client_test.go: Failed to init api client, err: %v", err)
	}
}

func TestAPIClientProxy(t *testing.T) {
	debug := false
	address := "127.0.0.1:8085"
	setupAPIClientServer(debug, address)
	defer shutdownAPIClientServer()

	/* Notice the intentional trailing / */
	opt := &apiClientOpt{
		uri:      fmt.Sprintf("http://%s/", address),
		insecure: false,
		headers:  make(map[string]string, 0),
		timeout:  2,
		debug:    debug,
		proxyURL: "http://localhost:3128",
	}
	_, err := NewAPIClient(opt)
	if err != nil {
		t.Fatalf("api_client_test.go: Failed to init api client, err: %v", err)
	}
}

func TestAPIClientBasicAuth(t *testing.T) {
	debug := false
	address := "127.0.0.1:8086"
	setupAPIClientServer(debug, address)
	defer shutdownAPIClientServer()

	/* Notice the intentional trailing / */
	opt := &apiClientOpt{
		uri:      fmt.Sprintf("http://%s/", address),
		insecure: false,
		username: "loki",
		password: "password",
		headers:  make(map[string]string, 0),
		timeout:  2,
		debug:    debug,
	}
	client, err := NewAPIClient(opt)
	if err != nil {
		t.Fatalf("api_client_test.go: Failed to init api client, err: %v", err)
	}
	var headers map[string]string
	_, err = client.sendRequest("", "GET", "/ok", "", headers)
	if err != nil {
		t.Fatalf("api_client_test.go: %s", err)
	}
}

func TestAPIClientBearerAuth(t *testing.T) {
	debug := false
	address := "127.0.0.1:8087"
	setupAPIClientServer(debug, address)
	defer shutdownAPIClientServer()

	/* Notice the intentional trailing / */
	opt := &apiClientOpt{
		uri:     fmt.Sprintf("http://%s/", address),
		token:   "supersecret",
		headers: make(map[string]string, 0),
		timeout: 2,
		debug:   debug,
	}
	client, err := NewAPIClient(opt)
	if err != nil {
		t.Fatalf("api_client_test.go: Failed to init api client, err: %v", err)
	}
	var headers map[string]string
	_, err = client.sendRequest("", "GET", "/ok", "", headers)
	if err != nil {
		t.Fatalf("api_client_test.go: %s", err)
	}
}

func TestAPIClientDebug(t *testing.T) {
	debug := true
	address := "127.0.0.1:8088"
	setupAPIClientServer(debug, address)
	defer shutdownAPIClientServer()

	/* Notice the intentional trailing / */
	opt := &apiClientOpt{
		uri:     fmt.Sprintf("http://%s/", address),
		headers: make(map[string]string, 0),
		timeout: 2,
		debug:   debug,
	}
	client, err := NewAPIClient(opt)
	if err != nil {
		t.Fatalf("api_client_test.go: Failed to init api client, err: %v", err)
	}
	var headers map[string]string
	_, err = client.sendRequest("", "GET", "/ok", "", headers)
	if err != nil {
		t.Fatalf("api_client_test.go: %s", err)
	}
}

func TestAPIClientHeaders(t *testing.T) {
	debug := false
	address := "127.0.0.1:8089"
	setupAPIClientServer(debug, address)
	defer shutdownAPIClientServer()

	headers := map[string]string{
		"Custom": "header set",
	}
	/* Notice the intentional trailing / */
	opt := &apiClientOpt{
		uri:     fmt.Sprintf("http://%s/", address),
		headers: headers,
		timeout: 2,
		debug:   debug,
	}
	client, err := NewAPIClient(opt)
	if err != nil {
		t.Fatalf("api_client_test.go: Failed to init api client, err: %v", err)
	}
	_, err = client.sendRequest("", "GET", "/ok", "", headers)
	if err != nil {
		t.Fatalf("api_client_test.go: %s", err)
	}
}

func setupAPIClientServer(debug bool, address string) {
	serverMux := http.NewServeMux()
	serverMux.HandleFunc("/ok", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("It works!"))
	})
	serverMux.HandleFunc("/slow", func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(9999 * time.Second)
		_, _ = w.Write([]byte("This will never return!!!!!"))
	})
	serverMux.HandleFunc("/redirect", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/ok", http.StatusPermanentRedirect)
	})

	apiClientServer = &http.Server{
		Addr:              address,
		Handler:           serverMux,
		ReadTimeout:       1 * time.Second,
		WriteTimeout:      1 * time.Second,
		IdleTimeout:       30 * time.Second,
		ReadHeaderTimeout: 2 * time.Second,
	}
	go func() {
		err := apiClientServer.ListenAndServe()
		if err != nil && debug {
			log.Println(err)
		}
	}()
	/* let the server start */
	time.Sleep(1 * time.Second)
}

func shutdownAPIClientServer() {
	apiClientServer.Close()
}

package mimir

// Largely copied from https://github.com/Mastercard/terraform-provider-restapi/blob/master/restapi/api_client_test.go

import (
	"log"
	"net/http"
	"testing"
	"time"
)

var apiClientServer *http.Server

func TestAPIClient(t *testing.T) {
	debug := false

	if debug {
		log.Println("client_test.go: Starting HTTP server")
	}
	setupAPIClientServer(debug)
	defer shutdownAPIClientServer()

	/* Notice the intentional trailing / */
	opt := &apiClientOpt{
		uri:             "http://127.0.0.1:8082/",
		rulerURI:        "http://127.0.0.1:8082/",
		alertmanagerURI: "http://127.0.0.1:8082/",
		insecure:        false,
		username:        "",
		password:        "",
		proxyURL:        "",
		token:           "",
		cert:            "",
		key:             "",
		ca:              "",
		headers:         make(map[string]string, 0),
		timeout:         2,
		debug:           debug,
	}
	client, err := NewAPIClient(opt)
	if err != nil {
		t.Fatalf("client_test.go: %s", err)
	}

	var res string

	if debug {
		log.Printf("api_client_test.go: Testing standard OK request\n")
	}
	var headers map[string]string
	res, err = client.sendRequest("", "GET", "/ok", "", headers)
	if err != nil {
		t.Fatalf("client_test.go: %s", err)
	}
	if res != "It works!" {
		t.Fatalf("client_test.go: Got back '%s' but expected 'It works!'\n", res)
	}

	if debug {
		log.Printf("api_client_test.go: Testing redirect request\n")
	}
	res, err = client.sendRequest("", "GET", "/redirect", "", headers)
	if err != nil {
		t.Fatalf("client_test.go: %s", err)
	}
	if res != "It works!" {
		t.Fatalf("client_test.go: Got back '%s' but expected 'It works!'\n", res)
	}

	/* Verify timeout works */
	if debug {
		log.Printf("api_client_test.go: Testing timeout aborts requests\n")
	}
	_, err = client.sendRequest("", "GET", "/slow", "", headers)
	if err != nil {
		if debug {
			log.Println("client_test.go: slow request expected")
		}
	} else {
		t.Fatalf("client_test.go: Timeout did not trigger on slow request")
	}

	if debug {
		log.Println("client_test.go: Stopping HTTP server")
	}
}

func setupAPIClientServer(debug bool) {
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
		Addr:              "127.0.0.1:8082",
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

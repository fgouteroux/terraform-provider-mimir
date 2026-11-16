package mimir

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// A 404 from Mimir's router (wrong URI, missing /prometheus prefix) must not be
// mistaken for "object deleted", or reads silently drop the resource from state.
func TestSendRequestRouteNotFound(t *testing.T) {
	path := rulesGroupPath("ns", "missing")
	mux := http.NewServeMux()
	mux.HandleFunc("/prometheus"+path, func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "group does not exist", http.StatusNotFound)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	newClient := func(opt *apiClientOpt) *apiClient {
		opt.headers = map[string]string{}
		opt.timeout = 2
		client, err := NewAPIClient(opt)
		if err != nil {
			t.Fatalf("NewAPIClient: %v", err)
		}
		return client
	}

	// ruler_uri carries the /prometheus prefix: Mimir answers "group does not exist".
	withPrefix := newClient(&apiClientOpt{uri: srv.URL, rulerURI: srv.URL + "/prometheus"})
	_, err := withPrefix.sendRequest("ruler", "GET", path, "", nil)
	if err == nil || !strings.Contains(err.Error(), "response code '404'") {
		t.Errorf("object not found: want a \"response code '404'\" error, got %v", err)
	}

	// Only uri, without the prefix: the router answers "404 page not found".
	withoutPrefix := newClient(&apiClientOpt{uri: srv.URL})
	_, err = withoutPrefix.sendRequest("ruler", "GET", path, "", nil)
	if err == nil {
		t.Fatal("route not found: want an error, got nil")
	}
	if strings.Contains(err.Error(), "response code '404'") {
		t.Errorf("route not found must not read as object not found, got %v", err)
	}
}

func TestDistributorStatsPath(t *testing.T) {
	for uri, want := range map[string]string{
		"http://mimir:8080/distributor": "/all_user_stats",
		"http://mimir:8080":             "/distributor/all_user_stats",
		"":                              "/distributor/all_user_stats",
	} {
		if got := distributorStatsPath(uri); got != want {
			t.Errorf("distributorStatsPath(%q) = %q, want %q", uri, got, want)
		}
	}
}

// Mimir's ruler accepts experimental functions, binop fill modifiers and the
// functions of its Prometheus fork, so the provider must not reject them.
func TestValidatePromQLExprMimirFeatures(t *testing.T) {
	for _, expr := range []string{
		"double_exponential_smoothing(rate(http_requests_total[5m])[1h:1m], 0.5, 0.5)",
		"a + on(job) fill(0) b",
		"max_of(scalar(a), 1)",
		"min_of(scalar(a), 1)",
		"step()",
	} {
		if _, errs := validatePromQLExpr(expr, "expr"); len(errs) > 0 {
			t.Errorf("%q: unexpected errors %v", expr, errs)
		}
	}

	if _, errs := validatePromQLExpr("rate(x[5m]", "expr"); len(errs) == 0 {
		t.Error("invalid expression accepted")
	}
}

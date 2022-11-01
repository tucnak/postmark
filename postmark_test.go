package postmark

import (
	"errors"
	"goji.io/pat"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"goji.io"
)

var (
	tMux    = goji.NewMux()
	tServer *httptest.Server
	client  *Client
)

func TestErrorHandling(t *testing.T) {
	unprocessable := `{
	  "ErrorCode": 402,
	  "Message": "Invalid JSON"
	}`

	// test 422
	tMux.HandleFunc(pat.Post("/unprocessable"), func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(unprocessable))
	})

	err := client.doRequest(parameters{
		Method: http.MethodPost,
		Path:   "unprocessable",
	}, nil)

	if err == nil {
		t.Errorf("expected error, got nil")
	}

	apiError := APIError{}
	if !errors.As(err, &apiError) {
		t.Errorf("expected APIError, got %T", err)
	}

	if apiError.ErrorCode != 402 {
		t.Errorf("expected internal error code 402, got %d", apiError.ErrorCode)
	}

	if apiError.Message != "Invalid JSON" {
		t.Errorf("expected 'Invalid JSON' message got %v", apiError.Message)
	}

	// test 404
	tMux.HandleFunc(pat.Post("/not-found"), func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte("not found"))
	})

	err = client.doRequest(parameters{
		Method: http.MethodPost,
		Path:   "not-found",
	}, nil)

	if err == nil {
		t.Errorf("expected error, got nil")
	}

	httpError := HttpError{}
	if !errors.As(err, &httpError) {
		t.Errorf("expected HttpError, got %T", err)
	}

	if httpError.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", httpError.StatusCode)
	}
}

func init() {
	tServer = httptest.NewServer(tMux)

	transport := &http.Transport{
		Proxy: func(req *http.Request) (*url.URL, error) {
			// Reroute...
			return url.Parse(tServer.URL)
		},
	}

	client = NewClient("", "")
	client.HTTPClient = &http.Client{Transport: transport}
	client.BaseURL = tServer.URL
}

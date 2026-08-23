package main

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/tarmac-project/sdk/httpclient"
)

type fakeHTTPClient struct {
	getURL string
	rsp    *httpclient.Response
	err    error
}

func (f *fakeHTTPClient) Get(rawURL string) (*httpclient.Response, error) {
	f.getURL = rawURL
	return f.rsp, f.err
}

func (*fakeHTTPClient) Post(string, string, io.Reader) (*httpclient.Response, error) {
	return nil, errors.New("unexpected Post call")
}

func (*fakeHTTPClient) Put(string, string, io.Reader) (*httpclient.Response, error) {
	return nil, errors.New("unexpected Put call")
}

func (*fakeHTTPClient) Delete(string) (*httpclient.Response, error) {
	return nil, errors.New("unexpected Delete call")
}

func (*fakeHTTPClient) Do(*httpclient.Request) (*httpclient.Response, error) {
	return nil, errors.New("unexpected Do call")
}

type fakeLogger struct {
	warnings []string
}

func (*fakeLogger) Info(string)  {}
func (*fakeLogger) Error(string) {}
func (*fakeLogger) Debug(string) {}
func (*fakeLogger) Trace(string) {}
func (f *fakeLogger) Warn(message string) {
	f.warnings = append(f.warnings, message)
}

func TestResolveAirportsCSVURL(t *testing.T) {
	tests := []struct {
		name    string
		payload string
		want    string
		wantErr bool
	}{
		{name: "default", want: airportsCSVURL},
		{name: "trimmed HTTP", payload: "  http://fixture/airports.csv\n", want: "http://fixture/airports.csv"},
		{name: "HTTPS", payload: "https://example.com/data.csv", want: "https://example.com/data.csv"},
		{name: "relative", payload: "/airports.csv", wantErr: true},
		{name: "unsupported scheme", payload: "file:///tmp/airports.csv", wantErr: true},
		{name: "missing host", payload: "https:///airports.csv", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := resolveAirportsCSVURL([]byte(tc.payload))
			if (err != nil) != tc.wantErr {
				t.Fatalf("resolveAirportsCSVURL() error = %v, wantErr %v", err, tc.wantErr)
			}
			if got != tc.want {
				t.Fatalf("resolveAirportsCSVURL() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestHandler(t *testing.T) {
	tests := []struct {
		name       string
		payload    string
		rsp        *httpclient.Response
		getErr     error
		want       string
		wantURL    string
		wantErrSub string
	}{
		{
			name:    "downloads default URL",
			rsp:     response(http.StatusOK, "csv-data"),
			want:    "csv-data",
			wantURL: airportsCSVURL,
		},
		{
			name:    "downloads supplied URL",
			payload: " http://fixture/airports.csv ",
			rsp:     response(http.StatusOK, "fixture-data"),
			want:    "fixture-data",
			wantURL: "http://fixture/airports.csv",
		},
		{
			name:       "rejects invalid URL before HTTP call",
			payload:    "ftp://fixture/airports.csv",
			wantErrSub: "absolute HTTP or HTTPS",
		},
		{
			name:       "rejects nil response",
			wantURL:    airportsCSVURL,
			wantErrSub: "empty response",
		},
		{
			name:       "returns HTTP client error",
			getErr:     errors.New("network unavailable"),
			wantURL:    airportsCSVURL,
			wantErrSub: "network unavailable",
		},
		{
			name:       "rejects non-success status",
			rsp:        response(http.StatusBadGateway, "bad gateway"),
			wantURL:    airportsCSVURL,
			wantErrSub: "returned 502",
		},
		{
			name:       "rejects nil body",
			rsp:        &httpclient.Response{StatusCode: http.StatusOK},
			wantURL:    airportsCSVURL,
			wantErrSub: "empty response body",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			httpClient := &fakeHTTPClient{rsp: tc.rsp, err: tc.getErr}
			function := &Function{logging: &fakeLogger{}, http: httpClient}

			got, err := function.Handler([]byte(tc.payload))
			if tc.wantErrSub == "" && err != nil {
				t.Fatalf("Handler() unexpected error: %v", err)
			}
			if tc.wantErrSub != "" && (err == nil || !strings.Contains(err.Error(), tc.wantErrSub)) {
				t.Fatalf("Handler() error = %v, want substring %q", err, tc.wantErrSub)
			}
			if string(got) != tc.want {
				t.Fatalf("Handler() = %q, want %q", got, tc.want)
			}
			if httpClient.getURL != tc.wantURL {
				t.Fatalf("HTTP URL = %q, want %q", httpClient.getURL, tc.wantURL)
			}
		})
	}
}

func response(status int, body string) *httpclient.Response {
	return &httpclient.Response{
		StatusCode: status,
		Body:       io.NopCloser(strings.NewReader(body)),
	}
}

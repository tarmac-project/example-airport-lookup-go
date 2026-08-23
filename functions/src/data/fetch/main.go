/*
This function fetches airport data from a remote CSV file via HTTP and returns it to the application. It is designed
for reuse throughout the application.
*/
package main

import (
	"fmt"
	"io"
	"net/url"
	"strings"

	sdk "github.com/tarmac-project/sdk"
	"github.com/tarmac-project/sdk/httpclient"
	"github.com/tarmac-project/sdk/logging"
)

const airportsCSVURL = "https://raw.githubusercontent.com/davidmegginson/ourairports-data/main/airports.csv"

// Function implements the airport CSV fetch behavior.
type Function struct {
	sdk     *sdk.SDK
	logging logging.Client
	http    httpclient.Client
}

// Handler downloads airport CSV data and returns the raw content.
func (f *Function) Handler(payload []byte) ([]byte, error) {
	csvURL, err := resolveAirportsCSVURL(payload)
	if err != nil {
		return nil, err
	}

	f.logging.Info("downloading airports.csv")

	rsp, err := f.fetchAirportCSV(csvURL)
	if err != nil {
		return nil, err
	}

	return rsp, nil
}

func resolveAirportsCSVURL(payload []byte) (string, error) {
	rawURL := strings.TrimSpace(string(payload))
	if rawURL == "" {
		return airportsCSVURL, nil
	}

	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil || parsed.Host == "" || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", fmt.Errorf("airport CSV URL must be an absolute HTTP or HTTPS URL")
	}

	return parsed.String(), nil
}

func (f *Function) fetchAirportCSV(csvURL string) ([]byte, error) {
	rsp, err := f.http.Get(csvURL)
	if err != nil {
		f.logging.Error(fmt.Sprintf("failed to get airports.csv: %v", err))
		return nil, fmt.Errorf("failed to get airports.csv: %w", err)
	}
	if rsp == nil {
		f.logging.Error("airports.csv download returned empty response")
		return nil, fmt.Errorf("failed to get airports.csv: empty response")
	}
	if rsp.Body != nil {
		defer func() {
			if closeErr := rsp.Body.Close(); closeErr != nil {
				f.logging.Warn(fmt.Sprintf("failed to close airports.csv response body: %v", closeErr))
			}
		}()
	}

	f.logging.Info(fmt.Sprintf("airports.csv downloaded with return code: %d", rsp.StatusCode))

	if rsp.StatusCode < 200 || rsp.StatusCode >= 300 {
		f.logging.Error(
			fmt.Sprintf("airports.csv download failed with return code: %d", rsp.StatusCode),
		)
		return nil, fmt.Errorf(
			"failed to get airports.csv: http request returned %d",
			rsp.StatusCode,
		)
	}

	if rsp.Body == nil {
		f.logging.Error("airports.csv download returned empty response body")
		return nil, fmt.Errorf("failed to get airports.csv: empty response body")
	}

	body, err := io.ReadAll(rsp.Body)
	if err != nil {
		f.logging.Error(fmt.Sprintf("failed to read airports.csv body: %v", err))
		return nil, fmt.Errorf("failed to read airports.csv body: %w", err)
	}

	return body, nil
}

// Initialize sets up SDK and HTTP clients required by the fetch function.
//
//go:wasmexport wapc_init
func Initialize() {
	var err error

	// Initialize Function
	f := &Function{}

	// Initialize the Tarmac SDK
	f.sdk, err = sdk.New(sdk.Config{
		Namespace: "tarmac",
		Handler:   f.Handler,
	})
	if err != nil {
		return
	}

	cfg := f.sdk.Config()

	// Initialize Logger client
	f.logging, err = logging.New(logging.Config{
		SDKConfig: cfg,
	})
	if err != nil {
		return
	}

	// Initialize HTTP client
	f.http, err = httpclient.New(httpclient.Config{
		SDKConfig: cfg,
	})
	if err != nil {
		f.logging.Error(fmt.Sprintf("failed to create HTTP client: %v", err))
		return
	}
}

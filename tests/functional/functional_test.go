package functional_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"
)

type lookupResponse struct {
	OK      bool   `json:"ok"`
	Source  string `json:"source"`
	Stage   string `json:"stage"`
	Airport struct {
		LocalCode string `json:"local_code"`
	} `json:"airport"`
}

type loadResponse struct {
	ParsedAirports   int `json:"parsed_airports"`
	SuccessfulUpsert int `json:"successful_upsert"`
	FailedUpsert     int `json:"failed_upsert"`
}

type lookupCase struct {
	name       string
	code       string
	wantOK     bool
	wantSource string
	wantStage  string
}

func TestAirportWorkflow(t *testing.T) {
	if os.Getenv("AIRPORT_FUNCTIONAL") != "1" {
		t.Skip("set AIRPORT_FUNCTIONAL=1 to run functional tests")
	}

	baseURL := envOrDefault("BASE_URL", "http://localhost:8080")
	artifactDir := envOrDefault("ARTIFACT_DIR", "artifacts/functional")
	if err := os.MkdirAll(artifactDir, 0o755); err != nil {
		t.Fatalf("create artifact directory: %v", err)
	}

	client := &http.Client{Timeout: 10 * time.Second}
	waitForRuntime(t, client, baseURL)

	lookupCases := []lookupCase{
		{name: "seed_sql", code: "PHX", wantOK: true, wantSource: "sql"},
		{name: "seed_cache", code: "PHX", wantOK: true, wantSource: "cache"},
	}
	for _, tc := range lookupCases {
		response := postLookup(t, client, baseURL, artifactDir, tc.name, tc.code)
		assertLookup(t, response, tc.code, tc.wantOK, tc.wantSource, tc.wantStage)
	}

	loadBody := post(t, client, baseURL+"/load", "text/plain", []byte("http://fixture-server/airports.csv"))
	writeArtifact(t, artifactDir, "load", loadBody)
	var loaded loadResponse
	if err := json.Unmarshal(loadBody, &loaded); err != nil {
		t.Fatalf("decode load response: %v", err)
	}
	if loaded.ParsedAirports != 1 || loaded.SuccessfulUpsert != 1 || loaded.FailedUpsert != 0 {
		t.Fatalf("load response = %+v, want one successful airport", loaded)
	}

	lookupCases = []lookupCase{
		{name: "fixture_sql", code: "CIX1", wantOK: true, wantSource: "sql"},
		{name: "fixture_cache", code: "CIX1", wantOK: true, wantSource: "cache"},
		{name: "invalid", code: "A-1", wantStage: "validation"},
		{name: "unknown", code: "ZZZZ", wantStage: "sql_query"},
	}
	for _, tc := range lookupCases {
		response := postLookup(t, client, baseURL, artifactDir, tc.name, tc.code)
		assertLookup(t, response, tc.code, tc.wantOK, tc.wantSource, tc.wantStage)
	}
}

func waitForRuntime(t *testing.T, client *http.Client, baseURL string) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		request, err := http.NewRequestWithContext(
			ctx,
			http.MethodPost,
			baseURL+"/",
			bytes.NewBufferString(`{"local_code":"A-1"}`),
		)
		if err != nil {
			t.Fatalf("create readiness request: %v", err)
		}
		request.Header.Set("Content-Type", "application/json")

		response, err := client.Do(request)
		if err == nil {
			_, _ = io.Copy(io.Discard, response.Body)
			_ = response.Body.Close()
			if response.StatusCode >= http.StatusOK && response.StatusCode < http.StatusMultipleChoices {
				return
			}
		}

		select {
		case <-ctx.Done():
			t.Fatalf("Tarmac did not become ready: %v", ctx.Err())
		case <-ticker.C:
		}
	}
}

func postLookup(
	t *testing.T,
	client *http.Client,
	baseURL string,
	artifactDir string,
	name string,
	code string,
) lookupResponse {
	t.Helper()

	payload, err := json.Marshal(map[string]string{"local_code": code})
	if err != nil {
		t.Fatalf("encode lookup request: %v", err)
	}
	body := post(t, client, baseURL+"/", "application/json", payload)
	writeArtifact(t, artifactDir, name, body)

	var response lookupResponse
	if err := json.Unmarshal(body, &response); err != nil {
		t.Fatalf("decode %s response: %v", name, err)
	}
	return response
}

func post(t *testing.T, client *http.Client, url string, contentType string, payload []byte) []byte {
	t.Helper()

	response, err := client.Post(url, contentType, bytes.NewReader(payload))
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		_ = response.Body.Close()
		t.Fatalf("read POST %s response: %v", url, err)
	}
	if err := response.Body.Close(); err != nil {
		t.Fatalf("close POST %s response: %v", url, err)
	}
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		t.Fatalf("POST %s status = %d, body = %q", url, response.StatusCode, body)
	}
	return body
}

func assertLookup(
	t *testing.T,
	response lookupResponse,
	code string,
	wantOK bool,
	wantSource string,
	wantStage string,
) {
	t.Helper()

	if response.OK != wantOK {
		t.Errorf("lookup %s ok = %t, want %t", code, response.OK, wantOK)
	}
	if response.Source != wantSource {
		t.Errorf("lookup %s source = %q, want %q", code, response.Source, wantSource)
	}
	if response.Stage != wantStage {
		t.Errorf("lookup %s stage = %q, want %q", code, response.Stage, wantStage)
	}
	if wantOK && response.Airport.LocalCode != code {
		t.Errorf("lookup %s airport code = %q", code, response.Airport.LocalCode)
	}
}

func writeArtifact(t *testing.T, directory string, name string, body []byte) {
	t.Helper()

	path := filepath.Join(directory, name+".json")
	if err := os.WriteFile(path, body, 0o644); err != nil {
		t.Fatalf("write %s artifact: %v", name, err)
	}
}

func envOrDefault(name string, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}

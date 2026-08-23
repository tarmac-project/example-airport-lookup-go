package main

import (
	"errors"
	"strings"
	"testing"

	airportpkg "github.com/tarmac-project/example-airport-lookup-go/pkg/airport"
	"github.com/tarmac-project/sdk/sql"
)

type fakeFunctionClient struct {
	name  string
	input []byte
	rsp   []byte
	err   error
}

func (f *fakeFunctionClient) Call(name string, input []byte) ([]byte, error) {
	f.name = name
	f.input = append([]byte(nil), input...)
	return f.rsp, f.err
}

type fakeSQLClient struct {
	calls  int
	failAt int
}

func (f *fakeSQLClient) Exec(string) (sql.ExecResult, error) {
	f.calls++
	if f.calls == f.failAt {
		return sql.ExecResult{}, errors.New("upsert failed")
	}
	return sql.ExecResult{RowsAffected: 1}, nil
}

func (*fakeSQLClient) Query(string) (sql.QueryResult, error) { return sql.QueryResult{}, nil }
func (*fakeSQLClient) Close() error                          { return nil }

type fakeLogger struct{}

func (*fakeLogger) Info(string)  {}
func (*fakeLogger) Warn(string)  {}
func (*fakeLogger) Error(string) {}
func (*fakeLogger) Debug(string) {}
func (*fakeLogger) Trace(string) {}

func TestEscapeSQL(t *testing.T) {
	input := "O'Hare \"line\"\nnext"
	got := escapeSQL(input)

	if !strings.Contains(got, "\\'") {
		t.Fatalf("expected escaped single quote, got %q", got)
	}
	if !strings.Contains(got, "\\\"") {
		t.Fatalf("expected escaped double quote, got %q", got)
	}
	if !strings.Contains(got, "\\n") {
		t.Fatalf("expected escaped newline, got %q", got)
	}
}

func TestBuildAirportUpsertQuery(t *testing.T) {
	ap := airportpkg.Airport{
		LocalCode:    "ORD",
		Name:         "Chicago O'Hare International Airport",
		Type:         "large_airport",
		TypeEmoji:    "✈️",
		Continent:    "NA",
		ISOCountry:   "US",
		ISORegion:    "US-IL",
		Municipality: "Chicago",
		Emoji:        "🇺🇸",
		Status:       "open",
	}

	query := buildAirportUpsertQuery(ap)

	if !strings.Contains(query, "INSERT INTO airports") {
		t.Fatalf("expected insert statement, got %q", query)
	}
	if !strings.Contains(query, "ON DUPLICATE KEY UPDATE") {
		t.Fatalf("expected upsert statement, got %q", query)
	}
	if !strings.Contains(query, "Chicago O\\'Hare International Airport") {
		t.Fatalf("expected escaped airport name, got %q", query)
	}
}

func TestMarshalLoadSummary(t *testing.T) {
	payload := marshalLoadSummary(loadSummary{
		FetchedBytes:   100,
		ParsedAirports: 20,
		SuccessUpsert:  19,
		FailedUpsert:   1,
	})

	got := string(payload)
	for _, expected := range []string{
		`"fetched_bytes":100`,
		`"parsed_airports":20`,
		`"successful_upsert":19`,
		`"failed_upsert":1`,
	} {
		if !strings.Contains(got, expected) {
			t.Fatalf("expected summary to contain %q, got %q", expected, got)
		}
	}
}

func TestHandlerForwardsFetchInput(t *testing.T) {
	csvData := []byte("1,CIX1,small_airport,CI Fixture Airport,0,0,0,NA,US,US-AZ,Phoenix,no,KCIX,,CIX1,,,")
	functionClient := &fakeFunctionClient{rsp: csvData}
	sqlClient := &fakeSQLClient{}
	function := &Function{
		logging:  &fakeLogger{},
		function: functionClient,
		sql:      sqlClient,
	}
	input := []byte("http://fixture-server/airports.csv")

	summary, err := function.Handler(input)
	if err != nil {
		t.Fatalf("Handler() unexpected error: %v", err)
	}
	if functionClient.name != "fetch" {
		t.Fatalf("called function %q, want fetch", functionClient.name)
	}
	if string(functionClient.input) != string(input) {
		t.Fatalf("forwarded input %q, want %q", functionClient.input, input)
	}
	if !strings.Contains(string(summary), "\"successful_upsert\":1") {
		t.Fatalf("summary = %s, want one successful upsert", summary)
	}
}

func TestHandlerFailures(t *testing.T) {
	validCSV := []byte("1,CIX1,small_airport,CI Fixture Airport,0,0,0,NA,US,US-AZ,Phoenix,no,KCIX,,CIX1,,,\n2,CIX2,small_airport,Second Fixture,0,0,0,NA,US,US-AZ,Tempe,no,KCIX2,,CIX2,,,")
	tests := []struct {
		name        string
		fetchRsp    []byte
		fetchErr    error
		failAt      int
		wantErrSub  string
		wantSummary string
	}{
		{name: "fetch error", fetchErr: errors.New("fetch failed"), wantErrSub: "failed to fetch airport data"},
		{name: "malformed CSV", fetchRsp: []byte("\"unterminated"), wantErrSub: "failed to parse airport data"},
		{name: "partial upsert failure", fetchRsp: validCSV, failAt: 2, wantErrSub: "1 failed upserts", wantSummary: "\"failed_upsert\":1"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			function := &Function{
				logging:  &fakeLogger{},
				function: &fakeFunctionClient{rsp: tc.fetchRsp, err: tc.fetchErr},
				sql:      &fakeSQLClient{failAt: tc.failAt},
			}

			summary, err := function.Handler(nil)
			if err == nil || !strings.Contains(err.Error(), tc.wantErrSub) {
				t.Fatalf("Handler() error = %v, want substring %q", err, tc.wantErrSub)
			}
			if tc.wantSummary != "" && !strings.Contains(string(summary), tc.wantSummary) {
				t.Fatalf("summary = %s, want substring %q", summary, tc.wantSummary)
			}
		})
	}
}
